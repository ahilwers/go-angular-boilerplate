import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { CommonModule } from '@angular/common';
import { Router } from '@angular/router';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { MenuModule } from 'primeng/menu';
import { MenuItem } from 'primeng/api';
import { Subject, takeUntil } from 'rxjs';

@Component({
  selector: 'app-sidebar-menu',
  standalone: true,
  imports: [
    CommonModule,
    TranslateModule,
    MenuModule
  ],
  templateUrl: './sidebar-menu.component.html',
  styleUrl: './sidebar-menu.component.scss'
})
export class SidebarMenuComponent implements OnInit, OnDestroy {
  private readonly translate = inject(TranslateService);
  private readonly router = inject(Router);
  private destroy$ = new Subject<void>();

  menuItems: MenuItem[] = [];

  ngOnInit(): void {
    // Initialize menu items
    this.updateMenuItems();

    // Subscribe to language changes to update menu items
    this.translate.onLangChange
      .pipe(takeUntil(this.destroy$))
      .subscribe(() => {
        this.updateMenuItems();
      });
  }

  ngOnDestroy(): void {
    this.destroy$.next();
    this.destroy$.complete();
  }

  private updateMenuItems(): void {
    this.menuItems = [
      {
        label: this.translate.instant('MENU.PROJECTS'),
        icon: 'pi pi-folder',
        command: () => this.router.navigate(['/projects'])
      }
    ];
  }
}
