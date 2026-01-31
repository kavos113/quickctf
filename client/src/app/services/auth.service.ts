import { Injectable, signal } from '@angular/core';
import { LoginRequest, LogoutRequest, RegisterRequest } from '../../gen/api/server/v1/client_pb';
import { userAuthClient } from './grpc-client';

export interface AuthState {
  isAuthenticated: boolean;
  token: string | null;
  userId: string | null;
  username: string | null;
}

@Injectable({
  providedIn: 'root',
})
export class AuthService {
  private readonly storageKey = 'auth_token';
  private readonly userKey = 'auth_user';

  readonly authState = signal<AuthState>({
    isAuthenticated: false,
    token: null,
    userId: null,
    username: null,
  });

  constructor() {
    this.loadStoredAuth();
  }

  private loadStoredAuth(): void {
    if (typeof window === 'undefined') return;

    const token = localStorage.getItem(this.storageKey);
    const user = localStorage.getItem(this.userKey);

    if (token && user) {
      const userData = JSON.parse(user);
      this.authState.set({
        isAuthenticated: true,
        token,
        userId: userData.userId,
        username: userData.username,
      });
    }
  }

  private saveAuth(token: string, userId: string, username: string): void {
    localStorage.setItem(this.storageKey, token);
    localStorage.setItem(this.userKey, JSON.stringify({ userId, username }));
    this.authState.set({
      isAuthenticated: true,
      token,
      userId,
      username,
    });
  }

  private clearAuth(): void {
    localStorage.removeItem(this.storageKey);
    localStorage.removeItem(this.userKey);
    this.authState.set({
      isAuthenticated: false,
      token: null,
      userId: null,
      username: null,
    });
  }

  async login(username: string, password: string): Promise<{ success: boolean; error?: string }> {
    try {
      const request = new LoginRequest({ username, password });
      const response = await userAuthClient.login(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      this.saveAuth(response.token, '', username);
      return { success: true };
    } catch (error) {
      console.error('Login error:', error);
      return { success: false, error: 'ログインに失敗しました' };
    }
  }

  async register(
    username: string,
    password: string,
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const request = new RegisterRequest({ username, password });
      const response = await userAuthClient.register(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true };
    } catch (error) {
      console.error('Register error:', error);
      return { success: false, error: '登録に失敗しました' };
    }
  }

  async logout(): Promise<void> {
    try {
      const token = this.authState().token;
      if (token) {
        const request = new LogoutRequest({ token });
        await userAuthClient.logout(request);
      }
    } catch (error) {
      console.error('Logout error:', error);
    } finally {
      this.clearAuth();
    }
  }

  isAuthenticated(): boolean {
    return this.authState().isAuthenticated;
  }

  getToken(): string | null {
    return this.authState().token;
  }
}
