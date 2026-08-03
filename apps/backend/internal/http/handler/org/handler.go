package org

import (
	"github.com/go-chi/chi/v5"
	domaincompany "github.com/tokenjoy/backend/internal/domain/company"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	httpdeps "github.com/tokenjoy/backend/internal/http/deps"
	"github.com/tokenjoy/backend/internal/http/handler/shared"
	httpmiddleware "github.com/tokenjoy/backend/internal/http/middleware"
	"github.com/tokenjoy/backend/internal/infra/permission"
)

type Handler struct {
	shared.ProtectedHandlerBase
	service    domainorg.Service
	companySvc domaincompany.Service
}

func NewHandler(p httpdeps.Protected, service domainorg.Service, companySvc domaincompany.Service) *Handler {
	return &Handler{
		ProtectedHandlerBase: shared.NewProtectedHandlerBase(p),
		service:              service,
		companySvc:           companySvc,
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	read := httpmiddleware.ReadRoutes(r, h.Protected, permission.OrgRead)
	read.Get("/data-source/status", h.DataSourceStatus)
	read.Get("/data-source/search", h.DataSourceSearch)
	read.Get("/data-source/field-mappings", h.FieldMappingsGet)
	read.Get("/sync/config", h.SyncConfigGet)
	read.Get("/sync/logs", h.SyncLogs)
	read.Get("/departments/tree", h.DepartmentTree)
	read.Get("/members", h.MembersList)
	read.Get("/roles", h.RolesList)
	read.Get("/roles/{roleId}/members", h.RoleMembersList)
	read.Get("/permissions", h.PermissionsList)

	write := httpmiddleware.ReadRoutes(r, h.Protected)

	adminWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgAdmin))
	adminWrite.Post("/data-source/test", h.DataSourceTest)
	adminWrite.Put("/data-source", h.DataSourceUpdate)
	adminWrite.Post("/data-source/import", h.DataSourceImport)
	adminWrite.Post("/data-source/import/retry", h.DataSourceImportRetry)
	adminWrite.Put("/data-source/field-mappings", h.FieldMappingsSave)
	adminWrite.Get("/data-source/field-mappings/test", h.FieldMappingsTest)
	adminWrite.Put("/sync/config", h.SyncConfigUpdate)

	r.With(httpmiddleware.AllowSyncTrigger(h.Protected, h.companySvc)).Post("/sync/trigger", h.SyncTrigger)

	adminWrite.Post("/departments", h.DepartmentCreate)
	adminWrite.Put("/departments/{id}", h.DepartmentUpdate)
	adminWrite.Delete("/departments/{id}", h.DepartmentDelete)

	manageWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgManage))
	manageWrite.Post("/members", h.MemberCreate)
	manageWrite.Put("/members/{id}", h.MemberUpdate)
	manageWrite.Put("/members/{id}/user", h.MemberUserUpdate)
	manageWrite.Delete("/members", h.MembersDelete)
	manageWrite.Put("/members/status", h.MembersStatus)
	manageWrite.Post("/members/transfer", h.MembersTransfer)
	manageWrite.Post("/members/{id}/invite-link", h.MemberInviteLink)
	manageWrite.Post("/members/batch-invite", h.MembersBatchInvite)
	manageWrite.Post("/members/batch-import", h.MembersBatchImport)

	adminWrite.Post("/roles", h.RoleCreate)
	adminWrite.Put("/roles/{id}", h.RoleUpdate)
	adminWrite.Delete("/roles/{id}", h.RoleDelete)
	adminWrite.Post("/roles/{roleId}/members", h.RoleMemberAdd)
	adminWrite.Delete("/roles/{roleId}/members/{memberId}", h.RoleMemberRemove)
}

// Mount registers the org handler on the given router under /org.
func Mount(r chi.Router, d httpdeps.Deps) {
	h := NewHandler(d.Protected(), d.OrgSvc, d.CompanySvc)
	r.Route("/org", h.RegisterRoutes)
}
