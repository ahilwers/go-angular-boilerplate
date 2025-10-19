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

// List godoc
// @Summary      List projects
// @Description  Get all projects with optional pagination
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        page   query  int  false  "Page number (1-based)"
// @Param        limit  query  int  false  "Items per page"
// @Success      200  {array}   entities.Project
// @Success      200  {object}  map[string]interface{}  "Paginated response with data, total, page, and limit"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects [get]
func (h *ProjectHandler) List(w http.ResponseWriter, r *http.Request) {
	page, limit := parsePaginationParams(r)
	if page > 0 && limit > 0 {
		offset := (page - 1) * limit
		projects, total, err := h.service.FindAllPaginated(limit, offset)
		if err != nil {
			h.logger.Error("failed to list projects", "error", err)
			respondError(w, "Failed to list projects", http.StatusInternalServerError)
			return
		}

		response := map[string]interface{}{
			"data":  projects,
			"total": total,
			"page":  page,
			"limit": limit,
		}
		respondJSON(w, response, http.StatusOK)
		return
	}
	projects, err := h.service.FindAll()
	if err != nil {
		h.logger.Error("failed to list projects", "error", err)
		respondError(w, "Failed to list projects", http.StatusInternalServerError)
		return
	}
	respondJSON(w, projects, http.StatusOK)
}

// Get godoc
// @Summary      Get project by ID
// @Description  Get a single project by its ID
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Project ID"
// @Success      200  {object}  entities.Project
// @Failure      400  {object}  map[string]string  "Missing project ID"
// @Failure      404  {object}  map[string]string  "Project not found"
// @Security     BearerAuth
// @Router       /api/v1/projects/{id} [get]
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

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string `json:"name" example:"My Project"`
	Description string `json:"description" example:"A sample project description"`
}

// Create godoc
// @Summary      Create project
// @Description  Create a new project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        project  body      CreateProjectRequest  true  "Project to create"
// @Success      201      {object}  entities.Project
// @Failure      400      {object}  map[string]string  "Invalid request body or missing name"
// @Failure      500      {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects [post]
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

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	Name        string `json:"name" example:"Updated Project"`
	Description string `json:"description" example:"Updated description"`
}

// Update godoc
// @Summary      Update project
// @Description  Update an existing project
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id       path      string                 true  "Project ID"
// @Param        project  body      UpdateProjectRequest   true  "Project updates"
// @Success      200      {object}  entities.Project
// @Failure      400      {object}  map[string]string  "Invalid request body or missing name"
// @Failure      404      {object}  map[string]string  "Project not found"
// @Failure      500      {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects/{id} [put]
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

// Delete godoc
// @Summary      Delete project
// @Description  Delete a project by ID
// @Tags         projects
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Project ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string  "Missing project ID"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects/{id} [delete]
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
