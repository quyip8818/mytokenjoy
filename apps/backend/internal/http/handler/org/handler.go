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

	datasourceWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgDatasource))
	datasourceWrite.Post("/data-source/test", h.DataSourceTest)
	datasourceWrite.Put("/data-source", h.DataSourceUpdate)
	datasourceWrite.Post("/data-source/import", h.DataSourceImport)
	datasourceWrite.Post("/data-source/import/retry", h.DataSourceImportRetry)
	datasourceWrite.Put("/data-source/field-mappings", h.FieldMappingsSave)
	datasourceWrite.Get("/data-source/field-mappings/test", h.FieldMappingsTest)
	datasourceWrite.Put("/sync/config", h.SyncConfigUpdate)

	r.With(httpmiddleware.AllowSyncTrigger(h.Protected, h.companySvc)).Post("/sync/trigger", h.SyncTrigger)

	structureWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgStructure))
	structureWrite.Post("/departments", h.DepartmentCreate)
	structureWrite.Put("/departments/{id}", h.DepartmentUpdate)
	structureWrite.Delete("/departments/{id}", h.DepartmentDelete)

	membersWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgMembers))
	membersWrite.Post("/members", h.MemberCreate)
	membersWrite.Put("/members/{id}", h.MemberUpdate)
	membersWrite.Delete("/members", h.MembersDelete)
	membersWrite.Put("/members/status", h.MembersStatus)
	membersWrite.Post("/members/transfer", h.MembersTransfer)
	membersWrite.Post("/members/invite", h.MembersInvite)
	membersWrite.Post("/members/batch-invite", h.MembersBatchInvite)
	membersWrite.Post("/members/batch-import", h.MembersBatchImport)

	rolesWrite := write.With(httpmiddleware.RequireAnyPermission(permission.OrgRoles))
	rolesWrite.Post("/roles", h.RoleCreate)
	rolesWrite.Put("/roles/{id}", h.RoleUpdate)
	rolesWrite.Delete("/roles/{id}", h.RoleDelete)
	rolesWrite.Post("/roles/{roleId}/members", h.RoleMemberAdd)
	rolesWrite.Delete("/roles/{roleId}/members/{memberId}", h.RoleMemberRemove)
}
