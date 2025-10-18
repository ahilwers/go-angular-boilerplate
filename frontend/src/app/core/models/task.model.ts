export enum TaskStatus {
  TODO = 'TODO',
  IN_PROGRESS = 'IN_PROGRESS',
  DONE = 'DONE'
}

export interface Task {
  id: string;
  projectId: string;
  title: string;
  status: TaskStatus;
  dueDate?: string;
  description: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTaskDto {
  projectId: string;
  title: string;
  status?: TaskStatus;
  dueDate?: string;
  description?: string;
}

export interface UpdateTaskDto {
  title?: string;
  status?: TaskStatus;
  dueDate?: string;
  description?: string;
}
