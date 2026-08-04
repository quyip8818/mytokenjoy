package structure

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/grants"
	"github.com/tokenjoy/backend/internal/domain/org/core"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	pkgorg "github.com/tokenjoy/backend/internal/pkg/org"
	"github.com/tokenjoy/backend/internal/store"
)

func (s *LocalService) BatchInvite(ctx context.Context, ids []uuid.UUID, callerMemberID uuid.UUID) (types.BatchInviteResult, error) {
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.BatchInviteResult{}, err
	}

	// Filter to pending members only.
	targets := make([]types.Member, 0)
	if len(ids) > 0 {
		idSet := make(map[uuid.UUID]struct{}, len(ids))
		for _, id := range ids {
			idSet[id] = struct{}{}
		}
		for _, member := range members {
			if _, ok := idSet[member.ID]; ok && member.Status == types.MemberStatusPending {
				targets = append(targets, member)
			}
		}
	} else {
		for _, member := range members {
			if member.Status == types.MemberStatusPending {
				targets = append(targets, member)
			}
		}
	}

	sent := 0
	for _, target := range targets {
		// Find or create invite for this member.
		invite, err := s.d.Store.Invite().GetInviteByMemberID(ctx, target.ID)
		if err != nil || invite == nil {
			// No invite exists — create one.
			code, cerr := randomHexCode()
			if cerr != nil {
				continue
			}
			now := time.Now().UTC()
			newInvite := store.CompanyInvite{
				ID:         uuid.Must(uuid.NewV7()),
				CompanyID:  store.CompanyID(ctx),
				MemberID:   target.ID,
				UserID:     target.UserID,
				Role:       "member",
				InviteCode: code,
				InvitedBy:  callerMemberID,
				ExpiresAt:  now.Add(s.d.Cfg.InviteExpireDuration()),
				CreatedAt:  now,
			}
			if err := s.d.Store.Invite().CreateInvite(ctx, newInvite); err != nil {
				s.d.Logger.Warn("batch invite: create invite failed", "memberID", target.ID, "error", err)
				continue
			}
			invite = &newInvite
		} else if time.Now().After(invite.ExpiresAt) {
			// Renew expiry if expired.
			newExpiry := time.Now().UTC().Add(s.d.Cfg.InviteExpireDuration())
			if err := s.d.Store.Invite().UpdateInviteExpiry(ctx, invite.ID, newExpiry); err != nil {
				s.d.Logger.Warn("batch invite: renew expiry failed", "memberID", target.ID, "error", err)
				continue
			}
			invite.ExpiresAt = newExpiry
		}

		// Lookup user to get phone/email for sending.
		user, err := s.d.Store.User().GetByID(ctx, target.UserID)
		if err != nil || user == nil {
			continue
		}

		s.sendInviteNotifications(ctx, invite.InviteCode, user.Phone, user.Email)
		sent++
	}

	return types.BatchInviteResult{Sent: sent}, nil
}

