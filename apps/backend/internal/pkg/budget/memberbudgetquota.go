package budget

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/org"
)

func GetMemberQuotaCapacity(deptNode types.BudgetNode) float64 {
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

func BuildMemberBudgetQuota(member types.Member, pools map[string]types.MemberQuotaPool, platformKeys []types.PlatformKey) types.MemberBudgetQuota {
	return types.MemberBudgetQuota{
		MemberID: member.ID, MemberName: member.Name, DepartmentID: member.DepartmentID,
		PersonalQuota: GetPersonalQuota(pools, member.ID),
		Allocated:     GetAllocatedKeyQuota(platformKeys, member.ID),
		Used:          GetUsedKeyQuota(platformKeys, member.ID),
	}
}

func ValidateMemberQuotaUpdate(
	tree []types.BudgetNode,
	members []types.Member,
	pools map[string]types.MemberQuotaPool,
	platformKeys []types.PlatformKey,
	memberID string,
	personalQuota float64,
) *string {
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		msg := "Member not found"
		return &msg
	}

	allocated := GetAllocatedKeyQuota(platformKeys, memberID)
	if personalQuota < allocated {
		msg := fmt.Sprintf("个人额度不能低于已分配 Key 额度（¥%s）", formatMoney(allocated))
		return &msg
	}

	deptNode := FindBudgetNode(tree, member.DepartmentID)
	if deptNode == nil {
		msg := "Department budget node not found"
		return &msg
	}

	capacity := GetMemberQuotaCapacity(*deptNode)
	otherSum := 0.0
	for _, m := range members {
		if m.DepartmentID == member.DepartmentID && m.ID != memberID {
			otherSum += GetPersonalQuota(pools, m.ID)
		}
	}
	if otherSum+personalQuota > capacity {
		remaining := capacity - otherSum
		if remaining < 0 {
			remaining = 0
		}
		msg := fmt.Sprintf("超出部门可分配成员额度，当前剩余约 ¥%s", formatMoney(remaining))
		return &msg
	}
	return nil
}

func ApplyMemberQuotaUpdate(
	members []types.Member,
	pools map[string]types.MemberQuotaPool,
	platformKeys []types.PlatformKey,
	memberID string,
	personalQuota float64,
) types.MemberBudgetQuota {
	SetPersonalQuota(pools, memberID, personalQuota)
	member, ok := org.FindMemberByID(members, memberID)
	if !ok {
		return types.MemberBudgetQuota{
			MemberID: memberID, PersonalQuota: personalQuota,
			Allocated: GetAllocatedKeyQuota(platformKeys, memberID),
			Used:      GetUsedKeyQuota(platformKeys, memberID),
		}
	}
	return BuildMemberBudgetQuota(*member, pools, platformKeys)
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
