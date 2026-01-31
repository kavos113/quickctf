import { Injectable, signal, effect } from '@angular/core';

export type Theme = 'dark' | 'light';

const THEME_KEY = 'app_theme';

@Injectable({
  providedIn: 'root',
})
export class ThemeService {
  readonly theme = signal<Theme>('dark');

  constructor() {
    this.loadTheme();
    effect(() => {
      this.applyTheme(this.theme());
    });
  }

  private loadTheme(): void {
    if (typeof window === 'undefined') return;

    const savedTheme = localStorage.getItem(THEME_KEY) as Theme | null;
    if (savedTheme && (savedTheme === 'dark' || savedTheme === 'light')) {
      this.theme.set(savedTheme);
    } else {
      // システムのテーマを確認
      const prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
      this.theme.set(prefersDark ? 'dark' : 'light');
    }
  }

  private applyTheme(theme: Theme): void {
    if (typeof document === 'undefined') return;

    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem(THEME_KEY, theme);
  }

  toggleTheme(): void {
    this.theme.set(this.theme() === 'dark' ? 'light' : 'dark');
  }

  isDark(): boolean {
    return this.theme() === 'dark';
  }
}