func (s *LocalService) BatchImport(ctx context.Context, rows []types.BatchImportRow, callerMemberID uuid.UUID) (types.MemberBatchImportResult, error) {
	members, err := s.d.Store.Org().Members(ctx)
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	if err := s.checkTrialMemberLimitBatch(ctx, members, len(rows)); err != nil {
		return types.MemberBatchImportResult{}, err
	}
	// Load OrgNode tree (full data with budget etc.) for department resolution.
	orgNodes, err := s.d.Store.Org().Nodes().Tree(ctx)
	if err != nil {
		return types.MemberBatchImportResult{}, err
	}
	departments := types.OrgNodesToDepartments(orgNodes)
	flat := pkgorg.FlattenDepartmentTree(departments)
	// deptTreeModified tracks whether we need to persist the tree after the loop.
	deptTreeModified := false
	failures := make([]types.MemberBatchImportFailure, 0)
	imported := 0

	type importedMember struct {
		inviteCode string
		phone      string
		email      string
	}
	var invitesToSend []importedMember

	companyID := store.CompanyID(ctx)
	now := time.Now().UTC()
	expiresAt := now.Add(s.d.Cfg.InviteExpireDuration())

	for index, row := range rows {
		if row.Phone == "" && row.Email == "" {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "手机号或邮箱至少填写一项",
			})
			continue
		}
		dept, created := resolveDepartmentByPath(row.DepartmentName, orgNodes, flat)
		if created {
			// Re-flatten since tree was modified.
			departments = types.OrgNodesToDepartments(orgNodes)
			flat = pkgorg.FlattenDepartmentTree(departments)
			deptTreeModified = true
		}
		if dept == nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "部门「" + row.DepartmentName + "」不存在",
			})
			continue
		}
		userID, uerr := s.resolveOrCreateUser(ctx, row.Phone, row.Email, row.Name)
		if uerr != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: uerr.Error(),
			})
			continue
		}

		// Check if this user is already a member (existing or just added in this batch).
		duplicate := false
		for _, m := range members {
			if m.UserID == userID {
				duplicate = true
				break
			}
		}
		if duplicate {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "该用户已存在（手机号或邮箱与已有成员重复）",
			})
			continue
		}

		memberID := generateID()
		members = append(members, types.Member{
			ID: memberID, UserID: userID,
			Alias:        row.Name,
			EmployeeID:   row.EmployeeId,
			JobTitle:     row.JobTitle,
			HireDate:     row.HireDate,
			DepartmentID: dept.ID, DepartmentName: dept.Name,
			Status: types.MemberStatusPending, Roles: []string{grants.RoleMember}, Source: types.MemberSourceCSV,
			PersonalBudget: common.DefaultPersonalBudget,
		})

		// Create invite for this member.
		code, cerr := randomHexCode()
		if cerr != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "创建邀请码失败",
			})
			continue
		}
		invite := store.CompanyInvite{
			ID:         uuid.Must(uuid.NewV7()),
			CompanyID:  companyID,
			MemberID:   memberID,
			Email:      row.Email,
			Phone:      row.Phone,
			UserID:     userID,
			Role:       "member",
			InviteCode: code,
			InvitedBy:  callerMemberID,
			ExpiresAt:  expiresAt,
			CreatedAt:  now,
		}
		if err := s.d.Store.Invite().CreateInvite(ctx, invite); err != nil {
			failures = append(failures, types.MemberBatchImportFailure{
				Row: index + 1, Reason: "创建邀请失败",
			})
			continue
		}

		invitesToSend = append(invitesToSend, importedMember{
			inviteCode: code, phone: row.Phone, email: row.Email,
		})
		imported++
	}

	if deptTreeModified {
		if err := s.d.Store.Org().Nodes().SetTree(ctx, orgNodes); err != nil {
			return types.MemberBatchImportResult{}, err
		}
	}

	if err := s.d.Store.Org().SetMembers(ctx, members); err != nil {
		return types.MemberBatchImportResult{Imported: imported, Failures: append(failures, types.MemberBatchImportFailure{
			Row: 0, Reason: "保存失败，请重试",
		})}, nil
	}
	if imported > 0 {
		if err := persistRecalculatedMemberCounts(ctx, s.d.Store, members); err != nil {
			return types.MemberBatchImportResult{}, err
		}
		if err := core.BumpAuthzRevision(ctx, s.d); err != nil {
			return types.MemberBatchImportResult{}, err
		}
	}

	// Best-effort: send invite notifications.
	for _, im := range invitesToSend {
		s.sendInviteNotifications(ctx, im.inviteCode, im.phone, im.email)
	}

	return types.MemberBatchImportResult{Imported: imported, Failures: failures}, nil
}

// resolveDepartmentByPath resolves a department by name or path (e.g. "技术部/前端组").
// If the department (or intermediate nodes) don't exist, they are auto-created in the OrgNode tree.
// Returns the resolved department (from flat) and whether the tree was modified.
func resolveDepartmentByPath(name string, tree []types.OrgNode, flat []types.Department) (*types.Department, bool) {
	// Simple name (no slash) — try exact match in flat list first.
	if !strings.Contains(name, "/") {
		for i := range flat {
			if flat[i].Name == name {
				return &flat[i], false
			}
		}
		// Not found — auto-create under root.
		if len(tree) == 0 {
			return nil, false
		}
		newNode := types.OrgNode{
			ID:       generateID(),
			Name:     name,
			ParentID: &tree[0].ID,
		}
		tree[0].Children = append(tree[0].Children, newNode)
		dept := types.OrgNodeToDepartment(newNode)
		return &dept, true
	}

	// Path notation: split and walk/create.
	segments := strings.Split(name, "/")
	modified := false
	current := tree // children at current level
	var parent *types.OrgNode

	// Start from root node.
	if len(tree) > 0 {
		if tree[0].Name == segments[0] {
			parent = &tree[0]
			current = tree[0].Children
			segments = segments[1:]
		} else {
			parent = &tree[0]
			current = tree[0].Children
		}
	}

	for _, seg := range segments {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			continue
		}
		found := false
		for i := range current {
			if current[i].Name == seg {
				parent = &current[i]
				current = current[i].Children
				found = true
				break
			}
		}
		if !found {
			if parent == nil {
				return nil, modified
			}
			newNode := types.OrgNode{
				ID:       generateID(),
				Name:     seg,
				ParentID: &parent.ID,
			}
			parent.Children = append(parent.Children, newNode)
			parent = &parent.Children[len(parent.Children)-1]
			current = parent.Children
			modified = true
		}
	}

	if parent == nil {
		return nil, modified
	}
	dept := types.OrgNodeToDepartment(*parent)
	return &dept, modified
}
