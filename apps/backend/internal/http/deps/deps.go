package deps

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/identity/authz"
	"github.com/tokenjoy/backend/internal/identity/credentials"
	"github.com/tokenjoy/backend/internal/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/store"
)

type Deps struct {
	Config               config.Config
	Logger               *slog.Logger
	Store                store.Store
	AuthzSvc             authz.Service
	Credentials          credentials.Service
	SessionToken         sessiontoken.Issuer
	PlatformSessionToken sessiontoken.Issuer
	OrgSvc               domainorg.Service
	BudgetSvc            domainbudget.Service
	KeysSvc              domainkeys.Service
	ModelsSvc            domainmodels.Service
	DashboardSvc         domaindashboard.Service
	AuditSvc             domainaudit.Service
	ReadModel            domainusage.ReadModel
	IngestSvc            domainusage.Ingestor
	IngestQueue          domainusage.Queue
	IngestMetrics        ingestmetrics.Recorder
	CompanySvc           domaincompany.Service
	BillingSvc           domainbilling.Service
	MemberAnalyticsSvc   domainmemberanalytics.Service
	WalletSvc            domaincompany.WalletService
	CompanyGate          *domaincompany.Gate
	NewAPIGateway        domaingateway.GatewayService
}
