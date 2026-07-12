package budget

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/common"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func GetPersonalBudget(members []types.Member, memberID string) float64 {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		return common.DefaultPersonalBudget
	}
	if member.PersonalBudget > 0 {
		return member.PersonalBudget
	}
	return common.DefaultPersonalBudget
}

func AddMemberPersonalBudget(members []types.Member, memberID string, amount float64) []types.Member {
	current := GetPersonalBudget(members, memberID)
	return SetMemberPersonalBudget(members, memberID, current+amount)
}

func SetMemberPersonalBudget(members []types.Member, memberID string, personalBudget float64) []types.Member {
	result := append([]types.Member{}, members...)
	for i := range result {
		if result[i].ID == memberID {
			result[i].PersonalBudget = personalBudget
			return result
		}
	}
	return append(result, types.Member{ID: memberID, PersonalBudget: personalBudget})
}

func GetAllocatedKeyBudget(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.ProjectID == nil && key.Status == "active" {
			sum += key.Budget
		}
	}
	return sum
}

func GetConsumedKeyBudget(platformKeys []types.PlatformKey, memberID string) float64 {
	sum := 0.0
	for _, key := range platformKeys {
		if key.MemberID != nil && *key.MemberID == memberID && key.ProjectID == nil && key.Status == "active" {
			sum += key.Consumed
		}
	}
	return sum
}

func GetBudgetRemaining(members []types.Member, platformKeys []types.PlatformKey, memberID string) float64 {
	remaining := GetPersonalBudget(members, memberID) - GetAllocatedKeyBudget(platformKeys, memberID)
	if remaining < 0 {
		return 0
	}
	return remaining
}

func BuildBudgetSummary(members []types.Member, platformKeys []types.PlatformKey, memberID string, reservedPool float64) types.MemberBudgetSummary {
	totalBudget := GetPersonalBudget(members, memberID)
	consumed := GetConsumedKeyBudget(platformKeys, memberID)
	remaining := GetBudgetRemaining(members, platformKeys, memberID)
	return types.MemberBudgetSummary{
		TotalBudget: totalBudget, Consumed: consumed, Remaining: remaining, ReservedPool: reservedPool,
	}
}

func GetMemberBudgetCapacity(deptNode types.BudgetNode) float64 {
	reserved := 0.0
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
	memberID string,
	personalBudget float64,
) *string {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		msg := "Member not found"
		return &msg
	}

	allocated := GetAllocatedKeyBudget(platformKeys, memberID)
	if personalBudget < allocated {
		msg := fmt.Sprintf("个人额度不能低于已分配 Key 额度（¥%s）", formatMoney(allocated))
		return &msg
	}

	deptNode := FindBudgetNode(tree, member.DepartmentID)
	if deptNode == nil {
		msg := "Department budget node not found"
		return &msg
	}

	capacity := GetMemberBudgetCapacity(*deptNode)
	otherSum := 0.0
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
		msg := fmt.Sprintf("超出部门可分配成员额度，当前剩余约 ¥%s", formatMoney(remaining))
		return &msg
	}
	return nil
}

func ApplyMemberBudgetUpdate(
	members []types.Member,
	platformKeys []types.PlatformKey,
	memberID string,
	personalBudget float64,
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

func formatMoney(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
