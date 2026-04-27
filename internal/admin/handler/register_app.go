package handler

import (
	"app/internal/admin/response"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

type RegisterAppRequest struct {
	Name   string `json:"name"`
	Secret string `json:"secret"`
}

type AppCreator interface {
	RegisterApp(ctx context.Context, name string, secret string) (int, error)
}

type AppHandler struct {
	appCreator AppCreator
	logger     *slog.Logger
}

func NewAppHandler(appCreator AppCreator, logger *slog.Logger) *AppHandler {
	return &AppHandler{
		appCreator: appCreator,
		logger:     logger,
	}
}

func (h *AppHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req RegisterAppRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("failed to decode request body", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		response.RespondWithError(w, http.StatusBadRequest, "'name' field is required")
		return
	}

	_, err := h.appCreator.RegisterApp(r.Context(), req.Name, req.Secret)
	if err != nil {
		h.logger.Error("failed to register app", slog.String("error", err.Error()))
		response.RespondWithError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	response.RespondWithJSON(w, http.StatusCreated, response.SuccessResponse{Message: "App registered successfully"})
}
