package memberbudgetquota

import (
	"fmt"

	"github.com/tokenjoy/backend/internal/domain/types"
	"github.com/tokenjoy/backend/internal/pkg/budgetutil"
	"github.com/tokenjoy/backend/internal/pkg/memberquota"
	"github.com/tokenjoy/backend/internal/pkg/queryutil"
)

func GetMemberQuotaCapacity(deptNode types.BudgetNode) float64 {
	reserved := 0.0
	if deptNode.ReservedPool != nil {
		reserved = *deptNode.ReservedPool
	}
	childrenSum := budgetutil.SumChildrenBudget(deptNode)
	capacity := deptNode.Budget - reserved - childrenSum
	if capacity < 0 {
		return 0
	}
	return capacity
}

func BuildMemberBudgetQuota(member types.Member, pools map[string]types.MemberQuotaPool, platformKeys []types.PlatformKey) types.MemberBudgetQuota {
	return types.MemberBudgetQuota{
		MemberID: member.ID, MemberName: member.Name, DepartmentID: member.DepartmentID,
		PersonalQuota: memberquota.GetPersonalQuota(pools, member.ID),
		Allocated:     memberquota.GetAllocatedKeyQuota(platformKeys, member.ID),
		Used:          memberquota.GetUsedKeyQuota(platformKeys, member.ID),
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
	member, ok := queryutil.FindMemberByID(members, memberID)
	if !ok {
		msg := "Member not found"
		return &msg
	}

	allocated := memberquota.GetAllocatedKeyQuota(platformKeys, memberID)
	if personalQuota < allocated {
		msg := fmt.Sprintf("个人额度不能低于已分配 Key 额度（¥%s）", formatMoney(allocated))
		return &msg
	}

	deptNode := budgetutil.FindBudgetNode(tree, member.DepartmentID)
	if deptNode == nil {
		msg := "Department budget node not found"
		return &msg
	}

	capacity := GetMemberQuotaCapacity(*deptNode)
	otherSum := 0.0
	for _, m := range members {
		if m.DepartmentID == member.DepartmentID && m.ID != memberID {
			otherSum += memberquota.GetPersonalQuota(pools, m.ID)
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
	memberquota.SetPersonalQuota(pools, memberID, personalQuota)
	member, ok := queryutil.FindMemberByID(members, memberID)
	if !ok {
		return types.MemberBudgetQuota{
			MemberID: memberID, PersonalQuota: personalQuota,
			Allocated: memberquota.GetAllocatedKeyQuota(platformKeys, memberID),
			Used:      memberquota.GetUsedKeyQuota(platformKeys, memberID),
		}
	}
	return BuildMemberBudgetQuota(*member, pools, platformKeys)
}

func formatMoney(value float64) string {
	return fmt.Sprintf("%.0f", value)
}
