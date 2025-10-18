import { Component, signal, OnInit, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule, Router } from '@angular/router';
import { ButtonModule } from 'primeng/button';
import { MenuModule } from 'primeng/menu';
import { AvatarModule } from 'primeng/avatar';
import { TooltipModule } from 'primeng/tooltip';
import { MenuItem } from 'primeng/api';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ButtonModule,
    MenuModule,
    AvatarModule,
    TooltipModule
  ],
  templateUrl: './layout.component.html',
  styleUrl: './layout.component.scss'
})
export class LayoutComponent implements OnInit {
  sidebarVisible = signal(true);
  isMobile = signal(false);

  menuItems: MenuItem[] = [
    {
      label: 'Projects',
      icon: 'pi pi-folder',
      command: () => this.router.navigate(['/projects'])
    }
  ];

  userMenuItems: MenuItem[] = [
    {
      label: 'Logout',
      icon: 'pi pi-sign-out',
      command: () => this.logout()
    }
  ];

  constructor(
    public authService: AuthService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.checkScreenSize();
  }

  @HostListener('window:resize', ['$event'])
  onResize(): void {
    this.checkScreenSize();
  }

  private checkScreenSize(): void {
    const width = window.innerWidth;
    const mobile = width <= 768;
    this.isMobile.set(mobile);

    // Auto-hide sidebar on mobile, show on desktop
    if (mobile) {
      this.sidebarVisible.set(false);
    } else {
      this.sidebarVisible.set(true);
    }
  }

  toggleSidebar(): void {
    this.sidebarVisible.update(v => !v);
  }

  logout(): void {
    this.authService.logout();
  }

  get displayName(): string {
    const user = this.authService.user();
    if (!user) return 'User';

    return user.name ||
           (user.given_name && user.family_name ? `${user.given_name} ${user.family_name}` : null) ||
           user.preferred_username ||
           user.email ||
           'User';
  }
}
