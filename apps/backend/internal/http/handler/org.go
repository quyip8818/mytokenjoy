package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	domainorg "github.com/tokenjoy/backend/internal/domain/org"
	"github.com/tokenjoy/backend/internal/http/response"
	"github.com/tokenjoy/backend/internal/pkg"
)

type OrgHandler struct {
	service domainorg.Service
}

func NewOrgHandler(service domainorg.Service) *OrgHandler {
	return &OrgHandler{service: service}
}

func (h *OrgHandler) DataSourceStatus(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetDataSourceStatus())
}

func (h *OrgHandler) DataSourceTest(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.TestDataSource(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *OrgHandler) DataSourceUpdate(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&struct{}{})
	h.service.UpdateDataSource()
	response.Void(w)
}

func (h *OrgHandler) DataSourceSearch(w http.ResponseWriter, r *http.Request) {
	keyword := r.URL.Query().Get("keyword")
	response.JSON(w, http.StatusOK, h.service.SearchDataSource(keyword))
}

func (h *OrgHandler) DataSourceImport(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.ImportDataSource(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *OrgHandler) DataSourceImportRetry(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.RetryImport(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *OrgHandler) SyncConfigGet(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetSyncConfig())
}

func (h *OrgHandler) SyncConfigUpdate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.SyncConfig
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	h.service.UpdateSyncConfig(body)
	response.Void(w)
}

func (h *OrgHandler) SyncTrigger(w http.ResponseWriter, r *http.Request) {
	result, err := h.service.TriggerSync(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, result)
}

func (h *OrgHandler) SyncLogs(w http.ResponseWriter, r *http.Request) {
	page := pkg.ParseIntParam(r.URL.Query().Get("page"), 1)
	pageSize := pkg.ParseIntParam(r.URL.Query().Get("pageSize"), 10)
	response.JSON(w, http.StatusOK, h.service.ListSyncLogs(page, pageSize))
}

func (h *OrgHandler) DepartmentTree(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.GetDepartmentTree())
}

func (h *OrgHandler) DepartmentCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name     string `json:"name"`
		ParentID string `json:"parentId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	response.JSON(w, http.StatusOK, h.service.CreateDepartment(body.Name, body.ParentID))
}

func (h *OrgHandler) DepartmentUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	id := chi.URLParam(r, "id")
	response.JSON(w, http.StatusOK, h.service.UpdateDepartment(id, body.Name))
}

func (h *OrgHandler) DepartmentDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteDepartment(id); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) MembersList(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	page := pkg.ParseIntParam(query.Get("page"), 1)
	pageSize := pkg.ParseIntParam(query.Get("pageSize"), 20)
	directOnly := query.Get("directOnly") == "true"
	response.JSON(w, http.StatusOK, h.service.ListMembers(
		query.Get("departmentId"), query.Get("keyword"), directOnly, page, pageSize,
	))
}

func (h *OrgHandler) MemberCreate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.Member
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	response.JSON(w, http.StatusOK, h.service.CreateMember(body))
}

func (h *OrgHandler) MemberUpdate(w http.ResponseWriter, r *http.Request) {
	var body domainorg.Member
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	id := chi.URLParam(r, "id")
	member, err := h.service.UpdateMember(id, body)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, member)
}

func (h *OrgHandler) MembersDelete(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.service.DeleteMembers(body.IDs); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) MembersStatus(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs    []string `json:"ids"`
		Status string   `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	if err := h.service.UpdateMemberStatus(body.IDs, body.Status); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) MembersTransfer(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs          []string `json:"ids"`
		DepartmentID string   `json:"departmentId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	if err := h.service.TransferMembers(body.IDs, body.DepartmentID); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) MembersInvite(w http.ResponseWriter, r *http.Request) {
	_ = json.NewDecoder(r.Body).Decode(&struct{}{})
	if err := h.service.InviteMember(); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) MembersBatchInvite(w http.ResponseWriter, r *http.Request) {
	var body struct {
		IDs []string `json:"ids"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	response.JSON(w, http.StatusOK, h.service.BatchInvite(body.IDs))
}

