package domain_test

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service/domain"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockTaskRepository struct {
	mock.Mock
}

func (m *MockTaskRepository) Insert(task *entities.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Update(task *entities.Task) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *MockTaskRepository) Delete(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockTaskRepository) FindByID(id string) (entities.Task, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return entities.Task{}, args.Error(1)
	}
	return args.Get(0).(entities.Task), args.Error(1)
}

func (m *MockTaskRepository) FindAll() ([]entities.Task, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Task), args.Error(1)
}

func (m *MockTaskRepository) FindByProjectID(projectID string) ([]entities.Task, error) {
	args := m.Called(projectID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]entities.Task), args.Error(1)
}

func createTestTask() entities.Task {
	return entities.Task{
		ID:        "test-task-id",
		Name:      "Test Task",
		ProjectID: "test-project-id",
	}
}

func setupMockForCreateTask(t *testing.T, mockRepo *MockTaskRepository, task entities.Task, expectedTask entities.Task, returnErr error) {
	t.Helper()
	mockRepo.On("Insert", mock.MatchedBy(func(t *entities.Task) bool {
		// Überprüfe, ob die übergebene Task die erwarteten Werte hat
		if task.Name != "" && t.Name != task.Name {
			return false
		}
		// Setze die erwarteten Werte auf die übergebene Task
		t.ID = expectedTask.ID
		t.Name = expectedTask.Name
		t.ProjectID = expectedTask.ProjectID
		return true
	})).Return(returnErr)
}

