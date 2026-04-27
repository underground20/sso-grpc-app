package handler

import (
	"app/internal/admin/response"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type CreateRoleRequest struct {
	Name        string   `json:"name"`
	Permissions []string `json:"permissions"`
}

type RoleCreator interface {
	CreateRole(ctx context.Context, name string, permissions []string) error
}

type RoleHandler struct {
	roleCreator RoleCreator
	logger      *slog.Logger
}

func NewRoleHandler(roleCreator RoleCreator, logger *slog.Logger) *RoleHandler {
	return &RoleHandler{
		roleCreator: roleCreator,
		logger:      logger,
	}
}

func (h *RoleHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req CreateRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.RespondWithError(w, http.StatusBadRequest, "'name' field is required")
		return
	}

	err := h.roleCreator.CreateRole(r.Context(), req.Name, req.Permissions)
	if err != nil {
		h.logger.Error("failed to create role", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, nil)
}
