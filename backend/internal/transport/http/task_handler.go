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

// ListByProject godoc
// @Summary      List tasks by project
// @Description  Get all tasks for a specific project
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Project ID"
// @Success      200  {array}   entities.Task
// @Failure      400  {object}  map[string]string  "Missing project ID"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects/{id}/tasks [get]
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

// Get godoc
// @Summary      Get task by ID
// @Description  Get a single task by its ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path      string  true  "Task ID"
// @Success      200  {object}  entities.Task
// @Failure      400  {object}  map[string]string  "Missing task ID"
// @Failure      404  {object}  map[string]string  "Task not found"
// @Security     BearerAuth
// @Router       /api/v1/tasks/{id} [get]
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

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	Title       string     `json:"title" example:"Implement feature X"`
	Status      string     `json:"status" example:"TODO" enums:"TODO,IN_PROGRESS,DONE"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-12-31T23:59:59Z"`
	Description string     `json:"description" example:"Detailed task description"`
}

// CreateForProject godoc
// @Summary      Create task for project
// @Description  Create a new task for a specific project
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      string              true  "Project ID"
// @Param        task  body      CreateTaskRequest   true  "Task to create"
// @Success      201   {object}  entities.Task
// @Failure      400   {object}  map[string]string  "Invalid request body, missing title, or invalid status"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/projects/{id}/tasks [post]
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

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title       *string    `json:"title,omitempty" example:"Updated task title"`
	Status      *string    `json:"status,omitempty" example:"IN_PROGRESS" enums:"TODO,IN_PROGRESS,DONE"`
	DueDate     *time.Time `json:"due_date,omitempty" example:"2024-12-31T23:59:59Z"`
	Description *string    `json:"description,omitempty" example:"Updated description"`
}

// Update godoc
// @Summary      Update task
// @Description  Update an existing task (partial updates supported)
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      string             true  "Task ID"
// @Param        task  body      UpdateTaskRequest  true  "Task updates"
// @Success      200   {object}  entities.Task
// @Failure      400   {object}  map[string]string  "Invalid request body, empty title, or invalid status"
// @Failure      404   {object}  map[string]string  "Task not found"
// @Failure      500   {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/tasks/{id} [put]
func (h *TaskHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		respondError(w, "Missing task ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Title       *string    `json:"title"`
		Status      *string    `json:"status"`
		DueDate     *time.Time `json:"due_date"`
		Description *string    `json:"description"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Fetch existing task to preserve timestamps and projectID
	existing, err := h.service.FindByID(id)
	if err != nil {
		h.logger.Error("failed to find task", "id", id, "error", err)
		respondError(w, "Task not found", http.StatusNotFound)
		return
	}

	// Start with existing values
	task := &entities.Task{
		ID:          id,
		ProjectID:   existing.ProjectID,
		Title:       existing.Title,
		Status:      existing.Status,
		DueDate:     existing.DueDate,
		Description: existing.Description,
		CreatedAt:   existing.CreatedAt,
	}

	// Update only provided fields
	if req.Title != nil {
		if *req.Title == "" {
			respondError(w, "Title cannot be empty", http.StatusBadRequest)
			return
		}
		task.Title = *req.Title
	}

	if req.Status != nil {
		status, err := entities.ParseTaskStatus(*req.Status)
		if err != nil {
			respondError(w, "Invalid status. Must be TODO, IN_PROGRESS, or DONE", http.StatusBadRequest)
			return
		}
		task.Status = status
	}

	if req.DueDate != nil {
		task.DueDate = req.DueDate
	}

	if req.Description != nil {
		task.Description = *req.Description
	}

	if err := h.service.Update(task); err != nil {
		h.logger.Error("failed to update task", "id", id, "error", err)
		respondError(w, "Failed to update task", http.StatusInternalServerError)
		return
	}

	respondJSON(w, task, http.StatusOK)
}

// Delete godoc
// @Summary      Delete task
// @Description  Delete a task by ID
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id   path  string  true  "Task ID"
// @Success      204  "No Content"
// @Failure      400  {object}  map[string]string  "Missing task ID"
// @Failure      500  {object}  map[string]string  "Internal server error"
// @Security     BearerAuth
// @Router       /api/v1/tasks/{id} [delete]
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
