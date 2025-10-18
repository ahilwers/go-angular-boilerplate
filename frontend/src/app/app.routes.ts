import { Routes } from '@angular/router';
import { authGuard } from './core/guards/auth.guard';
import { LayoutComponent } from './shared/components/layout/layout.component';
import { CallbackComponent } from './features/auth/callback/callback.component';
import { ProjectListComponent } from './features/projects/project-list/project-list.component';
import { TaskBoardComponent } from './features/tasks/task-board/task-board.component';

export const routes: Routes = [
  {
    path: 'auth/callback',
    component: CallbackComponent
  },
  {
    path: '',
    component: LayoutComponent,
    canActivate: [authGuard],
    children: [
      {
        path: '',
        redirectTo: 'projects',
        pathMatch: 'full'
      },
      {
        path: 'projects',
        component: ProjectListComponent
      },
      {
        path: 'projects/:projectId/tasks',
        component: TaskBoardComponent
      },
      {
        path: 'tasks',
        redirectTo: 'projects',
        pathMatch: 'full'
      }
    ]
  },
  {
    path: '**',
    redirectTo: ''
  }
];
