package domain_test

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service/domain"
	"boilerplate/internal/storage"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockProjectRepository struct {
	mock.Mock
}

func (m *MockProjectRepository) Insert(project *entities.Project) error {
	args := m.Called(project)
	// Die Methode gibt nur den Fehler zurück, der in den Testfällen definiert wurde
	return args.Error(0)
}

func (m *MockProjectRepository) Update(project *entities.Project) error {
	args := m.Called(project)
	return args.Error(0)
}

func (m *MockProjectRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockProjectRepository) FindByID(id string) (entities.Project, error) {
	args := m.Called(id)
	return args.Get(0).(entities.Project), args.Error(1)
}

func (m *MockProjectRepository) FindAll() ([]entities.Project, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Project), args.Error(1)
}

func createTestProject() entities.Project {
	return entities.Project{
		ID:   "test-id",
		Name: "Test Project",
	}
}

func setupMockForCreateProject(t *testing.T, mockRepo *MockProjectRepository, project entities.Project, expectedProject entities.Project, returnErr error) {
	t.Helper()
	mockRepo.On("Insert", mock.MatchedBy(func(p *entities.Project) bool {
		// Überprüfe, ob das übergebene Projekt die erwarteten Werte hat
		if project.Name != "" && p.Name != project.Name {
			return false
		}

		// Setze die erwartete ID und Name auf das übergebene Projekt
		p.ID = expectedProject.ID
		p.Name = expectedProject.Name
		return true
	})).Return(returnErr)
}

func TestProjectService_CreateProject(t *testing.T) {
	tests := []struct {
		name          string
		project       entities.Project
		expectedError error
		setupMock     func(*testing.T, *MockProjectRepository, entities.Project)
	}{
		{
			name:          "successful creation",
			project:       entities.Project{Name: "New Project"},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockProjectRepository, p entities.Project) {
				setupMockForCreateProject(t, m, p, entities.Project{ID: "new-id", Name: "New Project"}, nil)
			},
		},
		{
			name:          "empty name",
			project:       entities.Project{Name: ""},
			expectedError: errors.New("name is required"),
			setupMock: func(t *testing.T, m *MockProjectRepository, p entities.Project) {
				setupMockForCreateProject(t, m, p, entities.Project{}, errors.New("name is required"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.project)
			}

			service := domain.NewProjectService(mockRepo)
			projectToCreate := tt.project // Create a copy to avoid modifying the test case
			err := service.Insert(&projectToCreate)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, projectToCreate.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForGetProject(t *testing.T, mockRepo *MockProjectRepository, projectID string, returnProject entities.Project, returnErr error) {
	t.Helper()
	mockRepo.On("FindByID", projectID).Return(returnProject, returnErr)
}

func TestProjectService_GetProject(t *testing.T) {
	testProject := createTestProject()

	tests := []struct {
		name          string
		setupMock     func(*testing.T, *MockProjectRepository, string)
		projectID     string
		expectedError error
		expectedID    string
	}{
		{
			name:      "project found",
			projectID: testProject.ID,
			setupMock: func(t *testing.T, m *MockProjectRepository, id string) {
				setupMockForGetProject(t, m, id, testProject, nil)
			},
			expectedError: nil,
			expectedID:    testProject.ID,
		},
		{
			name:      "project not found",
			projectID: "non-existent",
			setupMock: func(t *testing.T, m *MockProjectRepository, id string) {
				setupMockForGetProject(t, m, id, entities.Project{}, storage.ErrNotFound)
			},
			expectedError: storage.ErrNotFound,
			expectedID:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.projectID)
			}

			service := domain.NewProjectService(mockRepo)
			project, err := service.FindByID(tt.projectID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, project.ID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForUpdateProject(t *testing.T, mockRepo *MockProjectRepository, project entities.Project, returnErr error) {
	t.Helper()
	mockRepo.On("Update", mock.MatchedBy(func(p *entities.Project) bool {
		return p.ID == project.ID && p.Name == project.Name
	})).Return(returnErr)
}

func TestProjectService_UpdateProject(t *testing.T) {
	testProject := createTestProject()

	tests := []struct {
		name          string
		setupMock     func(*testing.T, *MockProjectRepository, entities.Project)
		project       entities.Project
		expectedError error
	}{
		{
			name:    "successful update",
			project: testProject,
			setupMock: func(t *testing.T, m *MockProjectRepository, p entities.Project) {
				setupMockForUpdateProject(t, m, p, nil)
			},
			expectedError: nil,
		},
		{
			name:    "project not found",
			project: entities.Project{ID: "non-existent", Name: "Nonexistent"},
			setupMock: func(t *testing.T, m *MockProjectRepository, p entities.Project) {
				setupMockForUpdateProject(t, m, p, storage.ErrNotFound)
			},
			expectedError: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.project)
			}

			service := domain.NewProjectService(mockRepo)
			projectToUpdate := tt.project // Create a copy to avoid modifying the test case
			err := service.Update(&projectToUpdate)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForDeleteProject(t *testing.T, mockRepo *MockProjectRepository, projectID string, returnErr error) {
	t.Helper()
	mockRepo.On("Delete", projectID).Return(returnErr)
}

func TestProjectService_DeleteProject(t *testing.T) {
	tests := []struct {
		name          string
		setupMock     func(*testing.T, *MockProjectRepository, string)
		projectID     string
		expectedError error
	}{
		{
			name:      "successful deletion",
			projectID: "test-id",
			setupMock: func(t *testing.T, m *MockProjectRepository, id string) {
				setupMockForDeleteProject(t, m, id, nil)
			},
			expectedError: nil,
		},
		{
			name:      "project not found",
			projectID: "non-existent",
			setupMock: func(t *testing.T, m *MockProjectRepository, id string) {
				setupMockForDeleteProject(t, m, id, storage.ErrNotFound)
			},
			expectedError: storage.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.projectID)
			}

			service := domain.NewProjectService(mockRepo)
			err := service.Delete(tt.projectID)

			if tt.expectedError != nil {
				assert.ErrorIs(t, err, tt.expectedError)
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForListProjects(t *testing.T, mockRepo *MockProjectRepository, returnProjects []entities.Project, returnErr error) {
	t.Helper()
	mockRepo.On("FindAll").Return(returnProjects, returnErr)
}

func TestProjectService_ListProjects(t *testing.T) {
	testProjects := []entities.Project{
		{ID: "1", Name: "Project 1"},
		{ID: "2", Name: "Project 2"},
	}

	tests := []struct {
		name          string
		setupMock     func(*testing.T, *MockProjectRepository)
		expectedCount int
		expectedError error
	}{
		{
			name: "successful list",
			setupMock: func(t *testing.T, m *MockProjectRepository) {
				setupMockForListProjects(t, m, testProjects, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name: "empty list",
			setupMock: func(t *testing.T, m *MockProjectRepository) {
				setupMockForListProjects(t, m, []entities.Project{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name: "database error",
			setupMock: func(t *testing.T, m *MockProjectRepository) {
				setupMockForListProjects(t, m, nil, errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: errors.New("database error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockProjectRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo)
			}

			service := domain.NewProjectService(mockRepo)
			projects, err := service.FindAll()

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedCount, len(projects))
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
