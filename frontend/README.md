# Todo App Frontend

Modern Angular 18+ frontend application for the Todo App boilerplate with Keycloak authentication and PrimeNG UI components.

## Features

- **[Angular 18+](https://angular.dev)** with standalone components
- **Angular Signals** for reactive state management
- **[PrimeNG](https://primeng.org)** UI component library
- **OpenID Connect / [Keycloak](https://www.keycloak.org)** authentication
- **Responsive Design** with clean, modern UI
- **Clean Architecture** with separation of concerns
- **Localization Support** through [ngx-translate](https://ngx-translate.org)

## Project Structure

```
src/
├── app/
│   ├── core/                    # Core functionality
│   │   ├── guards/              # Route guards (auth guard)
│   │   ├── interceptors/        # HTTP interceptors (auth interceptor)
│   │   ├── models/              # TypeScript interfaces and models
│   │   └── services/            # Business logic services
│   ├── features/                # Feature modules
│   │   ├── auth/                # Authentication (login, callback)
│   │   ├── projects/            # Project management
│   │   └── tasks/               # Task management 
│   ├── shared/                  # Shared components
│   │   └── components/
│   │       └── layout/          # Main layout (header, sidebar, footer)
│   ├── app.component.ts         # Root component
│   ├── app.config.ts            # Application configuration
│   └── app.routes.ts            # Routing configuration
├── environments/                # Environment configurations
└── styles.scss                  # Global styles
```

## Prerequisites

- Node.js 18+ and npm
- Angular CLI: `npm install -g @angular/cli`
- Backend API running (see backend README)
- Keycloak running (see docker-compose.yml)

## Getting Started

### 1. Install Dependencies

```bash
cd frontend
npm install
```

### 2. Configure Environment

Edit `src/environments/environment.ts` to match your setup:

```typescript
export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api/v1',
  auth: {
    enabled: true,
    issuer: 'http://localhost:8081/realms/boilerplate',
    clientId: 'boilerplate-client',
    redirectUri: 'http://localhost:4200/auth/callback',
    scope: 'openid profile email'
  }
};
```

**Note:** These values must match your Keycloak realm and client configuration in `config/local.yaml`.

### 3. Run Development Server

```bash
npm start
```

Navigate to `http://localhost:4200/`. The application will automatically reload if you change any of the source files.

## Available Scripts

- `npm start` - Start development server
- `npm run build` - Build for production
- `npm test` - Run unit tests
- `npm run watch` - Build in watch mode

## Key Technologies

### State Management with Signals

The application uses Angular Signals for reactive state management:

```typescript
// In services
private projectsSignal = signal<Project[]>([]);
public projects = this.projectsSignal.asReadonly();

// In components
projects = this.projectService.projects;
```

### Authentication Flow

1. User navigates to any route (e.g., `/` or `/projects`)
2. Auth guard detects no authentication and redirects to Keycloak login automatically
3. User logs in via Keycloak
4. Keycloak redirects back to `/auth/callback` with authorization code
5. Application exchanges code for JWT tokens
6. Tokens stored in localStorage
7. User redirected to originally requested page (or `/projects` by default)
8. HTTP interceptor adds Authorization header to all API requests

## Components

### Layout Component
- Header with app title and user menu
- Collapsible sidebar with navigation
- Footer displaying logged-in user
- Router outlet for page content

### Project List Component
- Table view of all projects
- CRUD operations (Create, Read, Update, Delete)
- Navigate to project tasks
- Pagination and sorting

### Task Board Component
- Kanban-style board with three columns (TODO, IN_PROGRESS, DONE)
- Status updates with quick action buttons
- CRUD operations for tasks
- Due date support
- Back navigation to projects

### Authentication Components
- Auth guard that automatically redirects to Keycloak when not authenticated
- Callback page for OAuth flow handling
- Token management and expiration handling
- Automatic logout on token expiration

## API Integration

All API calls go through services that use Angular's HttpClient:

```typescript
// ProjectService
getAll(): Observable<Project[]>
getById(id: string): Observable<Project>
create(dto: CreateProjectDto): Observable<Project>
update(id: string, dto: UpdateProjectDto): Observable<Project>
delete(id: string): Observable<void>

// TaskService
getTasksByProject(projectId: string): Observable<Task[]>
getById(id: string): Observable<Task>
create(projectId: string, dto: CreateTaskDto): Observable<Task>
update(id: string, dto: UpdateTaskDto): Observable<Task>
delete(id: string): Observable<void>
```

## Security

- All routes protected by `authGuard` (except login and callback)
- JWT tokens stored in localStorage
- Authorization header automatically added to API requests
- Token expiration checking
- Automatic logout on expired tokens

## Building for Production

```bash
npm run build
```

Build artifacts stored in `dist/todo-app/`. Deploy the contents to your web server.

### Production Environment

Update `src/environments/environment.prod.ts` with production URLs:

```typescript
export const environment = {
  production: true,
  apiUrl: '/api/v1',
  auth: {
    enabled: true,
    issuer: 'https://auth.yourdomain.com/realms/todo-app',
    clientId: 'todo-app-frontend',
    redirectUri: 'https://app.yourdomain.com/auth/callback',
    scope: 'openid profile email'
  }
};
```

## Troubleshooting

### CORS Issues
Ensure backend CORS settings allow requests from `http://localhost:4200` during development.

### Authentication Failures
- Verify Keycloak is running and accessible
- Check client ID and realm configuration
- Ensure redirect URIs are correctly configured in Keycloak

### API Connection Issues
- Verify backend is running on configured port
- Check network tab in browser DevTools for errors
- Ensure environment.ts has correct API URL

## License

See the LICENSE file in the project root.
