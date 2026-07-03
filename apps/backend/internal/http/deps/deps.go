package deps

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainrelay "github.com/tokenjoy/backend/internal/domain/relay"
	"github.com/tokenjoy/backend/internal/domain/session"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/platformauth"
)

type Deps struct {
	Config       config.Config
	Logger       *slog.Logger
	SessionSvc   session.Service
	OrgSvc       domainorg.Service
	BudgetSvc    domainbudget.Service
	KeysSvc      domainkeys.Service
	ModelsSvc    domainmodels.Service
	DashboardSvc domaindashboard.Service
	AuditSvc     domainaudit.Service
	ReadModel    domainusage.ReadModel
	IngestSvc    domainusage.Ingestor
	CompanySvc   domaincompany.Service
	BillingSvc   domainbilling.Service
	PlatformSvc  platformauth.Service
	WalletSvc    domaincompany.WalletService
	CompanyGate  *domaincompany.Gate
	RelayGateway domainrelay.GatewayService
}
