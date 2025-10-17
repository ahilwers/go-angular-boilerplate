package http

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// TaskHandler handles task-related HTTP requests
type TaskHandler struct {
	service service.TaskService
	logger  *slog.Logger
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(svc service.TaskService, logger *slog.Logger) *TaskHandler {
	return &TaskHandler{
		service: svc,
		logger:  logger,
	}
}

// ListByProject handles GET /api/v1/projects/{id}/tasks
func (h *TaskHandler) ListByProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if projectID == "" {
		respondError(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	tasks, err := h.service.FindByProjectID(projectID)
	if err != nil {
		h.logger.Error("failed to list tasks for project", "project_id", projectID, "error", err)
		respondError(w, "Failed to list tasks", http.StatusInternalServerError)
		return
	}

	respondJSON(w, tasks, http.StatusOK)
}

// Get handles GET /api/v1/tasks/{id}
func (h *TaskHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	task, err := h.service.FindByID(id)
	if err != nil {
		h.logger.Error("failed to get task", "id", id, "error", err)
		respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

// CreateForProject handles POST /api/v1/projects/{id}/tasks
func (h *TaskHandler) CreateForProject(w http.ResponseWriter, r *http.Request) {
	projectID := r.PathValue("id")
	if projectID == "" {
		respondError(w, "Missing project ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string     `json:"title"`
		Status      string     `json:"status"`
		DueDate     *time.Time `json:"due_date"`
		Description string     `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		respondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Set default status if not provided
	status := entities.TaskStatusTodo
	if req.Status != "" {
		var err error
		status, err = entities.ParseTaskStatus(req.Status)
		if err != nil {
			respondError(w, "Invalid status. Must be TODO, IN_PROGRESS, or DONE", http.StatusBadRequest)
			return
		}
	}

	task := &entities.Task{
		ProjectID:   projectID,
		Title:       req.Title,
		Status:      status,
		DueDate:     req.DueDate,
		Description: req.Description,
	}

	if err := h.service.Insert(task); err != nil {
		h.logger.Error("failed to create task", "error", err)
		respondError(w, "Failed to create task", http.StatusInternalServerError)
		return
	}

	respondJSON(w, task, http.StatusCreated)
}

// Update handles PUT /api/v1/tasks/{id}
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       string     `json:"title"`
		Status      string     `json:"status"`
		DueDate     *time.Time `json:"due_date"`
		Description string     `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Title == "" {
		respondError(w, "Title is required", http.StatusBadRequest)
		return
	}

	// Validate status
	status, err := entities.ParseTaskStatus(req.Status)
	if err != nil {
		respondError(w, "Invalid status. Must be TODO, IN_PROGRESS, or DONE", http.StatusBadRequest)
		return
	}

	// Fetch existing task to preserve timestamps and projectID
	existing, err := h.service.FindByID(id)
	if err != nil {
		h.logger.Error("failed to find task", "id", id, "error", err)
		respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	task := &entities.Task{
		ID:          id,
		ProjectID:   existing.ProjectID,
		Title:       req.Title,
		Status:      status,
		DueDate:     req.DueDate,
		Description: req.Description,
		CreatedAt:   existing.CreatedAt,
	}

	if err := h.service.Update(task); err != nil {
		h.logger.Error("failed to update task", "id", id, "error", err)
		respondError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

// Delete handles DELETE /api/v1/tasks/{id}
func (h *TaskHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(id); err != nil {
		h.logger.Error("failed to delete task", "id", id, "error", err)
		respondError(w, "Failed to delete task", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
