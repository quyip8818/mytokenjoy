package seed

import "github.com/tokenjoy/backend/internal/domain/types"

const defaultPersonalQuota = 5000

func buildMemberQuotaPools() map[string]types.MemberQuotaPool {
	return map[string]types.MemberQuotaPool{
		IDMemberAdmin: {PersonalQuota: 50000},
		IDMember1:     {PersonalQuota: 10000},
		"m-2":         {PersonalQuota: 15000},
		"m-4":         {PersonalQuota: 12000},
		"m-auditor":   {PersonalQuota: 5000},
		"m-pure":      {PersonalQuota: 3000},
	}
}
