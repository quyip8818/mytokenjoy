package register

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/identity/sms"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

const registerSessionCookie = "tokenjoy_register_session"

// Handler implements POST /auth/register/* endpoints (SaaS only).
type Handler struct {
	companySvc          domaincompany.Service
	users               store.UserRepository
	sessions            store.SessionRepository
	sms                 *sms.Service
	registerToken       *registertoken.Issuer
	sessionToken        sessiontoken.Issuer
	secureCookie        bool
	registrationEnabled bool
	sessionTTLSec       int
	refreshTTLSec       int
}

func NewHandler(
	companySvc domaincompany.Service,
	users store.UserRepository,
	sessions store.SessionRepository,
	smsSvc *sms.Service,
	registerToken *registertoken.Issuer,
	sessionToken sessiontoken.Issuer,
	secureCookie bool,
	registrationEnabled bool,
	sessionTTLSec int,
	refreshTTLSec int,
) *Handler {
	return &Handler{
		companySvc:          companySvc,
		users:               users,
		sessions:            sessions,
		sms:                 smsSvc,
		registerToken:       registerToken,
		sessionToken:        sessionToken,
		secureCookie:        secureCookie,
		registrationEnabled: registrationEnabled,
		sessionTTLSec:       sessionTTLSec,
		refreshTTLSec:       refreshTTLSec,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/register/init", h.requireRegistration(h.Init))
	r.Post("/auth/register/accept", h.requireRegistration(h.Accept))
	r.Post("/auth/register/company", h.requireRegistration(h.Company))
}

func (h *Handler) requireRegistration(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !h.registrationEnabled {
			httputil.WriteStatus(w, http.StatusNotFound, "Not found")
			return
		}
		next(w, r)
	}
}

// --- Init ---

type initBody struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

type initResponseChoose struct {
	Action  string                        `json:"action"`
	Invites []domaincompany.PendingInvite `json:"invites"`
}

type initResponseLogin struct {
	Action string `json:"action"`
}

func (h *Handler) Init(w http.ResponseWriter, r *http.Request) {
	var body initBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Phone == "" || body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "phone and code required")
		return
	}

	ctx := r.Context()
	phone := sms.FormatPhone(body.Phone)

	// Step 1: Verify SMS code (same service as login; reuse does not consume a second attempt).
	vr := h.sms.Verify(ctx, phone, body.Code)
	if !vr.OK {
		status := http.StatusBadRequest
		if vr.Locked {
			status = http.StatusTooManyRequests
		}
		httputil.WriteJSON(w, status, map[string]string{"message": vr.Reason}, nil)
		return
	}

	// Step 2: Find or create user by phone.
	user, err := h.users.GetByPhone(ctx, phone)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if user == nil {
		// Create new user.
		userID := uuid.Must(uuid.NewV7())
		now := time.Now().UTC()
		newUser := store.User{
			ID:        userID,
			Phone:     phone,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.users.Create(ctx, newUser); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		user = &newUser
	} else {
		// User exists — check if they already have a member.
		hasMember, err := h.users.HasAnyMember(ctx, user.ID)
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		if hasMember {
			httputil.WriteJSON(w, http.StatusOK, initResponseLogin{Action: "login"}, nil)
			return
		}
	}

	// Query pending invites for this user.
	invites, err := h.companySvc.PendingInvitesForUser(ctx, domaincompany.PendingInvitesForUserRequest{
		Email:  user.Email,
		Phone:  user.Phone,
		UserID: user.ID,
	})
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if invites == nil {
		invites = []domaincompany.PendingInvite{}
	}

	// Issue register session token.
	regToken, err := h.registerToken.Issue(user.ID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	setRegisterSessionCookie(w, regToken, h.secureCookie)

	httputil.WriteJSON(w, http.StatusOK, initResponseChoose{
		Action:  "choose",
		Invites: invites,
	}, nil)
}

// --- Accept ---

type acceptBody struct {
	InviteCode string `json:"inviteCode"`
	Name       string `json:"name"`
}

type acceptResponse struct {
	MemberID  uuid.UUID `json:"memberId"`
	CompanyID uuid.UUID `json:"companyId"`
}

func (h *Handler) Accept(w http.ResponseWriter, r *http.Request) {
	var body acceptBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.InviteCode == "" || body.Name == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "inviteCode and name required")
		return
	}

	userID, err := h.resolveRegisterUser(r)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, "register session expired or invalid")
		return
	}

	member, err := h.companySvc.AcceptInvite(r.Context(), domaincompany.AcceptInviteRequest{
		UserID:     userID,
		InviteCode: body.InviteCode,
		Name:       body.Name,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}

	h.issueSessionAndRespond(w, r, member)
}

// --- Company ---

type companyBody struct {
	CompanyName string `json:"companyName"`
	Password    string `json:"password"`
}

func (h *Handler) Company(w http.ResponseWriter, r *http.Request) {
	var body companyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.CompanyName == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "companyName required")
		return
	}
	if len(body.Password) < 8 {
		httputil.WriteStatus(w, http.StatusBadRequest, "password too short (min 8)")
		return
	}

	userID, err := h.resolveRegisterUser(r)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, "register session expired or invalid")
		return
	}

	// Hash and persist password for the user.
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if err := h.users.UpdatePassword(r.Context(), userID, string(passwordHash)); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	result, err := h.companySvc.CreateCompany(r.Context(), domaincompany.CreateCompanyRequest{
		UserID: userID,
		Name:   body.CompanyName,
		Type:   store.CompanyTypeTrial,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}
	if result.Member == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	h.issueSessionAndRespond(w, r, *result.Member)
}

func (h *Handler) resolveRegisterUser(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(registerSessionCookie)
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := h.registerToken.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

func (h *Handler) issueSessionAndRespond(w http.ResponseWriter, r *http.Request, member types.Member) {
	sid := sessiontoken.NewSessionID()
	refreshRaw := sid + "." + sessiontoken.RandomHex(32)

	ttl := time.Duration(h.sessionTTLSec) * time.Second
	accessToken, err := sessiontoken.IssueAccessToken(
		h.sessionToken.Secret(), ttl,
		member.CompanyID, member.ID, member.UserID, sid,
	)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	now := time.Now().UTC()
	refreshTTL := time.Duration(h.refreshTTLSec) * time.Second
	sess := store.Session{
		ID:        sid,
		UserID:    member.UserID,
		MemberID:  member.ID,
		CompanyID: member.CompanyID,
		TokenHash: sessiontoken.SHA256Hex(refreshRaw),
		UserAgent: r.UserAgent(),
		IP:        r.RemoteAddr,
		CreatedAt: now,
		ExpiresAt: now.Add(refreshTTL),
	}
	if err := h.sessions.Create(r.Context(), sess); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	httpx.SetSessionCookie(w, accessToken, h.secureCookie)
	httpx.SetRefreshCookie(w, refreshRaw, h.secureCookie, h.refreshTTLSec)
	clearRegisterSessionCookie(w)
	httputil.WriteJSON(w, http.StatusOK, acceptResponse{
		MemberID:  member.ID,
		CompanyID: member.CompanyID,
	}, nil)
}

func setRegisterSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     registerSessionCookie,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   600, // 10 min
	})
}

func clearRegisterSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     registerSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}
