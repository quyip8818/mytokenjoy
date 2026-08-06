package auth

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	"github.com/tokenjoy/backend/internal/domain/identity/registertoken"
	"github.com/tokenjoy/backend/internal/domain/identity/verifycode"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/http/httpx"
	"github.com/tokenjoy/backend/internal/store"
	"github.com/tokenjoy/backend/internal/support/invitetoken"
	"github.com/tokenjoy/backend/internal/support/tenant"
	"golang.org/x/crypto/bcrypt"
)

type Handler struct {
	pub           httpdeps.Public // ponytail: embedded via Auth, kept as named field for h.pub.X access pattern
	companySvc    domaincompany.Service
	users         store.UserRepository
	sessions      store.SessionRepository
	invites       store.InviteRepository
	orgRepo       store.OrgRepository
	companies     store.CompanyRepository
	verifyCode    *verifycode.Service
	registerToken *registertoken.Issuer
	inviteToken   *invitetoken.Issuer
}

func NewHandler(d httpdeps.Auth) *Handler {
	return &Handler{
		pub:           d.Public,
		companySvc:    d.CompanySvc,
		users:         d.Users,
		sessions:      d.Sessions,
		invites:       d.Invites,
		orgRepo:       d.OrgRepo,
		companies:     d.Companies,
		verifyCode:    d.VerifyCode,
		registerToken: d.RegisterToken,
		inviteToken:   d.InviteToken,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/auth/login", h.Login)
	r.Post("/auth/logout", h.Logout)
	r.Post("/auth/refresh", h.Refresh)
	r.Get("/auth/invite-info", h.InviteInfo)
	r.Post("/auth/accept-invite", h.AcceptInvite)
	r.Post("/auth/set-password", h.SetPassword)
	r.Post("/auth/reset-password", h.ResetPassword)
	r.Post("/auth/select-company", h.SelectCompany)

	// Verification code endpoints — only register if service is available.
	if h.verifyCode != nil {
		r.Post("/auth/verify-code/send", h.SendCode)
		r.Post("/auth/verify-code/verify", h.VerifyCode)
	}
}

type loginBody struct {
	Email    string `json:"email"` // phone or email
	Password string `json:"password"`
}

// authenticateResponse is the unified response for identity verification endpoints.
// Frontend uses this to decide: auto-select (1 company), show picker (N), create company, or accept invite.
type authenticateResponse struct {
	Action    string                `json:"action"`              // "select_company" | "create_company" | "choose"
	Companies []companyOption       `json:"companies,omitempty"` // non-empty when action=select_company
	Invites   []pendingInviteOption `json:"invites,omitempty"`   // non-empty when action=choose
}

// Login verifies credentials and returns an authToken (register cookie) + company list.
// It does NOT issue a session — the frontend must call POST /auth/select-company to establish a session.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var body loginBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.Email == "" || body.Password == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "credentials required")
		return
	}

	ctx := r.Context()

	// Step 1: Resolve user by identifier (phone or email).
	user, err := h.resolveUserByIdentifier(ctx, body.Email)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if user == nil || user.PasswordHash == "" {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, domain.NewDomainError(401, "Invalid credentials"))
		return
	}

	// Step 2: Verify password.
	if verifyPassword(user.PasswordHash, body.Password) != nil {
		httputil.WriteJSON(w, http.StatusUnauthorized, nil, domain.NewDomainError(401, "Invalid credentials"))
		return
	}

	// Step 3: Issue authToken + return companies.
	h.respondWithAuthToken(w, r, user.ID)
}

