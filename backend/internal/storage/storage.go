package storage

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/storage/mongodb"
	"errors"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
	ErrInvalidID     = errors.New("invalid id")
)

type TaskRepository interface {
	Insert(task *entities.Task) error
	Update(task *entities.Task) error
	Delete(id string) error
	FindByID(id string) (entities.Task, error)
	FindAll() ([]entities.Task, error)
	FindByProjectID(projectID string) ([]entities.Task, error)
}

type ProjectRepository interface {
	Insert(project *entities.Project) error
	Update(project *entities.Project) error
	Delete(id string) error
	FindByID(id string) (entities.Project, error)
	FindAll() ([]entities.Project, error)
}

type Repository struct {
	ProjectRepository ProjectRepository
	TaskRepository    TaskRepository
}

func NewRepository(client *mongo.Client, database string) Repository {
	return Repository{
		ProjectRepository: mongodb.NewProjectRepository(client, database),
		TaskRepository:    mongodb.NewTaskRepository(client, database),
	}
}
