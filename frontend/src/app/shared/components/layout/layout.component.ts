import { Component, signal, OnInit, HostListener } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { TranslateModule } from '@ngx-translate/core';
import { ButtonModule } from 'primeng/button';
import { TooltipModule } from 'primeng/tooltip';
import { SidebarMenuComponent } from '../sidebar-menu/sidebar-menu.component';
import { UserMenuComponent } from '../user-menu/user-menu.component';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-layout',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    TranslateModule,
    ButtonModule,
    TooltipModule,
    SidebarMenuComponent,
    UserMenuComponent
  ],
  templateUrl: './layout.component.html',
  styleUrl: './layout.component.scss'
})
export class LayoutComponent implements OnInit {
  sidebarVisible = signal(true);
  isMobile = signal(false);

  constructor(private authService: AuthService) {}

  ngOnInit(): void {
    this.checkScreenSize();
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
}
