package auth

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainnotification "github.com/tokenjoy/backend/internal/domain/notification"
	"github.com/tokenjoy/backend/internal/http/httputil"
	"github.com/tokenjoy/backend/internal/identity/httpx"
	"github.com/tokenjoy/backend/internal/identity/registertoken"
	"github.com/tokenjoy/backend/internal/identity/verifycode"
	"github.com/tokenjoy/backend/internal/store"
)

// --- SendCode ---

type sendCodeBody struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
}

// SendCode sends a verification code via SMS or Email.
func (h *Handler) SendCode(w http.ResponseWriter, r *http.Request) {
	if h.verifyCode == nil {
		httputil.WriteStatus(w, http.StatusServiceUnavailable, "verification service not configured")
		return
	}

	var body sendCodeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}

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

	result := h.verifyCode.Send(r.Context(), channel, address)
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

// --- VerifyCode ---

type verifyCodeBody struct {
	Phone string `json:"phone"`
	Email string `json:"email"`
	Code  string `json:"code"`
}

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

// VerifyCode verifies a code and performs SMS/Email login flow.
func (h *Handler) VerifyCode(w http.ResponseWriter, r *http.Request) {
	if h.verifyCode == nil {
		httputil.WriteStatus(w, http.StatusServiceUnavailable, "verification service not configured")
		return
	}

	var body verifyCodeBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.Code == "" {
		httputil.WriteStatus(w, http.StatusBadRequest, "code required")
		return
	}

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

	// Step 1: Verify code.
	vr := h.verifyCode.Verify(ctx, channel, address, body.Code)
	if !vr.OK {
		status := http.StatusBadRequest
		if vr.Locked {
			status = http.StatusTooManyRequests
		}
		httputil.WriteJSON(w, status, map[string]string{"message": vr.Reason}, nil)
		return
	}

	// Step 2: Find user by phone or email.
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
		httputil.WriteJSON(w, http.StatusOK, map[string]any{"action": "not_found"}, nil)
		return
	}

	// Step 3: Query members for this user.
	members, err := h.users.ListMemberCompanies(ctx, user.ID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}

	// Step 4: Route based on member count.
	switch {
	case len(members) == 1:
		m := members[0]
		h.issueTokenPairAndRespond(w, r, m.CompanyID, m.MemberID, user.ID,
			map[string]any{"action": "enter"}, false)

	case len(members) >= 2:
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

		// No members and no invites — user needs to create a company.
		h.issueRegisterSessionAndRespond(w, user.ID, map[string]any{
			"action": "create_company",
		})
	}
}

// --- SelectCompany ---

type selectCompanyBody struct {
	CompanyID uuid.UUID `json:"companyId"`
}

// SelectCompany allows a user with multiple companies to pick one after code verification.
func (h *Handler) SelectCompany(w http.ResponseWriter, r *http.Request) {
	var body selectCompanyBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteStatus(w, http.StatusBadRequest, httputil.MsgBadBody)
		return
	}
	if body.CompanyID == uuid.Nil {
		httputil.WriteStatus(w, http.StatusBadRequest, "companyId required")
		return
	}

	userID, err := h.registerToken.ResolveUserID(r)
	if err != nil {
		httputil.WriteStatus(w, http.StatusUnauthorized, "register session expired or invalid")
		return
	}

	ctx := r.Context()
	members, err := h.users.ListMemberCompanies(ctx, userID)
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

	h.issueTokenPairAndRespond(w, r, target.CompanyID, target.MemberID, userID,
		map[string]any{"memberId": target.MemberID.String(), "companyId": target.CompanyID.String()}, true)
}

// --- Register session helpers ---

func (h *Handler) issueTokenPairAndRespond(w http.ResponseWriter, r *http.Request, companyID, memberID, userID uuid.UUID, payload map[string]any, clearRegister bool) {
	_, err := httpx.IssueTokenPair(r.Context(), w, r, httpx.TokenPairParams{
		Secret:        h.pub.SessionToken.Secret(),
		SessionTTLSec: h.pub.Cfg.SessionTTLSec,
		RefreshTTLSec: h.pub.Cfg.RefreshTokenTTLSec,
		SecureCookie:  h.pub.SecureCookie,
		SessionStore:  h.sessions,
	}, companyID, memberID, userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	if clearRegister {
		registertoken.ClearCookie(w)
	}
	httputil.WriteJSON(w, http.StatusOK, payload, nil)
}

func (h *Handler) issueRegisterSessionAndRespond(w http.ResponseWriter, userID uuid.UUID, payload map[string]any) {
	regToken, err := h.registerToken.Issue(userID)
	if err != nil {
		httputil.WriteStatus(w, http.StatusInternalServerError, httputil.MsgInternal)
		return
	}
	registertoken.SetCookie(w, regToken, h.pub.SecureCookie)
	httputil.WriteJSON(w, http.StatusOK, payload, nil)
}
