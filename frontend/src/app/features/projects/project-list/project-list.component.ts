import { Component, OnInit, signal } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TableModule } from 'primeng/table';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextarea } from 'primeng/inputtextarea';
import { ToastModule } from 'primeng/toast';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { CardModule } from 'primeng/card';
import { TooltipModule } from 'primeng/tooltip';
import { MessageService, ConfirmationService } from 'primeng/api';
import { ProjectService } from '../../../core/services/project.service';
import { Project, CreateProjectDto, UpdateProjectDto } from '../../../core/models';

@Component({
  selector: 'app-project-list',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TableModule,
    ButtonModule,
    DialogModule,
    InputTextModule,
    InputTextarea,
    ToastModule,
    ConfirmDialogModule,
    CardModule,
    TooltipModule
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: './project-list.component.html',
  styleUrl: './project-list.component.scss'
})
export class ProjectListComponent implements OnInit {
  loading = signal(false);
  dialogVisible = signal(false);
  editMode = signal(false);
  projects = signal<Project[]>([]);
  totalRecords = signal(0);
  lastLazyLoadEvent: any = null;

  currentProject: Project | null = null;
  projectForm = {
    name: '',
    description: ''
  };

  constructor(
    public projectService: ProjectService,
    private messageService: MessageService,
    private confirmationService: ConfirmationService,
    private router: Router
  ) {}

  ngOnInit(): void {
    // Initial load will be triggered by lazy loading event
  }

  loadProjects(event: any): void {
    this.loading.set(true);
    this.lastLazyLoadEvent = event;

    const page = (event.first / event.rows) + 1; // PrimeNG uses 0-indexed offset
    const limit = event.rows;

    this.projectService.getAllPaginated(page, limit).subscribe({
      next: (response) => {
        this.projects.set(response.data);
        this.totalRecords.set(response.total);
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Failed to load projects'
        });
        console.error('Failed to load projects:', err);
      }
    });
  }

  reloadCurrentPage(): void {
    if (this.lastLazyLoadEvent) {
      this.loadProjects(this.lastLazyLoadEvent);
    }
  }

  openCreateDialog(): void {
    this.editMode.set(false);
    this.currentProject = null;
    this.projectForm = { name: '', description: '' };
    this.dialogVisible.set(true);
  }

  openEditDialog(project: Project): void {
    this.editMode.set(true);
    this.currentProject = project;
    this.projectForm = {
      name: project.name,
      description: project.description
    };
    this.dialogVisible.set(true);
  }

  closeDialog(): void {
    this.dialogVisible.set(false);
    this.projectForm = { name: '', description: '' };
    this.currentProject = null;
  }

  saveProject(): void {
    if (!this.projectForm.name.trim()) {
      this.messageService.add({
        severity: 'warn',
        summary: 'Validation Error',
        detail: 'Project name is required'
      });
      return;
    }

    if (this.editMode()) {
      this.updateProject();
    } else {
      this.createProject();
    }
  }

  createProject(): void {
    const dto: CreateProjectDto = {
      name: this.projectForm.name.trim(),
      description: this.projectForm.description.trim()
    };

    this.projectService.create(dto).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Success',
          detail: 'Project created successfully'
        });
        this.closeDialog();
        this.reloadCurrentPage();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Failed to create project'
        });
        console.error('Failed to create project:', err);
      }
    });
  }

  updateProject(): void {
    if (!this.currentProject) return;

    const dto: UpdateProjectDto = {
      name: this.projectForm.name.trim(),
      description: this.projectForm.description.trim()
    };

    this.projectService.update(this.currentProject.id, dto).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: 'Success',
          detail: 'Project updated successfully'
        });
        this.closeDialog();
        this.reloadCurrentPage();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: 'Error',
          detail: 'Failed to update project'
        });
        console.error('Failed to update project:', err);
      }
    });
  }

  deleteProject(project: Project): void {
    this.confirmationService.confirm({
      message: `Are you sure you want to delete "${project.name}"? This will also delete all associated tasks.`,
      header: 'Confirm Delete',
      icon: 'pi pi-exclamation-triangle',
      accept: () => {
        this.projectService.delete(project.id).subscribe({
          next: () => {
            this.messageService.add({
              severity: 'success',
              summary: 'Success',
              detail: 'Project deleted successfully'
            });
            this.reloadCurrentPage();
          },
          error: (err) => {
            this.messageService.add({
              severity: 'error',
              summary: 'Error',
              detail: 'Failed to delete project'
            });
            console.error('Failed to delete project:', err);
          }
        });
      }
    });
  }

  viewTasks(project: Project): void {
    this.router.navigate(['/projects', project.id, 'tasks']);
  }
}
