package domain

import (
	"boilerplate/internal/entities"
	"boilerplate/internal/service"
	"boilerplate/internal/storage"
)

type projectService struct {
	projectRepo storage.ProjectRepository
}

func NewProjectService(projectRepo storage.ProjectRepository) service.ProjectService {
	return &projectService{
		projectRepo: projectRepo,
	}
}

func (s *projectService) Insert(project *entities.Project) error {
	return s.projectRepo.Insert(project)
}

func (s *projectService) Update(project *entities.Project) error {
	return s.projectRepo.Update(project)
}

func (s *projectService) Delete(id string) error {
	return s.projectRepo.Delete(id)
}

func (s *projectService) FindByID(id string) (entities.Project, error) {
	return s.projectRepo.FindByID(id)
}

func (s *projectService) FindAll() ([]entities.Project, error) {
	return s.projectRepo.FindAll()
}
