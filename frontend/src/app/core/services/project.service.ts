import { Injectable, signal } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable, tap } from 'rxjs';
import { environment } from '../../../environments/environment';
import { Project, CreateProjectDto, UpdateProjectDto, PaginatedResponse } from '../models';

@Injectable({
  providedIn: 'root'
})
export class ProjectService {
  private readonly apiUrl = `${environment.apiUrl}/projects`;
  private projectsSignal = signal<Project[]>([]);

  public projects = this.projectsSignal.asReadonly();

  constructor(private http: HttpClient) {}

  getAll(): Observable<Project[]> {
    return this.http.get<Project[]>(this.apiUrl).pipe(
      tap(projects => this.projectsSignal.set(projects))
    );
  }

  getAllPaginated(page: number, limit: number): Observable<PaginatedResponse<Project>> {
    const params = new HttpParams()
      .set('page', page.toString())
      .set('limit', limit.toString());

    return this.http.get<PaginatedResponse<Project>>(this.apiUrl, { params });
  }

  getById(id: string): Observable<Project> {
    return this.http.get<Project>(`${this.apiUrl}/${id}`);
  }

  create(dto: CreateProjectDto): Observable<Project> {
    return this.http.post<Project>(this.apiUrl, dto).pipe(
      tap(project => {
        const current = this.projectsSignal();
        this.projectsSignal.set([...current, project]);
      })
    );
  }

  update(id: string, dto: UpdateProjectDto): Observable<Project> {
    return this.http.put<Project>(`${this.apiUrl}/${id}`, dto).pipe(
      tap(updated => {
        const current = this.projectsSignal();
        const index = current.findIndex(p => p.id === id);
        if (index !== -1) {
          const newProjects = [...current];
          newProjects[index] = updated;
          this.projectsSignal.set(newProjects);
        }
      })
    );
  }

  delete(id: string): Observable<void> {
    return this.http.delete<void>(`${this.apiUrl}/${id}`).pipe(
      tap(() => {
        const current = this.projectsSignal();
        this.projectsSignal.set(current.filter(p => p.id !== id));
      })
    );
  }

  refreshProjects(): void {
    this.getAll().subscribe();
  }
}
