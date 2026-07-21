package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"github.com/tokenjoy/backend/internal/pkg/ctxcompany"
	"github.com/tokenjoy/backend/internal/store"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	pub           httpdeps.Public
	companySvc    domaincompany.Service
	users         store.UserRepository
	org           store.OrgRepository
	sessions      store.SessionRepository
	invites       store.InviteRepository
	verifyCode    *verifycode.Service
	registerToken *registertoken.Issuer
}

func NewHandler(pub httpdeps.Public, companySvc domaincompany.Service,
	users store.UserRepository, org store.OrgRepository, sessions store.SessionRepository,
	invites store.InviteRepository, vc *verifycode.Service, regToken *registertoken.Issuer) *Handler {
	return &Handler{
		pub:           pub,
		companySvc:    companySvc,
		users:         users,
		org:           org,
		sessions:      sessions,
		invites:       invites,
		verifyCode:    vc,
		registerToken: regToken,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/refresh", h.Refresh)
	r.Post("/auth/accept-invite", h.AcceptInvite)
	r.Post("/auth/set-password", h.SetPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Get("/auth/invites/pending", h.PendingInvites)

	// Verification code endpoints — only register if service is available.
	if h.verifyCode != nil {
		r.Post("/auth/sms/send", h.SendCode)
		r.Post("/auth/sms/verify", h.VerifyCode)
		r.Post("/auth/sms/select", h.SelectCompany)
		r.Post("/auth/verify-code/send", h.SendCode)
		r.Post("/auth/verify-code/verify", h.VerifyCode)
	}
}

type loginBody struct {
	Email     string    `json:"email"`
	Password  string    `json:"password"`
	CompanyID uuid.UUID `json:"companyId"`
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.Email == "" || body.Password == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "credentials required")
		return
	}

	ctx := r.Context()
	identifier := body.Email

	// Detect if identifier is a phone number (digits, optional +86 prefix).
	isPhone := isPhoneNumber(identifier)

	if isPhone {
		// Phone + password login: look up all companies for this phone, verify password.
		results, err := h.org.MemberByPhone(ctx, verifycode.FormatPhone(identifier))
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		// Filter to those with matching password.
		var matched []store.MemberCompanyAuth
		for _, r := range results {
			if r.PasswordHash != "" && verifyPassword(r.PasswordHash, body.Password) == nil {
				matched = append(matched, r)
			}
		}
		if len(matched) == 0 {
			httputil.WriteJSON(w, http.StatusUnauthorized, nil, domain.NewDomainError(401, "Invalid credentials"))
			return
		}
		if len(matched) == 1 {
			m := matched[0].Member
			if _, err := h.issueTokenPair(w, r, m.CompanyID, m.ID, m.UserID); err != nil {
				httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				return
			}
			httputil.WriteJSON(w, http.StatusOK, map[string]string{"memberId": m.ID.String()}, nil)
			return
		}
		// Multiple companies → return list for selection (same as SMS flow).
		companies := make([]map[string]any, len(matched))
		for i, mc := range matched {
			companies[i] = map[string]any{
				"companyId":   mc.Member.CompanyID,
				"companyName": mc.CompanyName,
				"memberId":    mc.Member.ID,
			}
		}
		httputil.WriteJSON(w, http.StatusOK, map[string]any{
			"action":    "select_company",
			"companies": companies,
		}, nil)
		return
	}

	// Email + password login: original flow (requires companyID in SaaS mode).
	var companyID uuid.UUID
	if h.pub.Cfg.SupportSaas {
		if body.CompanyID == uuid.Nil {
			httputil.WriteJSON(w, http.StatusBadRequest, nil, domain.BadRequest("company id required for email login"))
			return
		}
		companyCtx, err := h.companySvc.ResolveCompanyContext(ctx, body.CompanyID)
		if err != nil {
			httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
			return
		}
		companyID = companyCtx.CompanyID
	} else {
		companyCtx, ok := ctxcompany.From(ctx)
		if !ok {
			httputil.WriteStatus(w, http.StatusBadRequest, "Company not found")
			return
		}
		companyID = companyCtx.CompanyID
	}
	member, err := h.pub.Credentials.AuthenticateMember(ctx, companyID, body.Email, body.Password)
	if err != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, err)
		return
	}
	if _, err := h.issueTokenPair(w, r, member.CompanyID, member.ID, member.UserID); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]string{"memberId": member.ID.String()}, nil)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if claims, ok := httpx.ResolveMemberClaims(r, h.pub.SessionToken); ok && claims.Sid != "" {
		_ = h.sessions.Revoke(r.Context(), claims.Sid)
	}
	httpx.ClearSessionCookie(w)
	httpx.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

type acceptInviteBody struct {
	InviteCode string `json:"inviteCode"`
	Name       string `json:"name"`
	Password   string `json:"password,omitempty"` // required only for unauthenticated (email link)
}

