package authsms

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/identity/sms"
	"github.com/tokenjoy/backend/internal/store"
)

// Handler implements POST /auth/sms/* endpoints (SaaS only).
type Handler struct {
	sms           *sms.Service
	companySvc    domaincompany.Service
	store         store.Store
	registerToken *registertoken.Issuer
	sessionToken  sessiontoken.Issuer
	secureCookie  bool
	sessionTTLSec int
	refreshTTLSec int
}

func NewHandler(
	smsSvc *sms.Service,
	companySvc domaincompany.Service,
	st store.Store,
	registerToken *registertoken.Issuer,
	sessionToken sessiontoken.Issuer,
	secureCookie bool,
	sessionTTLSec int,
	refreshTTLSec int,
) *Handler {
	if smsSvc == nil {
		return nil
	}
	return &Handler{
		sms:           smsSvc,
		companySvc:    companySvc,
		store:         st,
		registerToken: registerToken,
		sessionToken:  sessionToken,
		secureCookie:  secureCookie,
		sessionTTLSec: sessionTTLSec,
		refreshTTLSec: refreshTTLSec,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/sms/send", h.Send)
	r.Post("/auth/sms/verify", h.Verify)
	r.Post("/auth/sms/select", h.Select)
}

// --- Send ---

type sendBody struct {
	Phone string `json:"phone"`
	// TODO: captcha token validation (上线阻塞项)
}

func (h *Handler) Send(w http.ResponseWriter, r *http.Request) {
	var body sendBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Phone == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "phone required")
		return
	}

	phone := sms.FormatPhone(body.Phone)
	result := h.sms.Send(r.Context(), phone)
	if !result.OK {
		status := http.StatusTooManyRequests
		resp := map[string]any{"message": result.Reason}
		if result.RetryAfter > 0 {
			resp["retryAfter"] = result.RetryAfter
		}
		httputil.WriteJSON(w, status, resp, nil)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Verify ---

type verifyBody struct {
	Phone string `json:"phone"`
	Code  string `json:"code"`
}

// SmsVerifyResult matches the design doc's discriminated union response.
type companyOption struct {
	CompanyID   uuid.UUID `json:"companyId"`
	CompanyName string    `json:"companyName"`
	Role        string    `json:"role"`
}

type pendingInviteOption struct {
	InviteCode  string    `json:"inviteCode"`
	CompanyID   uuid.UUID `json:"companyId"`
	CompanyName string    `json:"companyName"`
	Role        string    `json:"role"`
	ExpiresAt   string    `json:"expiresAt"`
}

func (h *Handler) Verify(w http.ResponseWriter, r *http.Request) {
	var body verifyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Phone == "" || body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "phone and code required")
		return
	}

	phone := sms.FormatPhone(body.Phone)
	ctx := r.Context()

	// Step 1: Verify SMS code.
	vr := h.sms.Verify(ctx, phone, body.Code)
	if !vr.OK {
		status := http.StatusBadRequest
		if vr.Locked {
			status = http.StatusTooManyRequests
		}
		httputil.WriteJSON(w, status, map[string]string{"message": vr.Reason}, nil)
		return
	}

	// Step 2: Find or create user.
	user, err := h.store.User().GetByPhone(ctx, phone)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if user == nil {
		// Create new user.
		now := time.Now().UTC()
		newUser := store.User{
			ID:        uuid.Must(uuid.NewV7()),
			Phone:     phone,
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := h.store.User().Create(ctx, newUser); err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		user = &newUser
	}

	// Step 3: Query members for this user.
	members, err := h.store.User().ListMemberCompanies(ctx, user.ID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Step 4: Route based on member count.
	switch {
	case len(members) == 1:
		// Single member → direct enter (§21: don't expose memberId).
		m := members[0]
		h.issueTokenPairAndRespond(w, r, m.CompanyID, m.MemberID, user.ID,
			map[string]any{"action": "enter"}, false)
		return

	case len(members) >= 2:
		// Multiple members → select_company.
		companies := make([]companyOption, len(members))
		for i, m := range members {
			companies[i] = companyOption{
				CompanyID:   m.CompanyID,
				CompanyName: m.CompanyName,
				Role:        m.Role,
			}
		}
		h.issueRegisterSessionAndRespond(w, user.ID, map[string]any{
			"action":    "select_company",
			"companies": companies,
		})
		return

	default:
		// 0 members → check invites.
		invites, err := h.companySvc.PendingInvitesForUser(ctx, domaincompany.PendingInvitesForUserRequest{
			Email:  user.Email,
			Phone:  user.Phone,
			UserID: user.ID,
		})
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}

		if len(invites) > 0 {
			opts := make([]pendingInviteOption, len(invites))
			for i, inv := range invites {
				opts[i] = pendingInviteOption{
					InviteCode:  inv.InviteCode,
					CompanyID:   inv.CompanyID,
					CompanyName: inv.CompanyName,
					Role:        inv.Role,
					ExpiresAt:   inv.ExpiresAt.Format(time.RFC3339),
				}
			}
			h.issueRegisterSessionAndRespond(w, user.ID, map[string]any{
				"action":  "choose",
				"invites": opts,
			})
			return
		}

		// No members, no invites → onboard.
		h.issueRegisterSessionAndRespond(w, user.ID, map[string]any{
			"action": "onboard",
		})
	}
}

