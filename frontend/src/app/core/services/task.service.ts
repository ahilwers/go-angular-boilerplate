import { Injectable, signal } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Task, CreateTaskDto, UpdateTaskDto } from '../models';

@Injectable({
  providedIn: 'root'
})
export class TaskService {
  private readonly apiUrl = environment.apiUrl;
  private tasksSignal = signal<Task[]>([]);

  public tasks = this.tasksSignal.asReadonly();

  constructor(private http: HttpClient) {}

  getTasksByProject(projectId: string): Observable<Task[]> {
    return this.http.get<Task[]>(`${this.apiUrl}/projects/${projectId}/tasks`).pipe(
      tap(tasks => this.tasksSignal.set(tasks))
    );
  }

  getById(id: string): Observable<Task> {
    return this.http.get<Task>(`${this.apiUrl}/tasks/${id}`);
  }

  create(projectId: string, dto: CreateTaskDto): Observable<Task> {
    return this.http.post<Task>(`${this.apiUrl}/projects/${projectId}/tasks`, dto).pipe(
      tap(task => {
        const current = this.tasksSignal();
        this.tasksSignal.set([...current, task]);
      })
    );
  }

  update(id: string, dto: UpdateTaskDto): Observable<Task> {
    return this.http.put<Task>(`${this.apiUrl}/tasks/${id}`, dto).pipe(
      tap(updated => {
        const current = this.tasksSignal();
        const index = current.findIndex(t => t.id === id);
        if (index !== -1) {
          const newTasks = [...current];
          newTasks[index] = updated;
          this.tasksSignal.set(newTasks);
        }
      })
    );
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/tasks/${id}`).pipe(
      tap(() => {
        const current = this.tasksSignal();
        this.tasksSignal.set(current.filter(t => t.id !== id));
      })
    );
  }

  refreshTasks(projectId: string): void {
    this.getTasksByProject(projectId).subscribe();
  }

  clearTasks(): void {
    this.tasksSignal.set([]);
  }
}
