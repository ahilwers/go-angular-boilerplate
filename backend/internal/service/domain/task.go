package domain

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service"
	"boilerplate/internal/storage"
)

type taskService struct {
	taskRepo storage.TaskRepository
}

func NewTaskService(taskRepo storage.TaskRepository) service.TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) Insert(task *entities.Task) error {
	return s.taskRepo.Insert(task)
}

func (s *taskService) Update(task *entities.Task) error {
	return s.taskRepo.Update(task)
}

func (s *taskService) Delete(id string) error {
	return s.taskRepo.Delete(id)
}

func (s *taskService) FindByID(id string) (entities.Task, error) {
	return s.taskRepo.FindByID(id)
}

func (s *taskService) FindAll() ([]entities.Task, error) {
	return s.taskRepo.FindAll()
}

func (s *taskService) FindByProjectID(projectID string) ([]entities.Task, error) {
	return s.taskRepo.FindByProjectID(projectID)
}
