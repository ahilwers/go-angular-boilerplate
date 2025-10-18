import { Injectable, Renderer2, RendererFactory2, inject } from '@angular/core';
import { DOCUMENT } from '@angular/common';

/**
 * Service to manage application theme based on system preferences.
 * Automatically detects and applies dark/light mode based on the user's system settings.
 */
@Injectable({
  providedIn: 'root'
})
export class ThemeService {
  private readonly document = inject(DOCUMENT);
  private readonly rendererFactory = inject(RendererFactory2);
  private renderer: Renderer2;
  private mediaQueryList: MediaQueryList | null = null;

  constructor() {
    this.renderer = this.rendererFactory.createRenderer(null, null);
  }

  /**
   * Initialize theme detection and apply the appropriate theme.
   * This should be called once when the application starts.
   */
  initializeTheme(): void {
    // Check if window and matchMedia are available (for SSR compatibility)
    if (typeof window === 'undefined' || !window.matchMedia) {
      return;
    }

    // Create a media query to detect dark mode preference
    this.mediaQueryList = window.matchMedia('(prefers-color-scheme: dark)');

    // Apply initial theme
    this.applyTheme(this.mediaQueryList.matches);

    // Listen for changes in system theme preference
    this.mediaQueryList.addEventListener('change', (event) => {
      this.applyTheme(event.matches);
    });
  }

  /**
   * Apply the theme by adding/removing the dark-mode class on the document element.
   * @param isDark - Whether to apply dark mode
   */
  private applyTheme(isDark: boolean): void {
    const htmlElement = this.document.documentElement;

    if (isDark) {
      this.renderer.addClass(htmlElement, 'dark-mode');
    } else {
      this.renderer.removeClass(htmlElement, 'dark-mode');
    }
  }

  /**
   * Check if dark mode is currently active.
   * @returns true if dark mode is active
   */
  isDarkMode(): boolean {
    return this.document.documentElement.classList.contains('dark-mode');
  }

  /**
   * Get the current system preference for dark mode.
   * @returns true if system prefers dark mode
   */
  getSystemPreference(): boolean {
    if (typeof window === 'undefined' || !window.matchMedia) {
      return false;
    }
    return window.matchMedia('(prefers-color-scheme: dark)').matches;
  }

  /**
   * Clean up event listeners when the service is destroyed.
   */
  destroy(): void {
    if (this.mediaQueryList) {
      // Note: Modern browsers support removeEventListener, but older ones might need removeListener
      this.mediaQueryList.removeEventListener('change', (event) => {
        this.applyTheme(event.matches);
      });
    }
  }
}
