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
	"github.com/tokenjoy/backend/internal/store"
)

const registerSessionCookie = "tokenjoy_register_session"

// Handler implements POST /auth/register/* endpoints (SaaS only).
type Handler struct {
	companySvc          domaincompany.Service
	store               store.Store
	registerToken       *registertoken.Issuer
	sessionToken        sessiontoken.Issuer
	secureCookie        bool
	registrationEnabled bool
}

func NewHandler(
	companySvc domaincompany.Service,
	st store.Store,
	registerToken *registertoken.Issuer,
	sessionToken sessiontoken.Issuer,
	secureCookie bool,
	registrationEnabled bool,
) *Handler {
	return &Handler{
		companySvc:          companySvc,
		store:               st,
		registerToken:       registerToken,
		sessionToken:        sessionToken,
		secureCookie:        secureCookie,
		registrationEnabled: registrationEnabled,
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
	Token string `json:"token"` // SMS verification token (validated upstream)
}

type initResponseChoose struct {
	Action  string                      `json:"action"`
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
	if body.Phone == "" || body.Token == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "phone and token required")
		return
	}

	// TODO: validate SMS token (body.Token) against SMS verification service.
	// For now we trust the token was validated by the SMS layer.

	ctx := r.Context()

	// Find or create user by phone.
	user, err := h.store.User().GetByPhone(ctx, body.Phone)
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
			Phone:     body.Phone,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.store.User().Create(ctx, newUser); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		user = &newUser
	} else {
		// User exists — check if they already have a member.
		hasMember, err := h.store.User().HasAnyMember(ctx, user.ID)
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

	h.issueSessionAndRespond(w, member)
}

// --- Company ---

type companyBody struct {
	CompanyName string `json:"companyName"`
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

	userID, err := h.resolveRegisterUser(r)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, "register session expired or invalid")
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

	h.issueSessionAndRespond(w, *result.Member)
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

func (h *Handler) issueSessionAndRespond(w http.ResponseWriter, member types.Member) {
	token, err := h.sessionToken.IssueWithUser(member.CompanyID, member.ID, member.UserID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httpx.SetSessionCookie(w, token, h.secureCookie)
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