// --- Select ---

type selectBody struct {
	CompanyID uuid.UUID `json:"companyId"`
}

func (h *Handler) Select(w http.ResponseWriter, r *http.Request) {
	var body selectBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.CompanyID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "companyId required")
		return
	}

	// Resolve user from register session.
	userID, err := h.resolveRegisterUser(r)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, "register session expired or invalid")
		return
	}

	ctx := r.Context()

	// Verify the user actually has a member in this company.
	members, err := h.store.User().ListMemberCompanies(ctx, userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	var target *store.MemberCompany
	for i := range members {
		if members[i].CompanyID == body.CompanyID {
			target = &members[i]
			break
		}
	}
	if target == nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "not a member of this company")
		return
	}

	// Issue token pair.
	h.issueTokenPairAndRespond(w, r, target.CompanyID, target.MemberID, userID,
		map[string]any{"memberId": target.MemberID.String(), "companyId": target.CompanyID.String()}, true)
}

// --- Helpers ---

const registerCookieName = "tokenjoy_register_session"

func (h *Handler) resolveRegisterUser(r *http.Request) (uuid.UUID, error) {
	cookie, err := r.Cookie(registerCookieName)
	if err != nil {
		return uuid.Nil, err
	}
	claims, err := h.registerToken.Parse(cookie.Value)
	if err != nil {
		return uuid.Nil, err
	}
	return claims.UserID, nil
}

// issueTokenPairAndRespond issues session cookies and responds with a JSON payload.
// If clearRegister is true, the register session cookie is removed.
func (h *Handler) issueTokenPairAndRespond(w http.ResponseWriter, r *http.Request, companyID, memberID, userID uuid.UUID, payload map[string]any, clearRegister bool) {
	_, err := httpx.IssueTokenPair(r.Context(), w, r, httpx.TokenPairParams{
		Secret:        h.sessionToken.Secret(),
		SessionTTLSec: h.sessionTTLSec,
		RefreshTTLSec: h.refreshTTLSec,
		SecureCookie:  h.secureCookie,
		SessionStore:  h.store.Session(),
	}, companyID, memberID, userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if clearRegister {
		clearRegisterSessionCookie(w)
	}
	httputil.WriteJSON(w, http.StatusOK, payload, nil)
}

func (h *Handler) issueRegisterSessionAndRespond(w http.ResponseWriter, userID uuid.UUID, payload map[string]any) {
	regToken, err := h.registerToken.Issue(userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	setRegisterSessionCookie(w, regToken, h.secureCookie)
	httputil.WriteJSON(w, http.StatusOK, payload, nil)
}

func setRegisterSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     registerCookieName,
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
		Name:     registerCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
}
