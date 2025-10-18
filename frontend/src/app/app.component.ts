import { Component, OnInit, OnDestroy, inject } from '@angular/core';
import { RouterOutlet } from '@angular/router';
import { TranslateService } from '@ngx-translate/core';
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
  private readonly translate = inject(TranslateService);
  title = 'todo-app';

  ngOnInit(): void {
    // Initialize theme detection on app startup
    this.themeService.initializeTheme();

    // Initialize translation service with browser language detection
    this.initializeTranslation();
  }

  ngOnDestroy(): void {
    // Clean up theme service listeners
    this.themeService.destroy();
  }

  private async initializeTranslation(): Promise<void> {
    // Set available languages
    const availableLanguages = ['en', 'de'];
    this.translate.addLangs(availableLanguages);

    // Set default language as fallback
    this.translate.setDefaultLang('en');

    // Load translation files
    const translationPromises = availableLanguages.map(async lang => {
      const translations = await import(`../assets/i18n/${lang}.json`);
      this.translate.setTranslation(lang, translations.default || translations, false);
    });

    await Promise.all(translationPromises);

    // Get browser language
    const browserLang = this.translate.getBrowserLang();

    // Use browser language if available, otherwise use default (en)
    const langToUse = browserLang && availableLanguages.includes(browserLang) ? browserLang : 'en';
    this.translate.use(langToUse);
  }
}
