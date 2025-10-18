package http

import (
	"boilerplate/internal/entities"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// Mock ProjectService for testing
type mockProjectService struct {
	insertFunc           func(*entities.Project) error
	updateFunc           func(*entities.Project) error
	deleteFunc           func(string) error
	findByIDFunc         func(string) (entities.Project, error)
	findAllFunc          func() ([]entities.Project, error)
	findAllPaginatedFunc func(int, int) ([]entities.Project, int64, error)
}

func (m *mockProjectService) Insert(project *entities.Project) error {
	if m.insertFunc != nil {
		return m.insertFunc(project)
	}
	return nil
}

func (m *mockProjectService) Update(project *entities.Project) error {
	if m.updateFunc != nil {
		return m.updateFunc(project)
	}
	return nil
}

func (m *mockProjectService) Delete(id string) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(id)
	}
	return nil
}

func (m *mockProjectService) FindByID(id string) (entities.Project, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(id)
	}
	return entities.Project{}, errors.New("not found")
}

func (m *mockProjectService) FindAll() ([]entities.Project, error) {
	if m.findAllFunc != nil {
		return m.findAllFunc()
	}
	return []entities.Project{}, nil
}

func (m *mockProjectService) FindAllPaginated(limit, offset int) ([]entities.Project, int64, error) {
	if m.findAllPaginatedFunc != nil {
		return m.findAllPaginatedFunc(limit, offset)
	}
	return []entities.Project{}, 0, nil
}

func TestProjectHandler_List(t *testing.T) {
	mockService := &mockProjectService{
		findAllFunc: func() ([]entities.Project, error) {
			return []entities.Project{
				{
					ID:          "123",
					Name:        "Test Project",
					Description: "Test Description",
					CreatedAt:   time.Now(),
				},
			}, nil
		},
	}

	handler := NewProjectHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects", nil)
	w := httptest.NewRecorder()

	handler.List(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var projects []entities.Project
	if err := json.NewDecoder(w.Body).Decode(&projects); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if len(projects) != 1 {
		t.Errorf("expected 1 project, got %d", len(projects))
	}
}

func TestProjectHandler_Get(t *testing.T) {
	mockService := &mockProjectService{
		findByIDFunc: func(id string) (entities.Project, error) {
			if id == "123" {
				return entities.Project{
					ID:          "123",
					Name:        "Test Project",
					Description: "Test Description",
					CreatedAt:   time.Now(),
				}, nil
			}
			return entities.Project{}, errors.New("not found")
		},
	}

	handler := NewProjectHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodGet, "/api/v1/projects/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Get(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var project entities.Project
	if err := json.NewDecoder(w.Body).Decode(&project); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if project.ID != "123" {
		t.Errorf("expected project ID '123', got '%s'", project.ID)
	}
}

func TestProjectHandler_Create(t *testing.T) {
	mockService := &mockProjectService{
		insertFunc: func(project *entities.Project) error {
			project.ID = "new-id"
			project.CreatedAt = time.Now()
			return nil
		},
	}

	handler := NewProjectHandler(mockService, testLogger())

	reqBody := map[string]string{
		"name":        "New Project",
		"description": "New Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/projects", bytes.NewReader(body))
	w := httptest.NewRecorder()

	handler.Create(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, w.Code)
	}

	var project entities.Project
	if err := json.NewDecoder(w.Body).Decode(&project); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if project.Name != "New Project" {
		t.Errorf("expected project name 'New Project', got '%s'", project.Name)
	}

	if project.ID == "" {
		t.Error("expected project ID to be set")
	}
}

func TestProjectHandler_Update(t *testing.T) {
	mockService := &mockProjectService{
		findByIDFunc: func(id string) (entities.Project, error) {
			if id == "123" {
				return entities.Project{
					ID:        "123",
					Name:      "Old Name",
					CreatedAt: time.Now(),
				}, nil
			}
			return entities.Project{}, errors.New("not found")
		},
		updateFunc: func(project *entities.Project) error {
			return nil
		},
	}

	handler := NewProjectHandler(mockService, testLogger())

	reqBody := map[string]string{
		"name":        "Updated Project",
		"description": "Updated Description",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest(http.MethodPut, "/api/v1/projects/123", bytes.NewReader(body))
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Update(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	var project entities.Project
	if err := json.NewDecoder(w.Body).Decode(&project); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if project.Name != "Updated Project" {
		t.Errorf("expected project name 'Updated Project', got '%s'", project.Name)
	}
}

func TestProjectHandler_Delete(t *testing.T) {
	mockService := &mockProjectService{
		deleteFunc: func(id string) error {
			if id == "123" {
				return nil
			}
			return errors.New("not found")
		},
	}

	handler := NewProjectHandler(mockService, testLogger())

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/projects/123", nil)
	req.SetPathValue("id", "123")
	w := httptest.NewRecorder()

	handler.Delete(w, req)

	if w.Code != http.StatusNoContent {
		t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
	}
}
