package memory

import (
	"github.com/tokenjoy/backend/internal/store"
)

var (
	_ store.CompanyRepository    = (*memoryCompanyRepo)(nil)
	_ store.OrgRepository        = (*memoryOrgRepo)(nil)
	_ store.BudgetRepository     = (*memoryBudgetRepo)(nil)
	_ store.KeysRepository       = (*memoryKeysRepo)(nil)
	_ store.ModelsRepository     = (*memoryModelsRepo)(nil)
	_ store.AuditRepository      = (*memoryAuditRepo)(nil)
	_ store.RelayRepository      = (*memoryRelayRepo)(nil)
	_ store.CredentialRepository = (*memoryCredentialRepo)(nil)
)
