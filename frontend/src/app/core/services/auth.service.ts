import { Injectable, signal, computed } from '@angular/core';
import { Router } from '@angular/router';
import { environment } from '../../../environments/environment';
import { User, AuthTokens } from '../models';

@Injectable({
  providedIn: 'root'
})
export class AuthService {
  private readonly storageKey = 'auth_tokens';
  private readonly userKey = 'auth_user';

  private tokensSignal = signal<AuthTokens | null>(this.loadTokensFromStorage());
  private userSignal = signal<User | null>(this.loadUserFromStorage());

  public tokens = this.tokensSignal.asReadonly();
  public user = this.userSignal.asReadonly();
  public isAuthenticated = computed(() => !!this.tokensSignal());

  constructor(private router: Router) {
    // Check token expiration on service initialization
    this.checkTokenExpiration();
  }

  login(): void {
    if (!environment.auth.enabled) {
      console.warn('Authentication is disabled');
      return;
    }

    const state = this.generateRandomState();
    const nonce = this.generateRandomState();

    const params = new URLSearchParams({
      client_id: environment.auth.clientId,
      redirect_uri: environment.auth.redirectUri,
      response_type: 'token id_token',
      scope: environment.auth.scope,
      state: state,
      nonce: nonce
    });

    sessionStorage.setItem('auth_state', state);
    sessionStorage.setItem('auth_nonce', nonce);
    window.location.href = `${environment.auth.issuer}/protocol/openid-connect/auth?${params.toString()}`;
  }

  async handleCallback(fragment: string): Promise<void> {
    try {
      // Parse the URL fragment (hash)
      const params = new URLSearchParams(fragment);
      const accessToken = params.get('access_token');
      const idToken = params.get('id_token');
      const state = params.get('state');
      const expiresIn = params.get('expires_in');
      const error = params.get('error');

      if (error) {
        throw new Error(`Authentication error: ${error}`);
      }

      if (!accessToken || !state) {
        throw new Error('Missing access token or state');
      }

      const savedState = sessionStorage.getItem('auth_state');
      if (state !== savedState) {
        throw new Error('Invalid state parameter');
      }

      sessionStorage.removeItem('auth_state');
      sessionStorage.removeItem('auth_nonce');

      const user = this.decodeToken(accessToken);

      const tokens: AuthTokens = {
        accessToken: accessToken,
        idToken: idToken || undefined,
        refreshToken: undefined,
        expiresIn: expiresIn ? parseInt(expiresIn, 10) : 3600
      };

      this.setTokens(tokens);
      this.setUser(user);

      // Get redirect URL or default to projects
      const redirectUrl = sessionStorage.getItem('redirect_url') || '/projects';
      sessionStorage.removeItem('redirect_url');

      await this.router.navigate([redirectUrl]);
    } catch (error) {
      console.error('Failed to handle auth callback:', error);
      throw error;
    }
  }

  logout(): void {
    this.clearTokens();
    this.clearUser();

    if (environment.auth.enabled) {
      const params = new URLSearchParams({
        client_id: environment.auth.clientId,
        post_logout_redirect_uri: window.location.origin
      });

      window.location.href = `${environment.auth.issuer}/protocol/openid-connect/logout?${params.toString()}`;
    } else {
      // If auth is disabled, just redirect to home which will trigger login
      this.router.navigate(['/']);
    }
  }

  getAccessToken(): string | null {
    return this.tokensSignal()?.accessToken || null;
  }


  private decodeToken(token: string): User {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) {
        throw new Error('Invalid token format');
      }

      const payload = JSON.parse(atob(parts[1]));
      return payload as User;
    } catch (error) {
      console.error('Failed to decode token:', error);
      throw error;
    }
  }

  private setTokens(tokens: AuthTokens): void {
    this.tokensSignal.set(tokens);
    localStorage.setItem(this.storageKey, JSON.stringify(tokens));

    // Set expiration timer
    if (tokens.expiresIn) {
      setTimeout(() => this.checkTokenExpiration(), tokens.expiresIn * 1000);
    }
  }

  private setUser(user: User): void {
    this.userSignal.set(user);
    localStorage.setItem(this.userKey, JSON.stringify(user));
  }

  private clearTokens(): void {
    this.tokensSignal.set(null);
    localStorage.removeItem(this.storageKey);
  }

  private clearUser(): void {
    this.userSignal.set(null);
    localStorage.removeItem(this.userKey);
  }

  private loadTokensFromStorage(): AuthTokens | null {
    const stored = localStorage.getItem(this.storageKey);
    return stored ? JSON.parse(stored) : null;
  }

  private loadUserFromStorage(): User | null {
    const stored = localStorage.getItem(this.userKey);
    return stored ? JSON.parse(stored) : null;
  }

  private checkTokenExpiration(): void {
    const tokens = this.tokensSignal();
    if (!tokens) return;

    try {
      const user = this.decodeToken(tokens.accessToken);
      const now = Date.now() / 1000;

      if (user.sub && typeof (user as any).exp === 'number' && (user as any).exp < now) {
        console.warn('Token expired, logging out');
        this.logout();
      }
    } catch (error) {
      console.error('Error checking token expiration:', error);
      this.logout();
    }
  }

  private generateRandomState(): string {
    const array = new Uint8Array(32);
    crypto.getRandomValues(array);
    return Array.from(array, byte => byte.toString(16).padStart(2, '0')).join('');
  }
}
