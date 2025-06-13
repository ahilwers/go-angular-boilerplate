package storage

import "boilerplate/internal/entities"

type TaskRepository interface {
    Insert(task *entities.Task) error
    Update(task *entities.Task) error
    Delete(id string) error
    FindByID(id string) (entities.Task, error)
    FindAll() ([]entities.Task, error)
}

type ProjectRepository interface {
    Save(project entities.Project) (int, error)
    FindByID(id string) (entities.Project, error)
    FindAll() ([]entities.Project, error)
}

type Repository struct {
    ProjectRepository ProjectRepository
    TaskRepository    TaskRepository
}