// AcceptInvite handles both logged-in users (session → userID) and
// unauthenticated users (email invite link → password creates/updates User).
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var body acceptInviteBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "Bad request")
		return
	}
	if body.InviteCode == "" || body.Name == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "inviteCode and name required")
		return
	}

	ctx := r.Context()
	var userID uuid.UUID

	// Try to resolve from session (logged-in user).
	if claims, ok := httpx.ResolveMemberClaims(r, h.pub.SessionToken); ok && claims.UserID != uuid.Nil {
		userID = claims.UserID
	} else {
		// Unauthenticated path: need password + resolve user from invite email.
		if len(body.Password) < 8 {
			httputil.WriteStatus(w, http.StatusBadRequest, "password too short (min 8)")
			return
		}
		// Validate invite early — fail before mutating user if invite is bad.
		invite, err := h.invites.GetInviteByCode(ctx, body.InviteCode)
		if err != nil || invite == nil {
			httputil.WriteStatus(w, http.StatusNotFound, "invite not found")
			return
		}
		if invite.AcceptedAt != nil {
			httputil.WriteStatus(w, http.StatusBadRequest, "invite already accepted")
			return
		}
		if time.Now().After(invite.ExpiresAt) {
			httputil.WriteStatus(w, http.StatusBadRequest, "invite expired")
			return
		}

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		// Find or create user by email.
		user, err := h.users.GetByEmail(ctx, invite.Email)
		if err != nil {
			httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
			return
		}
		if user == nil {
			now := time.Now().UTC()
			newUser := store.User{
				ID:           uuid.Must(uuid.NewV7()),
				Email:        invite.Email,
				PasswordHash: string(passwordHash),
				Status:       "active",
				CreatedAt:    now,
				UpdatedAt:    now,
			}
			if err := h.users.Create(ctx, newUser); err != nil {
				httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				return
			}
			userID = newUser.ID
		} else {
			if err := h.users.UpdatePassword(ctx, user.ID, string(passwordHash)); err != nil {
				httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
				return
			}
			userID = user.ID
		}
	}

	member, err := h.companySvc.AcceptInvite(ctx, domaincompany.AcceptInviteRequest{
		UserID:     userID,
		InviteCode: body.InviteCode,
		Name:       body.Name,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}
	if _, err := h.issueTokenPair(w, r, member.CompanyID, member.ID, member.UserID); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"memberId":  member.ID,
		"companyId": member.CompanyID,
	}, nil)
}

// PendingInvites returns pending invites for the currently logged-in user.
func (h *Handler) PendingInvites(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.ResolveMemberClaims(r, h.pub.SessionToken)
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	ctx := r.Context()
	user, err := h.users.GetByID(ctx, claims.UserID)
	if err != nil || user == nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

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
	httputil.WriteJSON(w, http.StatusOK, invites, nil)
}

// --- SetPassword ---

type setPasswordBody struct {
	Password string `json:"password"`
}

// SetPassword allows a logged-in user (e.g. after SMS login) to set or update their password.
func (h *Handler) SetPassword(w http.ResponseWriter, r *http.Request) {
	claims, ok := httpx.ResolveMemberClaims(r, h.pub.SessionToken)
	if !ok || claims.UserID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, httputil.MsgUnauthorized)
		return
	}

	var body setPasswordBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if len(body.Password) < 8 {
		httputil.WriteStatus(w, http.StatusBadRequest, "password too short (min 8)")
		return
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(body.Password), bcrypt.DefaultCost)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if err := h.users.UpdatePassword(r.Context(), claims.UserID, string(passwordHash)); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Reset Password ---

type resetPasswordBody struct {
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	Code        string `json:"code"`
	NewPassword string `json:"newPassword"`
}

// ResetPassword verifies a code (sent via SMS or Email) then sets a new password.
func (h *Handler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	var body resetPasswordBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Code == "" || len(body.NewPassword) < 8 {
		httputil.WriteStatus(w, http.StatusBadRequest, "code and newPassword (min 8) required")
		return
	}

	// Determine channel and address.
	var channel, address string
	switch {
	case body.Phone != "":
		channel = domainnotification.ChannelSMS
		address = verifycode.FormatPhone(body.Phone)
	case body.Email != "":
		channel = domainnotification.ChannelEmail
		address = body.Email
	default:
		httputil.WriteStatus(w, http.StatusBadRequest, "phone or email required")
		return
	}

	ctx := r.Context()

	if h.verifyCode == nil {
		httputil.WriteStatus(w, http.StatusServiceUnavailable, "verification service not configured")
		return
	}
	vr := h.verifyCode.Verify(ctx, channel, address, body.Code)
	if !vr.OK {
		status := http.StatusBadRequest
		if vr.Locked {
			status = http.StatusTooManyRequests
		}
		httputil.WriteJSON(w, status, map[string]string{"message": vr.Reason}, nil)
		return
	}

	// Find user by phone or email.
	var user *store.User
	var err error
	if channel == domainnotification.ChannelSMS {
		user, err = h.users.GetByPhone(ctx, address)
	} else {
		user, err = h.users.GetByEmail(ctx, address)
	}
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if user == nil {
		httputil.WriteStatus(w, http.StatusNotFound, "user not found")
		return
	}

	// Hash and save new password.
	hash, err := bcrypt.GenerateFromPassword([]byte(body.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if err := h.users.UpdatePassword(ctx, user.ID, string(hash)); err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Helpers ---

func isPhoneNumber(s string) bool {
	cleaned := s
	if len(cleaned) > 0 && cleaned[0] == '+' {
		cleaned = cleaned[1:]
	}
	if len(cleaned) < 11 {
		return false
	}
	for _, c := range cleaned {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}

func verifyPassword(hash, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}