// respondWithAuthToken issues a register token and returns the company list (or invites).
// This is the shared "step 1 complete" response for both Login and VerifyCode.
func (h *Handler) respondWithAuthToken(w http.ResponseWriter, r *http.Request, userID uuid.UUID) {
	ctx := r.Context()
	members, err := h.users.ListMemberCompanies(ctx, userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	regToken, err := h.registerToken.Issue(userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	registertoken.SetCookie(w, regToken, h.pub.SecureCookie)

	// Has companies → select_company
	if len(members) > 0 {
		companies := make([]companyOption, len(members))
		for i, m := range members {
			companies[i] = companyOption{
				CompanyID:   m.CompanyID,
				CompanyName: m.CompanyName,
				Role:        m.Role,
			}
		}
		httputil.WriteJSON(w, http.StatusOK, authenticateResponse{
			Action:    "select_company",
			Companies: companies,
		}, nil)
		return
	}

	// No companies → check pending invites
	user, _ := h.users.GetByID(ctx, userID)
	if user != nil {
		invites, err := h.companySvc.PendingInvitesForUser(ctx, domaincompany.PendingInvitesForUserRequest{
			Email:  user.Email,
			Phone:  user.Phone,
			UserID: userID,
		})
		if err == nil && len(invites) > 0 {
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
			httputil.WriteJSON(w, http.StatusOK, authenticateResponse{
				Action:  "choose",
				Invites: opts,
			}, nil)
			return
		}
	}

	// No companies, no invites → create_company
	httputil.WriteJSON(w, http.StatusOK, authenticateResponse{
		Action: "create_company",
	}, nil)
}

// resolveUserByIdentifier looks up user by phone (if identifier looks like a phone) or email.
func (h *Handler) resolveUserByIdentifier(ctx context.Context, identifier string) (*store.User, error) {
	if isPhoneNumber(identifier) {
		return h.users.GetByPhone(ctx, verifycode.FormatPhone(identifier))
	}
	return h.users.GetByEmail(ctx, identifier)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	if claims, ok := httpx.ResolveMemberClaims(r, h.pub.SessionToken); ok && claims.Sid != "" {
		_ = h.sessions.Revoke(r.Context(), claims.Sid)
	}
	httpx.ClearSessionCookie(w)
	httpx.ClearRefreshCookie(w)
	w.WriteHeader(http.StatusNoContent)
}

// InviteInfo returns pre-fill data for the invite acceptance page.
// GET /auth/invite-info?code=<encrypted_token>
func (h *Handler) InviteInfo(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "code required")
		return
	}
	if h.inviteToken == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, "invite secret not configured")
		return
	}
	payload, err := h.inviteToken.Decrypt(code)
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid or expired invite token")
		return
	}

	ctx := r.Context()
	invite, err := h.invites.GetInviteByCode(ctx, payload.Code)
	if err != nil || invite == nil {
		httputil.WriteStatus(w, http.StatusNotFound, "invite not found")
		return
	}
	if invite.AcceptedAt != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invite already accepted")
		return
	}

	// Lookup member alias using invite's company context.
	var alias string
	tenantCtx := tenant.With(ctx, tenant.Info{CompanyID: invite.CompanyID})
	if invite.MemberID != uuid.Nil {
		if member, err := h.orgRepo.MemberByID(tenantCtx, invite.MemberID); err == nil && member != nil {
			alias = member.Alias
		}
	}

	// Lookup company name.
	var companyName string
	if co, err := h.companies.GetByID(ctx, invite.CompanyID); err == nil && co != nil {
		companyName = co.Name
	}

	// Lookup inviter name.
	var inviterName string
	if invite.InvitedBy != uuid.Nil {
		if inviter, err := h.orgRepo.MemberByID(tenantCtx, invite.InvitedBy); err == nil && inviter != nil {
			inviterName = inviter.Alias
			if inviterName == "" {
				inviterName = inviter.Name
			}
		}
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{
		"alias":       alias,
		"companyName": companyName,
		"inviterName": inviterName,
	}, nil)
}

type acceptInviteBody struct {
	InviteCode string `json:"inviteCode"`
	Name       string `json:"name"`
	Password   string `json:"password,omitempty"` // required only for unauthenticated (email link)
}

// AcceptInvite handles both logged-in users (session → userID) and
// unauthenticated users (email invite link → password creates/updates User).
// The inviteCode is an encrypted token containing the raw invite code and delivery channel.
func (h *Handler) AcceptInvite(w http.ResponseWriter, r *http.Request) {
	var body acceptInviteBody
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
		return
	}
	if body.InviteCode == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "inviteCode required")
		return
	}

	ctx := r.Context()

	// Decrypt token to extract raw invite code and channel.
	if h.inviteToken == nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, "invite secret not configured")
		return
	}
	payload, err := h.inviteToken.Decrypt(body.InviteCode)
	if err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "invalid invite token")
		return
	}
	rawCode := payload.Code
	channel := payload.Channel

	var userID uuid.UUID

	// Accept-invite is always an unauthenticated flow — ignore any existing session.
	// The invited user creates their own identity (or reuses one found via invite metadata).
	if len(body.Password) < 8 {
		httputil.WriteStatus(w, http.StatusBadRequest, "password too short (min 8)")
		return
	}
	// Validate invite early — fail before mutating user if invite is bad.
	invite, err := h.invites.GetInviteByCode(ctx, rawCode)
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

	// Find or create user.
	// For member-invites (admin_link), invite may have no email/phone — use UserID if set.
	var user *store.User
	if invite.UserID != uuid.Nil {
		user, err = h.users.GetByID(ctx, invite.UserID)
	} else if invite.Email != "" {
		user, err = h.users.GetByEmail(ctx, invite.Email)
	} else if invite.Phone != "" {
		user, err = h.users.GetByPhone(ctx, invite.Phone)
	}
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	if user == nil {
		now := time.Now().UTC()
		newUser := store.User{
			ID:           uuid.Must(uuid.NewV7()),
			Name:         body.Name,
			Email:        invite.Email,
			Phone:        invite.Phone,
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

	// Build accepted_meta from request context.
	acceptedMeta := map[string]any{
		"ip": r.RemoteAddr,
		"ua": r.UserAgent(),
	}

	member, err := h.companySvc.AcceptInvite(ctx, domaincompany.AcceptInviteRequest{
		UserID:              userID,
		InviteCode:          rawCode,
		Name:                body.Name,
		RegistrationChannel: channel,
		AcceptedMeta:        acceptedMeta,
	})
	if err != nil {
		httputil.WriteJSON(w, http.StatusBadRequest, nil, err)
		return
	}
	if _, err := h.issueTokenPair(w, r, member.CompanyID, member.ID, member.UserID); err != nil {
		slog.Error("accept-invite: issue token pair", "error", err, "memberID", member.ID, "userID", member.UserID)
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, map[string]any{
		"memberId":  member.ID,
		"companyId": member.CompanyID,
	}, nil)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
	if err := httputil.DecodeJSON(r, &body); err != nil {
		httputil.WriteError(w, err)
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
