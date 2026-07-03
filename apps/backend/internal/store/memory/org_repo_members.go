package memory

import (
	"context"
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/store"
)

func (r *memoryOrgRepo) Members(ctx context.Context) ([]types.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	return store.CloneMembers(r.store.companySnapshot(store.CompanyID(ctx)).Members), nil
}

func (r *memoryOrgRepo) MemberByID(ctx context.Context, memberID string) (*types.Member, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, member := range r.store.companySnapshot(store.CompanyID(ctx)).Members {
		if member.ID == memberID {
			cloned := store.CloneMember(member)
			return &cloned, nil
		}
	}
	return nil, nil
}

func (r *memoryOrgRepo) MemberByEmail(ctx context.Context, companyID int64, email string) (*types.Member, string, error) {
	if err := ctx.Err(); err != nil {
		return nil, "", err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, member := range r.store.companySnapshot(companyID).Members {
		if member.Email == email {
			cloned := store.CloneMember(member)
			hash := r.store.memberPasswordHashes[memberPasswordKey(companyID, member.ID)]
			return &cloned, hash, nil
		}
	}
	return nil, "", nil
}

func (r *memoryOrgRepo) GetMemberAuthz(ctx context.Context, companyID int64, memberID string) (*store.MemberAuthz, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	var member *types.Member
	for _, item := range r.store.companySnapshot(companyID).Members {
		if item.ID == memberID {
			cloned := store.CloneMember(item)
			member = &cloned
			break
		}
	}
	if member == nil {
		return nil, nil
	}
	company, ok := r.store.companies[companyID]
	revision := int64(0)
	if ok {
		revision = company.AuthzRevision
	}
	return &store.MemberAuthz{
		Member:        *member,
		Roles:         store.CloneRoles(r.store.companySnapshot(companyID).Roles),
		AuthzRevision: revision,
	}, nil
}

func (r *memoryOrgRepo) MemberPersonalQuota(ctx context.Context, memberID string) (float64, bool, error) {
	if err := ctx.Err(); err != nil {
		return 0, false, err
	}
	r.store.mu.RLock()
	defer r.store.mu.RUnlock()
	for _, member := range r.store.companySnapshot(store.CompanyID(ctx)).Members {
		if member.ID == memberID {
			return member.PersonalQuota, true, nil
		}
	}
	return 0, false, nil
}

func (r *memoryOrgRepo) SetMembers(ctx context.Context, members []types.Member) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	snap.Members = store.CloneMembers(members)
	r.store.setCompanySnapshot(tid, snap)
	return nil
}

func (r *memoryOrgRepo) UpdateMemberPersonalQuota(ctx context.Context, memberID string, personalQuota float64) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	tid := store.CompanyID(ctx)
	snap := r.store.companySnapshot(tid)
	for i := range snap.Members {
		if snap.Members[i].ID == memberID {
			snap.Members[i].PersonalQuota = personalQuota
			r.store.setCompanySnapshot(tid, snap)
			return nil
		}
	}
	return nil
}

func memberPasswordKey(companyID int64, memberID string) string {
	return fmt.Sprintf("%d:%s", companyID, memberID)
}

func (r *memoryOrgRepo) SetMemberPasswordHash(ctx context.Context, memberID, passwordHash string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	r.store.mu.Lock()
	defer r.store.mu.Unlock()
	if r.store.memberPasswordHashes == nil {
		r.store.memberPasswordHashes = make(map[string]string)
	}
	r.store.memberPasswordHashes[memberPasswordKey(store.CompanyID(ctx), memberID)] = passwordHash
	return nil
}
