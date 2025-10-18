import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { ThemeService } from './core/services/theme.service';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent implements OnInit, OnDestroy {
  private readonly themeService = inject(ThemeService);
  title = 'todo-app';

  ngOnInit(): void {
    // Initialize theme detection on app startup
    this.themeService.initializeTheme();
  }

  ngOnDestroy(): void {
    // Clean up theme service listeners
    this.themeService.destroy();
  }
}
