package service

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service/domain"
	"boilerplate/internal/storage"
)

// TaskService defines the interface for task-related operations
type TaskService interface {
	Insert(task *entities.Task) error
	Update(task *entities.Task) error
	Delete(id string) error
	FindByID(id string) (entities.Task, error)
	FindAll() ([]entities.Task, error)
	FindByProjectID(projectID string) ([]entities.Task, error)
}

// ProjectService defines the interface for project-related operations
type ProjectService interface {
	Insert(project *entities.Project) error
	Update(project *entities.Project) error
	Delete(id string) error
	FindByID(id string) (entities.Project, error)
	FindAll() ([]entities.Project, error)
}

// Service combines all services
type Service struct {
	Task    TaskService
	Project ProjectService
}

// NewService creates a new service instance with the given repositories
func NewService(repo *storage.Repository) *Service {
	return &Service{
		Task:    domain.NewTaskService(repo.TaskRepository),
		Project: domain.NewProjectService(repo.ProjectRepository),
	}
}
