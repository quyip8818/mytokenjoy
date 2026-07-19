package budget

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func GetPersonalBudget(members []types.Member, memberID uuid.UUID) int64 {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		return common.DefaultPersonalBudget
	}
	if member.PersonalBudget > 0 {
		return member.PersonalBudget
	}
	return common.DefaultPersonalBudget
}

func AddMemberPersonalBudget(members []types.Member, memberID uuid.UUID, amount int64) []types.Member {
	current := GetPersonalBudget(members, memberID)
	return SetMemberPersonalBudget(members, memberID, current+amount)
}

func SetMemberPersonalBudget(members []types.Member, memberID uuid.UUID, personalBudget int64) []types.Member {
	result := append([]types.Member{}, members...)
	for i := range result {
		if result[i].ID == memberID {
			result[i].PersonalBudget = personalBudget
			return result
		}
	}
	return append(result, types.Member{ID: memberID, PersonalBudget: personalBudget})
}

func GetAllocatedKeyBudget(platformKeys []types.PlatformKey, memberID uuid.UUID) int64 {
	return sumMemberScopeKeyBudget(platformKeys, memberID, uuid.Nil)
}

func sumMemberScopeKeyBudget(platformKeys []types.PlatformKey, memberID, excludeKeyID uuid.UUID) int64 {
	var sum int64
	for _, key := range platformKeys {
		if key.Scope != types.PlatformKeyScopeMember {
			continue
		}
		if key.MemberID == nil || *key.MemberID != memberID || key.Status != "active" {
			continue
		}
		if excludeKeyID != uuid.Nil && key.ID == excludeKeyID {
			continue
		}
		sum += key.Budget
	}
	return sum
}

func GetConsumedKeyBudget(platformKeys []types.PlatformKey, memberID uuid.UUID) int64 {
	var sum int64
	for _, key := range platformKeys {
		if key.Scope != types.PlatformKeyScopeMember {
			continue
		}
		if key.MemberID != nil && *key.MemberID == memberID && key.Status == "active" {
			sum += key.Consumed
		}
	}
	return sum
}

func GetBudgetRemaining(members []types.Member, platformKeys []types.PlatformKey, memberID uuid.UUID) int64 {
	return memberScopeBudgetRemaining(members, platformKeys, memberID, uuid.Nil)
}

func memberScopeBudgetRemaining(members []types.Member, platformKeys []types.PlatformKey, memberID, excludeKeyID uuid.UUID) int64 {
	remaining := GetPersonalBudget(members, memberID) - sumMemberScopeKeyBudget(platformKeys, memberID, excludeKeyID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func BuildBudgetSummary(members []types.Member, platformKeys []types.PlatformKey, memberID uuid.UUID, reservedPool int64) types.MemberBudgetSummary {
	totalBudget := GetPersonalBudget(members, memberID)
	consumed := GetConsumedKeyBudget(platformKeys, memberID)
	remaining := GetBudgetRemaining(members, platformKeys, memberID)
	return types.MemberBudgetSummary{
		TotalBudget: totalBudget, Consumed: consumed, Remaining: remaining, ReservedPool: reservedPool,
	}
}

func GetMemberBudgetCapacity(deptNode types.BudgetNode) int64 {
	var reserved int64
	if deptNode.ReservedPool != nil {
		reserved = *deptNode.ReservedPool
	}
	childrenSum := SumChildrenBudget(deptNode)
	capacity := deptNode.Budget - reserved - childrenSum
	if capacity < 0 {
		return 0
	}
	return capacity
}

func BuildMemberBudget(member types.Member, platformKeys []types.PlatformKey) types.MemberBudget {
	return types.MemberBudget{
		MemberID: member.ID, MemberName: member.Name, DepartmentID: member.DepartmentID,
		PersonalBudget: member.PersonalBudget,
		Allocated:      GetAllocatedKeyBudget(platformKeys, member.ID),
		Consumed:       GetConsumedKeyBudget(platformKeys, member.ID),
	}
}

func ValidateMemberBudgetUpdate(
	tree []types.BudgetNode,
	members []types.Member,
	platformKeys []types.PlatformKey,
	memberID uuid.UUID,
	personalBudget int64,
) *string {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		msg := "Member not found"
		return &msg
	}

	allocated := GetAllocatedKeyBudget(platformKeys, memberID)
	if personalBudget < allocated {
		msg := fmt.Sprintf("个人额度不能低于已分配 Key 额度（%d quota）", allocated)
		return &msg
	}

	deptNode := FindBudgetNode(tree, member.DepartmentID)
	if deptNode == nil {
		msg := "Department budget node not found"
		return &msg
	}

	capacity := GetMemberBudgetCapacity(*deptNode)
	var otherSum int64
	for _, m := range members {
		if m.DepartmentID == member.DepartmentID && m.ID != memberID {
			otherSum += GetPersonalBudget(members, m.ID)
		}
	}
	if otherSum+personalBudget > capacity {
		remaining := capacity - otherSum
		if remaining < 0 {
			remaining = 0
		}
		msg := fmt.Sprintf("超出部门可分配成员额度，当前剩余约 %d quota", remaining)
		return &msg
	}
	return nil
}

func ApplyMemberBudgetUpdate(
	members []types.Member,
	platformKeys []types.PlatformKey,
	memberID uuid.UUID,
	personalBudget int64,
) (types.MemberBudget, []types.Member) {
	updatedMembers := SetMemberPersonalBudget(members, memberID, personalBudget)
	member, ok := org.FindMemberByID(updatedMembers, memberID)
	if !ok {
		return types.MemberBudget{
			MemberID: memberID, PersonalBudget: personalBudget,
			Allocated: GetAllocatedKeyBudget(platformKeys, memberID),
			Consumed:  GetConsumedKeyBudget(platformKeys, memberID),
		}, updatedMembers
	}
	return BuildMemberBudget(*member, platformKeys), updatedMembers
}
