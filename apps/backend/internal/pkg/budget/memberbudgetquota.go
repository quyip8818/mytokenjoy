package budget

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

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

func BuildMemberBudgetQuota(member types.Member, platformKeys []types.PlatformKey) types.MemberBudgetQuota {
	return types.MemberBudgetQuota{
		MemberID: member.ID, MemberName: member.Name, DepartmentID: member.DepartmentID,
		PersonalBudget: member.PersonalBudget,
		Allocated:      GetAllocatedKeyBudget(platformKeys, member.ID),
		Used:           GetUsedKeyBudget(platformKeys, member.ID),
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
) (types.MemberBudgetQuota, []types.Member) {
	updatedMembers := SetMemberPersonalBudget(members, memberID, personalBudget)
	member, ok := org.FindMemberByID(updatedMembers, memberID)
	if !ok {
		return types.MemberBudgetQuota{
			MemberID: memberID, PersonalBudget: personalBudget,
			Allocated: GetAllocatedKeyBudget(platformKeys, memberID),
			Used:      GetUsedKeyBudget(platformKeys, memberID),
		}, updatedMembers
	}
	return BuildMemberBudgetQuota(*member, platformKeys), updatedMembers
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
