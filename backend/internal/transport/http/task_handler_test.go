package http

import (
	"boilerplate/internal/entities"
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

// Mock TaskService for testing
type mockTaskService struct {
	insertFunc         func(*entities.Task) error
	updateFunc         func(*entities.Task) error
	deleteFunc         func(string) error
	findByIDFunc       func(string) (entities.Task, error)
	findAllFunc        func() ([]entities.Task, error)
	findByProjectIDFunc func(string) ([]entities.Task, error)
}

func (m *mockTaskService) Insert(task *entities.Task) error {
	if m.insertFunc != nil {
		return m.insertFunc(task)
	}
	return nil
}

func (m *mockTaskService) Update(task *entities.Task) error {
	if m.updateFunc != nil {
		return m.updateFunc(task)
	}
	return nil
}

func (m *mockTaskService) Delete(id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(id)
	}
	return nil
}

func (m *mockTaskService) FindByID(id string) (entities.Task, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return entities.Task{}, errors.New("not found")
}

func (m *mockTaskService) FindAll() ([]entities.Task, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc()
	}
	return []entities.Task{}, nil
}

func (m *mockTaskService) FindByProjectID(projectID string) ([]entities.Task, error) {
	if m.findByProjectIDFunc != nil {
		return m.findByProjectIDFunc(projectID)
	}
	return []entities.Task{}, nil
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelError, // Only show errors in tests
	}))
}

func TestTaskHandler_ListByProject(t *testing.T) {
	mockService := &mockTaskService{
		findByProjectIDFunc: func(projectID string) ([]entities.Task, error) {
			if projectID == "123" {
				return []entities.Task{
					{
						ID:        "task1",
						ProjectID: "123",
						Title:     "Test Task",
						Status:    entities.TaskStatusTodo,
						CreatedAt: time.Now(),
						UpdatedAt: time.Now(),
					},
				}, nil
			}
			return []entities.Task{}, nil
		},
	}

	handler := NewTaskHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/123/tasks", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.ListByProject(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var tasks []entities.Task
	if err := json.NewDecoder(w.Body).Decode(&tasks); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}
}

func TestTaskHandler_Get(t *testing.T) {
	mockService := &mockTaskService{
		findByIDFunc: func(id string) (entities.Task, error) {
			if id == "task1" {
				return entities.Task{
					ID:        "task1",
					ProjectID: "123",
					Title:     "Test Task",
					Status:    entities.TaskStatusTodo,
					CreatedAt: time.Now(),
					UpdatedAt: time.Now(),
				}, nil
			}
			return entities.Task{}, errors.New("not found")
		},
	}

	handler := NewTaskHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/tasks/task1", nil)
	req.SetPathValue("id", "task1")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var task entities.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if task.ID != "task1" {
		t.Errorf("expected task ID 'task1', got '%s'", task.ID)
	}
}

func TestTaskHandler_CreateForProject(t *testing.T) {
	mockService := &mockTaskService{
		insertFunc: func(task *entities.Task) error {
			task.ID = "new-task-id"
			task.CreatedAt = time.Now()
			task.UpdatedAt = time.Now()
			return nil
		},
	}

	handler := NewTaskHandler(mockService, testLogger())

	reqBody := map[string]interface{}{
		"title":       "New Task",
		"status":      "TODO",
		"description": "New Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects/123/tasks", bytes.NewReader(body))
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.CreateForProject(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var task entities.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if task.Title != "New Task" {
		t.Errorf("expected task title 'New Task', got '%s'", task.Title)
	}

	if task.ProjectID != "123" {
		t.Errorf("expected project ID '123', got '%s'", task.ProjectID)
	}
}

func TestTaskHandler_Update(t *testing.T) {
	mockService := &mockTaskService{
		findByIDFunc: func(id string) (entities.Task, error) {
			if id == "task1" {
				return entities.Task{
					ID:        "task1",
					ProjectID: "123",
					Title:     "Old Title",
					Status:    entities.TaskStatusTodo,
					CreatedAt: time.Now(),
				}, nil
			}
			return entities.Task{}, errors.New("not found")
		},
		updateFunc: func(task *entities.Task) error {
			return nil
		},
	}

	handler := NewTaskHandler(mockService, testLogger())

	reqBody := map[string]interface{}{
		"title":       "Updated Task",
		"status":      "IN_PROGRESS",
		"description": "Updated Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/tasks/task1", bytes.NewReader(body))
	req.SetPathValue("id", "task1")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var task entities.Task
	if err := json.NewDecoder(w.Body).Decode(&task); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if task.Title != "Updated Task" {
		t.Errorf("expected task title 'Updated Task', got '%s'", task.Title)
	}

	if task.Status != entities.TaskStatusInProgress {
		t.Errorf("expected status 'IN_PROGRESS', got '%s'", task.Status)
	}
}

func TestTaskHandler_Delete(t *testing.T) {
	mockService := &mockTaskService{
		deleteFunc: func(id string) error {
			if id == "task1" {
				return nil
			}
			return errors.New("not found")
		},
	}

	handler := NewTaskHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/tasks/task1", nil)
	req.SetPathValue("id", "task1")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}
