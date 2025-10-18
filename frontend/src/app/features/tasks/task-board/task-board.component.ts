import { Component, OnInit, signal, computed, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { CdkDragDrop, DragDropModule, moveItemInArray, transferArrayItem } from '@angular/cdk/drag-drop';
import { CardModule } from 'primeng/card';
import { ButtonModule } from 'primeng/button';
import { DialogModule } from 'primeng/dialog';
import { InputTextModule } from 'primeng/inputtext';
import { InputTextarea } from 'primeng/inputtextarea';
import { DropdownModule } from 'primeng/dropdown';
import { CalendarModule } from 'primeng/calendar';
import { ToastModule } from 'primeng/toast';
import { ConfirmDialogModule } from 'primeng/confirmdialog';
import { ChipModule } from 'primeng/chip';
import { TooltipModule } from 'primeng/tooltip';
import { MessageService, ConfirmationService } from 'primeng/api';
import { TaskService } from '../../../core/services/task.service';
import { ProjectService } from '../../../core/services/project.service';
import { Task, TaskStatus, CreateTaskDto, UpdateTaskDto } from '../../../core/models';

@Component({
  selector: 'app-task-board',
  standalone: true,
  imports: [
    CommonModule,
    FormsModule,
    TranslateModule,
    DragDropModule,
    CardModule,
    ButtonModule,
    DialogModule,
    InputTextModule,
    InputTextarea,
    DropdownModule,
    CalendarModule,
    ToastModule,
    ConfirmDialogModule,
    ChipModule,
    TooltipModule
  ],
  providers: [MessageService, ConfirmationService],
  templateUrl: './task-board.component.html',
  styleUrl: './task-board.component.scss'
})
export class TaskBoardComponent implements OnInit {
  private readonly translate = inject(TranslateService);
  projectId = signal<string | null>(null);
  project = computed(() => {
    const id = this.projectId();
    if (!id) return null;
    return this.projectService.projects().find(p => p.id === id);
  });

  loading = signal(false);
  dialogVisible = signal(false);
  editMode = signal(false);

  get tasks() {
    return this.taskService.tasks;
  }

  todoTasks = computed(() => this.tasks().filter(t => t.status === TaskStatus.TODO));
  inProgressTasks = computed(() => this.tasks().filter(t => t.status === TaskStatus.IN_PROGRESS));
  doneTasks = computed(() => this.tasks().filter(t => t.status === TaskStatus.DONE));

  currentTask: Task | null = null;
  taskForm = {
    title: '',
    description: '',
    status: TaskStatus.TODO,
    dueDate: null as Date | null
  };

  statusOptions = [
    { label: this.translate.instant('TASKS.STATUS.TODO'), value: TaskStatus.TODO },
    { label: this.translate.instant('TASKS.STATUS.IN_PROGRESS'), value: TaskStatus.IN_PROGRESS },
    { label: this.translate.instant('TASKS.STATUS.DONE'), value: TaskStatus.DONE }
  ];

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private taskService: TaskService,
    private projectService: ProjectService,
    private messageService: MessageService,
    private confirmationService: ConfirmationService
  ) {}

  ngOnInit(): void {
    this.route.paramMap.subscribe(params => {
      const id = params.get('projectId');
      this.projectId.set(id);
      if (id) {
        this.loadProject(id);
        this.loadTasks(id);
      }
    });
  }

  loadProject(id: string): void {
    if (!this.project()) {
      this.projectService.getById(id).subscribe({
        error: (err) => {
          this.messageService.add({
            severity: 'error',
            summary: this.translate.instant('ERRORS.ERROR'),
            detail: this.translate.instant('TASKS.MESSAGES.ERROR_LOAD_PROJECT')
          });
          console.error('Failed to load project:', err);
          this.router.navigate(['/projects']);
        }
      });
    }
  }

  loadTasks(projectId: string): void {
    this.loading.set(true);
    this.taskService.getTasksByProject(projectId).subscribe({
      next: () => {
        this.loading.set(false);
      },
      error: (err) => {
        this.loading.set(false);
        this.messageService.add({
          severity: 'error',
          summary: this.translate.instant('ERRORS.ERROR'),
          detail: this.translate.instant('TASKS.MESSAGES.ERROR_LOAD')
        });
        console.error('Failed to load tasks:', err);
      }
    });
  }

  openCreateDialog(status: TaskStatus = TaskStatus.TODO): void {
    this.editMode.set(false);
    this.currentTask = null;
    this.taskForm = {
      title: '',
      description: '',
      status: status,
      dueDate: null
    };
    this.dialogVisible.set(true);
  }

  openEditDialog(task: Task): void {
    this.editMode.set(true);
    this.currentTask = task;
    this.taskForm = {
      title: task.title,
      description: task.description,
      status: task.status,
      dueDate: task.dueDate ? new Date(task.dueDate) : null
    };
    this.dialogVisible.set(true);
  }

  closeDialog(): void {
    this.dialogVisible.set(false);
    this.taskForm = {
      title: '',
      description: '',
      status: TaskStatus.TODO,
      dueDate: null
    };
    this.currentTask = null;
  }

  saveTask(): void {
    if (!this.taskForm.title.trim()) {
      this.messageService.add({
        severity: 'warn',
        summary: this.translate.instant('ERRORS.VALIDATION_ERROR'),
        detail: this.translate.instant('TASKS.MESSAGES.VALIDATION_TITLE_REQUIRED')
      });
      return;
    }

    if (this.editMode()) {
      this.updateTask();
    } else {
      this.createTask();
    }
  }

  createTask(): void {
    const projectId = this.projectId();
    if (!projectId) return;

    const dto: CreateTaskDto = {
      projectId: projectId,
      title: this.taskForm.title.trim(),
      description: this.taskForm.description.trim(),
      status: this.taskForm.status,
      dueDate: this.taskForm.dueDate?.toISOString()
    };

    this.taskService.create(projectId, dto).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: this.translate.instant('ERRORS.SUCCESS'),
          detail: this.translate.instant('TASKS.MESSAGES.CREATED')
        });
        this.closeDialog();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translate.instant('ERRORS.ERROR'),
          detail: this.translate.instant('TASKS.MESSAGES.ERROR_CREATE')
        });
        console.error('Failed to create task:', err);
      }
    });
  }

  updateTask(): void {
    if (!this.currentTask) return;

    const dto: UpdateTaskDto = {
      title: this.taskForm.title.trim(),
      description: this.taskForm.description.trim(),
      status: this.taskForm.status,
      dueDate: this.taskForm.dueDate?.toISOString()
    };

    this.taskService.update(this.currentTask.id, dto).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: this.translate.instant('ERRORS.SUCCESS'),
          detail: this.translate.instant('TASKS.MESSAGES.UPDATED')
        });
        this.closeDialog();
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translate.instant('ERRORS.ERROR'),
          detail: this.translate.instant('TASKS.MESSAGES.ERROR_UPDATE')
        });
        console.error('Failed to update task:', err);
      }
    });
  }

  deleteTask(task: Task): void {
    this.confirmationService.confirm({
      message: this.translate.instant('TASKS.DELETE_CONFIRM', { title: task.title }),
      header: this.translate.instant('TASKS.CONFIRM_DELETE_HEADER'),
      icon: 'pi pi-exclamation-triangle',
      accept: () => {
        this.taskService.delete(task.id).subscribe({
          next: () => {
            this.messageService.add({
              severity: 'success',
              summary: this.translate.instant('ERRORS.SUCCESS'),
              detail: this.translate.instant('TASKS.MESSAGES.DELETED')
            });
          },
          error: (err) => {
            this.messageService.add({
              severity: 'error',
              summary: this.translate.instant('ERRORS.ERROR'),
              detail: this.translate.instant('TASKS.MESSAGES.ERROR_DELETE')
            });
            console.error('Failed to delete task:', err);
          }
        });
      }
    });
  }

  changeTaskStatus(task: Task, newStatus: TaskStatus): void {
    const dto: UpdateTaskDto = { status: newStatus };

    this.taskService.update(task.id, dto).subscribe({
      next: () => {
        this.messageService.add({
          severity: 'success',
          summary: this.translate.instant('ERRORS.SUCCESS'),
          detail: this.translate.instant('TASKS.MESSAGES.STATUS_UPDATED')
        });
      },
      error: (err) => {
        this.messageService.add({
          severity: 'error',
          summary: this.translate.instant('ERRORS.ERROR'),
          detail: this.translate.instant('TASKS.MESSAGES.ERROR_STATUS')
        });
        console.error('Failed to update task status:', err);
      }
    });
  }

  backToProjects(): void {
    this.taskService.clearTasks();
    this.router.navigate(['/projects']);
  }

  getStatusLabel(status: TaskStatus): string {
    const option = this.statusOptions.find(o => o.value === status);
    return option?.label || status;
  }

  getStatusSeverity(status: TaskStatus): string {
    switch (status) {
      case TaskStatus.TODO:
        return 'info';
      case TaskStatus.IN_PROGRESS:
        return 'warning';
      case TaskStatus.DONE:
        return 'success';
      default:
        return 'info';
    }
  }

  onTaskDrop(event: CdkDragDrop<Task[]>, targetStatus: TaskStatus): void {
    const task = event.item.data as Task;

    // If task is dropped in the same column, do nothing
    if (task.status === targetStatus) {
      return;
    }

    // Update task status
    this.changeTaskStatus(task, targetStatus);
  }
}
