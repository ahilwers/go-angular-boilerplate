import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MenuModule } from 'primeng/menu';
import { AvatarModule } from 'primeng/avatar';
import { TooltipModule } from 'primeng/tooltip';
import { MenuItem } from 'primeng/api';
import { Subject, takeUntil } from 'rxjs';
import { AuthService } from '../../../core/services/auth.service';

@Component({
  selector: 'app-user-menu',
  standalone: true,
  imports: [
    CommonModule,
    TranslateModule,
    MenuModule,
    AvatarModule,
    TooltipModule
  ],
  templateUrl: './user-menu.component.html',
  styleUrl: './user-menu.component.scss'
})
export class UserMenuComponent implements OnInit, OnDestroy {
  private readonly translate = inject(TranslateService);
  private readonly authService = inject(AuthService);
  private destroy$ = new Subject<void>();

  userMenuItems: MenuItem[] = [];

  ngOnInit(): void {
    // Initialize user menu items
    this.updateUserMenuItems();

    // Subscribe to language changes to update menu items
    this.translate.onLangChange
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        this.updateUserMenuItems();
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private updateUserMenuItems(): void {
    this.userMenuItems = [
      {
        label: this.translate.instant('MENU.LOGOUT'),
        icon: 'pi pi-sign-out',
        command: () => this.logout()
      }
    ];
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