func (h *OrgHandler) MembersBatchImport(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Rows []domainorg.BatchImportRow `json:"rows"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	response.JSON(w, http.StatusOK, h.service.BatchImport(body.Rows))
}

func (h *OrgHandler) RolesList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListRoles())
}

func (h *OrgHandler) RoleCreate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	response.JSON(w, http.StatusOK, h.service.CreateRole(body.Name, body.Permissions))
}

func (h *OrgHandler) RoleUpdate(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name        string   `json:"name"`
		Permissions []string `json:"permissions"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		response.Error(w, http.StatusBadRequest, "Invalid request body")
		return
	}
	id := chi.URLParam(r, "id")
	role, err := h.service.UpdateRole(id, body.Name, body.Permissions)
	if err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.JSON(w, http.StatusOK, role)
}

func (h *OrgHandler) RoleDelete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if err := h.service.DeleteRole(id); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) RoleMembersList(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	response.JSON(w, http.StatusOK, h.service.ListRoleMembers(roleID))
}

func (h *OrgHandler) RoleMemberAdd(w http.ResponseWriter, r *http.Request) {
	var body struct {
		MemberID string `json:"memberId"`
	}
	_ = json.NewDecoder(r.Body).Decode(&body)
	roleID := chi.URLParam(r, "roleId")
	if err := h.service.AddRoleMember(roleID, body.MemberID); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) RoleMemberRemove(w http.ResponseWriter, r *http.Request) {
	roleID := chi.URLParam(r, "roleId")
	memberID := chi.URLParam(r, "memberId")
	if err := h.service.RemoveRoleMember(roleID, memberID); err != nil {
		if writeDomainError(w, err) {
			return
		}
		response.Error(w, http.StatusInternalServerError, "Internal server error")
		return
	}
	response.Void(w)
}

func (h *OrgHandler) PermissionsList(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, h.service.ListPermissions())
}

func (h *OrgHandler) RegisterRoutes(r chi.Router) {
	r.Get("/data-source/status", h.DataSourceStatus)
	r.Post("/data-source/test", h.DataSourceTest)
	r.Put("/data-source", h.DataSourceUpdate)
	r.Get("/data-source/search", h.DataSourceSearch)
	r.Post("/data-source/import", h.DataSourceImport)
	r.Post("/data-source/import/retry", h.DataSourceImportRetry)
	r.Get("/sync/config", h.SyncConfigGet)
	r.Put("/sync/config", h.SyncConfigUpdate)
	r.Post("/sync/trigger", h.SyncTrigger)
	r.Get("/sync/logs", h.SyncLogs)
	r.Get("/departments/tree", h.DepartmentTree)
	r.Post("/departments", h.DepartmentCreate)
	r.Put("/departments/{id}", h.DepartmentUpdate)
	r.Delete("/departments/{id}", h.DepartmentDelete)
	r.Get("/members", h.MembersList)
	r.Post("/members", h.MemberCreate)
	r.Put("/members/{id}", h.MemberUpdate)
	r.Delete("/members", h.MembersDelete)
	r.Put("/members/status", h.MembersStatus)
	r.Post("/members/transfer", h.MembersTransfer)
	r.Post("/members/invite", h.MembersInvite)
	r.Post("/members/batch-invite", h.MembersBatchInvite)
	r.Post("/members/batch-import", h.MembersBatchImport)
	r.Get("/roles", h.RolesList)
	r.Post("/roles", h.RoleCreate)
	r.Put("/roles/{id}", h.RoleUpdate)
	r.Delete("/roles/{id}", h.RoleDelete)
	r.Get("/roles/{roleId}/members", h.RoleMembersList)
	r.Post("/roles/{roleId}/members", h.RoleMemberAdd)
	r.Delete("/roles/{roleId}/members/{memberId}", h.RoleMemberRemove)
	r.Get("/permissions", h.PermissionsList)
}
