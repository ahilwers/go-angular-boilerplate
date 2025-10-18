import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router } from '@angular/router';
import { ProgressSpinnerModule } from 'primeng/progressspinner';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-callback',
  standalone: true,
  imports: [CommonModule, ProgressSpinnerModule],
  templateUrl: './callback.component.html',
  styleUrl: './callback.component.scss'
})
export class CallbackComponent implements OnInit {
  error: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private authService: AuthService
  ) {}

  async ngOnInit(): Promise<void> {
    try {
      // Get the hash fragment (everything after #)
      const hash = window.location.hash.substring(1); // Remove the leading #

      if (!hash) {
        throw new Error('No authentication response received');
      }

      await this.authService.handleCallback(hash);
    } catch (err) {
      console.error('Callback error:', err);
      this.error = err instanceof Error ? err.message : 'Authentication failed';
      setTimeout(() => {
        // Clear any stored state and redirect to home (which will trigger login)
        sessionStorage.clear();
        window.location.href = '/';
      }, 3000);
    }
  }
}
