package deps

import (
	"log/slog"

	"github.com/tokenjoy/backend/internal/config"
	"github.com/tokenjoy/backend/internal/domain/adminport"
	domainapproval "github.com/tokenjoy/backend/internal/domain/approval"
	domainaudit "github.com/tokenjoy/backend/internal/domain/audit"
	domainbilling "github.com/tokenjoy/backend/internal/domain/billing"
	domainbudget "github.com/tokenjoy/backend/internal/domain/budget"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domaindashboard "github.com/tokenjoy/backend/internal/domain/dashboard"
	domaingateway "github.com/tokenjoy/backend/internal/domain/gateway"
	"github.com/tokenjoy/backend/internal/domain/identity/authz"
	"github.com/tokenjoy/backend/internal/domain/identity/credentials"
	"github.com/tokenjoy/backend/internal/domain/identity/sessiontoken"
	"github.com/tokenjoy/backend/internal/domain/identity/verifycode"
	domainkeys "github.com/tokenjoy/backend/internal/domain/keys"
	domainmemberanalytics "github.com/tokenjoy/backend/internal/domain/memberanalytics"
	domainmodels "github.com/tokenjoy/backend/internal/domain/models"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	domainusage "github.com/tokenjoy/backend/internal/domain/usage"
	"github.com/tokenjoy/backend/internal/infra/ingestmetrics"
	"github.com/tokenjoy/backend/internal/infra/jobs"
	"github.com/tokenjoy/backend/internal/infra/notification"
	"github.com/tokenjoy/backend/internal/integration/newapisync/devapi"
	"github.com/tokenjoy/backend/internal/store"
	pkgrl "github.com/tokenjoy/backend/internal/support/ratelimit"
)

type Deps struct {
	Config              config.Config
	Logger              *slog.Logger
	Store               store.Store
	AdminPort           adminport.Port
	AuthzSvc            authz.Service
	Credentials         credentials.Service
	SessionToken        sessiontoken.Issuer
	OrgSvc              domainorg.Service
	BudgetSvc           domainbudget.Service
	KeysSvc             domainkeys.Service
	ModelsSvc           domainmodels.Service
	DashboardSvc        domaindashboard.Service
	AuditSvc            domainaudit.Service
	ReadModel           domainusage.ReadModel
	IngestSvc           domainusage.Ingestor
	IngestEnqueuer      jobs.Enqueuer
	IngestMetrics       ingestmetrics.Recorder
	CompanySvc          domaincompany.Service
	BillingSvc          domainbilling.Service
	MemberAnalyticsSvc  domainmemberanalytics.Service
	CompanyGate         *domaincompany.Gate
	ApprovalEngine      *domainapproval.Engine
	Gateway             domaingateway.GatewayService
	DevBearerResolver   devapi.BearerResolver
	DevReadinessChecker devapi.ReadinessChecker
	NotificationSvc     *notification.Service
	RateLimiter         pkgrl.Limiter
	VerifyCodeSvc       *verifycode.Service
}

// Narrow repo accessors — avoid leaking Store into handler layer.

func (d Deps) Users() store.UserRepository       { return d.Store.User() }
func (d Deps) Sessions() store.SessionRepository { return d.Store.Session() }
func (d Deps) Invites() store.InviteRepository   { return d.Store.Invite() }
func (d Deps) Org() store.OrgRepository          { return d.Store.Org() }
func (d Deps) Budget() store.BudgetRepository    { return d.Store.Budget() }
func (d Deps) Company() store.CompanyRepository  { return d.Store.Company() }
func (d Deps) Notifications() store.NotificationRepository {
	return d.Store.Notification()
}
func (d Deps) NotificationPreferences() store.NotificationPreferenceRepository {
	return d.Store.NotificationPreference()
}
func (d Deps) ModelDiscount() store.ModelDiscountRepository {
	return d.Store.ModelDiscount()
}
