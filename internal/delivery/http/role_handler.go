package http

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/zomzem/identity-service/internal/usecase"
)

type RoleHandler struct {
	roleUC usecase.RoleUseCase
}

func NewRoleHandler(r chi.Router, roleUC usecase.RoleUseCase) {
	handler := &RoleHandler{roleUC: roleUC}

	r.Get("/roles", handler.ListRoles)
	r.Post("/roles", handler.CreateRole)
	r.Get("/roles/{id}", handler.GetRoleByID)
	r.Put("/roles/{id}", handler.UpdateRole)
	r.Delete("/roles/{id}", handler.DeleteRole)

	r.Get("/permissions", handler.ListPermissions)
	r.Post("/roles/{id}/permissions", handler.AssignPermission)
	r.Delete("/roles/{roleId}/permissions/{permissionId}", handler.RemovePermission)
}

func (h *RoleHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	roles, err := h.roleUC.ListRoles(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, roles)
}

func (h *RoleHandler) GetRoleByID(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	role, err := h.roleUC.GetRoleByID(r.Context(), int32(id))
	if err != nil {
		http.Error(w, "Role not found", http.StatusNotFound)
		return
	}
	renderJSON(w, role)
}

func (h *RoleHandler) CreateRole(w http.ResponseWriter, r *http.Request) {
	var req usecase.CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	role, err := h.roleUC.CreateRole(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, role)
}

func (h *RoleHandler) UpdateRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var req usecase.UpdateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	role, err := h.roleUC.UpdateRole(r.Context(), int32(id), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, role)
}

func (h *RoleHandler) DeleteRole(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	if err := h.roleUC.DeleteRole(r.Context(), int32(id)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *RoleHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	perms, err := h.roleUC.ListPermissions(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	renderJSON(w, perms)
}

func (h *RoleHandler) AssignPermission(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, _ := strconv.Atoi(idStr)

	var req usecase.AssignPermissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	if err := h.roleUC.AssignPermission(r.Context(), int32(id), req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *RoleHandler) RemovePermission(w http.ResponseWriter, r *http.Request) {
	roleIdStr := chi.URLParam(r, "roleId")
	roleId, _ := strconv.Atoi(roleIdStr)
	permissionIdStr := chi.URLParam(r, "permissionId")
	permissionId, _ := strconv.Atoi(permissionIdStr)

	if err := h.roleUC.RemovePermission(r.Context(), int32(roleId), int32(permissionId)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func renderJSON(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}
