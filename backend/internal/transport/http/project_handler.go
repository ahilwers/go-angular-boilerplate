package http

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
)

// ProjectHandler handles project-related HTTP requests
type ProjectHandler struct {
	service service.ProjectService
	logger  *slog.Logger
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(svc service.ProjectService, logger *slog.Logger) *ProjectHandler {
	return &ProjectHandler{
		service: svc,
		logger:  logger,
	}
}

// List handles GET /api/v1/projects
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	projects, err := h.service.FindAll()
	if err != nil {
		h.logger.Error("failed to list projects", "error", err)
		respondError(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}

	respondJSON(w, projects, http.StatusOK)
}

// Get handles GET /api/v1/projects/{id}
func (h *ProjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	project, err := h.service.FindByID(id)
	if err != nil {
		h.logger.Error("failed to get project", "id", id, "error", err)
		respondError(w, "Project not found", http.StatusNotFound)
		return
	}

	respondJSON(w, project, http.StatusOK)
}

// Create handles POST /api/v1/projects
func (h *ProjectHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		respondError(w, "Name is required", http.StatusBadRequest)
		return
	}

	project := &entities.Project{
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.service.Insert(project); err != nil {
		h.logger.Error("failed to create project", "error", err)
		respondError(w, "Failed to create project", http.StatusInternalServerError)
		return
	}

	respondJSON(w, project, http.StatusCreated)
}

// Update handles PUT /api/v1/projects/{id}
func (h *ProjectHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name == "" {
		respondError(w, "Name is required", http.StatusBadRequest)
		return
	}

	// Fetch existing project to preserve timestamps
	existing, err := h.service.FindByID(id)
	if err != nil {
		h.logger.Error("failed to find project", "id", id, "error", err)
		respondError(w, "Project not found", http.StatusNotFound)
		return
	}

	project := &entities.Project{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   existing.CreatedAt,
	}

	if err := h.service.Update(project); err != nil {
		h.logger.Error("failed to update project", "id", id, "error", err)
		respondError(w, "Failed to update project", http.StatusInternalServerError)
		return
	}

	respondJSON(w, project, http.StatusOK)
}

// Delete handles DELETE /api/v1/projects/{id}
func (h *ProjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		h.logger.Error("failed to delete project", "id", id, "error", err)
		respondError(w, "Failed to delete project", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

