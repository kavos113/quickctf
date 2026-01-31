import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../services/auth.service';

@Component({
  selector: 'app-register',
  imports: [FormsModule, RouterLink],
  templateUrl: './register.html',
  styleUrl: './register.css',
})
export class RegisterComponent {
  private readonly authService = inject(AuthService);
  private readonly router = inject(Router);

  username = signal('');
  password = signal('');
  confirmPassword = signal('');
  errorMessage = signal('');
  successMessage = signal('');
  isLoading = signal(false);

  async onSubmit(): Promise<void> {
    this.errorMessage.set('');
    this.successMessage.set('');

    if (!this.username() || !this.password() || !this.confirmPassword()) {
      this.errorMessage.set('すべての項目を入力してください');
      return;
    }

    if (this.password() !== this.confirmPassword()) {
      this.errorMessage.set('パスワードが一致しません');
      return;
    }

    if (this.password().length < 8) {
      this.errorMessage.set('パスワードは8文字以上で入力してください');
      return;
    }

    this.isLoading.set(true);

    const result = await this.authService.register(this.username(), this.password());

    this.isLoading.set(false);

    if (result.success) {
      this.successMessage.set('登録が完了しました。ログインページへ移動します...');
      setTimeout(() => {
        this.router.navigate(['/login']);
      }, 2000);
    } else {
      this.errorMessage.set(result.error || '登録に失敗しました');
    }
  }

  updateUsername(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.username.set(input.value);
  }

  updatePassword(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.password.set(input.value);
  }

  updateConfirmPassword(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.confirmPassword.set(input.value);
  }
}