func TestTaskService_CreateTask(t *testing.T) {
	tests := []struct {
		name          string
		task          entities.Task
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository, entities.Task)
	}{
		{
			name: "successful creation",
			task: entities.Task{
				Name:      "New Task",
				ProjectID: "project-1",
			},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, task entities.Task) {
				setupMockForCreateTask(t, m, task, entities.Task{
					ID:        "new-task-id",
					Name:      task.Name,
					ProjectID: task.ProjectID,
				}, nil)
			},
		},
		{
			name: "empty name",
			task: entities.Task{
				Name:      "",
				ProjectID: "project-1",
			},
			expectedError: errors.New("name is required"),
			setupMock: func(t *testing.T, m *MockTaskRepository, task entities.Task) {
				setupMockForCreateTask(t, m, task, entities.Task{}, errors.New("name is required"))
			},
		},
		{
			name: "empty project id",
			task: entities.Task{
				Name:      "Task without project",
				ProjectID: "",
			},
			expectedError: errors.New("project ID is required"),
			setupMock: func(t *testing.T, m *MockTaskRepository, task entities.Task) {
				setupMockForCreateTask(t, m, task, entities.Task{}, errors.New("project ID is required"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.task)
			}

			service := domain.NewTaskService(mockRepo)
			taskToCreate := tt.task // Create a copy to avoid modifying the test case
			err := service.Insert(&taskToCreate)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, taskToCreate.ID)
				assert.Equal(t, tt.task.Name, taskToCreate.Name)
				assert.Equal(t, tt.task.ProjectID, taskToCreate.ProjectID)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForGetTask(t *testing.T, mockRepo *MockTaskRepository, taskID string, returnTask entities.Task, returnErr error) {
	t.Helper()
	mockRepo.On("FindByID", taskID).Return(returnTask, returnErr)
}

func TestTaskService_GetTask(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		expectedTask  entities.Task
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository, string)
	}{
		{
			name:   "task found",
			taskID: "existing-task-id",
			expectedTask: entities.Task{
				ID:        "existing-task-id",
				Name:      "Existing Task",
				ProjectID: "project-1",
			},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, id string) {
				setupMockForGetTask(t, m, id, entities.Task{
					ID:        id,
					Name:      "Existing Task",
					ProjectID: "project-1",
				}, nil)
			},
		},
		{
			name:          "task not found",
			taskID:        "non-existent-id",
			expectedTask:  entities.Task{},
			expectedError: errors.New("task not found"),
			setupMock: func(t *testing.T, m *MockTaskRepository, id string) {
				setupMockForGetTask(t, m, id, entities.Task{}, errors.New("task not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.taskID)
			}

			service := domain.NewTaskService(mockRepo)
			task, err := service.FindByID(tt.taskID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTask, task)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForUpdateTask(t *testing.T, mockRepo *MockTaskRepository, task entities.Task, returnErr error) {
	t.Helper()
	mockRepo.On("Update", mock.MatchedBy(func(t *entities.Task) bool {
		return t.ID == task.ID
	})).Return(returnErr)
}

func TestTaskService_UpdateTask(t *testing.T) {
	tests := []struct {
		name          string
		task          entities.Task
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository, entities.Task)
	}{
		{
			name: "successful update",
			task: entities.Task{
				ID:        "existing-task-id",
				Name:      "Updated Task",
				ProjectID: "project-1",
			},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, task entities.Task) {
				setupMockForUpdateTask(t, m, task, nil)
			},
		},
		{
			name: "task not found",
			task: entities.Task{
				ID:        "non-existent-id",
				Name:      "Non-existent Task",
				ProjectID: "project-1",
			},
			expectedError: errors.New("task not found"),
			setupMock: func(t *testing.T, m *MockTaskRepository, task entities.Task) {
				setupMockForUpdateTask(t, m, task, errors.New("task not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.task)
			}

			service := domain.NewTaskService(mockRepo)
			taskToUpdate := tt.task // Create a copy to avoid modifying the test case
			err := service.Update(&taskToUpdate)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForDeleteTask(t *testing.T, mockRepo *MockTaskRepository, taskID string, returnErr error) {
	t.Helper()
	mockRepo.On("Delete", taskID).Return(returnErr)
}

func TestTaskService_DeleteTask(t *testing.T) {
	tests := []struct {
		name          string
		taskID        string
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository, string)
	}{
		{
			name:          "successful deletion",
			taskID:        "existing-task-id",
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, id string) {
				setupMockForDeleteTask(t, m, id, nil)
			},
		},
		{
			name:          "task not found",
			taskID:        "non-existent-id",
			expectedError: errors.New("task not found"),
			setupMock: func(t *testing.T, m *MockTaskRepository, id string) {
				setupMockForDeleteTask(t, m, id, errors.New("task not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.taskID)
			}

			service := domain.NewTaskService(mockRepo)
			err := service.Delete(tt.taskID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForListTasks(t *testing.T, mockRepo *MockTaskRepository, returnTasks []entities.Task, returnErr error) {
	t.Helper()
	mockRepo.On("FindAll").Return(returnTasks, returnErr)
}

func TestTaskService_ListTasks(t *testing.T) {
	tests := []struct {
		name          string
		expectedTasks []entities.Task
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository)
	}{
		{
			name: "successful list",
			expectedTasks: []entities.Task{
				{ID: "task-1", Name: "Task 1", ProjectID: "project-1"},
				{ID: "task-2", Name: "Task 2", ProjectID: "project-1"},
			},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository) {
				setupMockForListTasks(t, m, []entities.Task{
					{ID: "task-1", Name: "Task 1", ProjectID: "project-1"},
					{ID: "task-2", Name: "Task 2", ProjectID: "project-1"},
				}, nil)
			},
		},
		{
			name:          "empty list",
			expectedTasks: []entities.Task{},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository) {
				setupMockForListTasks(t, m, []entities.Task{}, nil)
			},
		},
		{
			name:          "error listing tasks",
			expectedTasks: nil,
			expectedError: errors.New("failed to list tasks"),
			setupMock: func(t *testing.T, m *MockTaskRepository) {
				setupMockForListTasks(t, m, nil, errors.New("failed to list tasks"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo)
			}

			service := domain.NewTaskService(mockRepo)
			tasks, err := service.FindAll()

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTasks, tasks)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}

func setupMockForFindByProjectID(t *testing.T, mockRepo *MockTaskRepository, projectID string, returnTasks []entities.Task, returnErr error) {
	t.Helper()
	mockRepo.On("FindByProjectID", projectID).Return(returnTasks, returnErr)
}

func TestTaskService_FindByProjectID(t *testing.T) {
	tests := []struct {
		name          string
		projectID     string
		expectedTasks []entities.Task
		expectedError error
		setupMock     func(*testing.T, *MockTaskRepository, string)
	}{
		{
			name:      "tasks found",
			projectID: "project-1",
			expectedTasks: []entities.Task{
				{ID: "task-1", Name: "Task 1", ProjectID: "project-1"},
				{ID: "task-2", Name: "Task 2", ProjectID: "project-1"},
			},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, projectID string) {
				setupMockForFindByProjectID(t, m, projectID, []entities.Task{
					{ID: "task-1", Name: "Task 1", ProjectID: "project-1"},
					{ID: "task-2", Name: "Task 2", ProjectID: "project-1"},
				}, nil)
			},
		},
		{
			name:          "no tasks found",
			projectID:     "project-2",
			expectedTasks: []entities.Task{},
			expectedError: nil,
			setupMock: func(t *testing.T, m *MockTaskRepository, projectID string) {
				setupMockForFindByProjectID(t, m, projectID, []entities.Task{}, nil)
			},
		},
		{
			name:          "error finding tasks",
			projectID:     "project-3",
			expectedTasks: nil,
			expectedError: errors.New("failed to find tasks"),
			setupMock: func(t *testing.T, m *MockTaskRepository, projectID string) {
				setupMockForFindByProjectID(t, m, projectID, nil, errors.New("failed to find tasks"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := new(MockTaskRepository)
			if tt.setupMock != nil {
				tt.setupMock(t, mockRepo, tt.projectID)
			}

			service := domain.NewTaskService(mockRepo)
			tasks, err := service.FindByProjectID(tt.projectID)

			if tt.expectedError != nil {
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedTasks, tasks)
			}

			mockRepo.AssertExpectations(t)
		})
	}
}
