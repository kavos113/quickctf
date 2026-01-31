import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Router } from '@angular/router';
import { AdminService } from '../../../services/admin.service';

@Component({
  selector: 'app-admin-activate',
  imports: [FormsModule],
  templateUrl: './admin-activate.html',
  styleUrl: './admin-activate.css',
})
export class AdminActivateComponent {
  private readonly adminService = inject(AdminService);
  private readonly router = inject(Router);

  activationCode = '';
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);

  async onSubmit(): Promise<void> {
    if (!this.activationCode.trim()) {
      this.error.set('アクティベーションコードを入力してください');
      return;
    }

    this.isLoading.set(true);
    this.error.set(null);

    const result = await this.adminService.activateAdmin(this.activationCode);

    this.isLoading.set(false);

    if (result.success) {
      this.router.navigate(['/admin/challenges']);
    } else {
      this.error.set(result.error || 'アクティベーションに失敗しました');
    }
  }

  goBack(): void {
    this.router.navigate(['/challenges']);
  }
}
