import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { AdminService } from '../../../services/admin.service';

@Component({
  selector: 'app-challenge-form',
  imports: [FormsModule],
  templateUrl: './challenge-form.html',
  styleUrl: './challenge-form.css',
})
export class ChallengeFormComponent implements OnInit {
  private readonly adminService = inject(AdminService);
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);

  readonly isEditMode = signal(false);
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);
  readonly selectedFile = signal<File | null>(null);
  readonly uploadProgress = signal(0);

  name = '';
  description = '';
  flag = '';
  points = 100;
  genre = '';

  private originalName = '';

  ngOnInit(): void {
    this.route.queryParams.subscribe((params) => {
      if (params['name']) {
        this.isEditMode.set(true);
        this.originalName = params['name'];
        this.name = params['name'] || '';
        this.description = params['description'] || '';
        this.flag = params['flag'] || '';
        this.points = Number(params['points']) || 100;
        this.genre = params['genre'] || '';
      }
    });
  }

  async onSubmit(): Promise<void> {
    if (!this.validateForm()) {
      return;
    }

    this.isLoading.set(true);
    this.error.set(null);

    const challengeData = {
      name: this.name,
      description: this.description,
      flag: this.flag,
      points: this.points,
      genre: this.genre,
    };

    let challengeId: string | undefined;
    let result: { success: boolean; challengeId?: string; error?: string };

    if (this.isEditMode()) {
      const updateResult = await this.adminService.updateChallenge(
        this.originalName,
        challengeData,
      );
      result = updateResult;
      challengeId = this.originalName; // 編集モードでは元の名前をIDとして使用
    } else {
      const createResult = await this.adminService.createChallenge(challengeData);
      result = createResult;
      challengeId = createResult.challengeId;
    }

    if (result.success && challengeId && this.selectedFile()) {
      // イメージファイルをアップロード
      this.uploadProgress.set(10);
      const uploadResult = await this.adminService.uploadChallengeImage(
        challengeId,
        this.selectedFile()!,
      );
      this.uploadProgress.set(100);

      if (!uploadResult.success) {
        this.isLoading.set(false);
        this.error.set(uploadResult.error || 'イメージのアップロードに失敗しました');
        return;
      }
    }

    this.isLoading.set(false);

    if (result.success) {
      this.router.navigate(['/admin/challenges']);
    } else {
      this.error.set(result.error || '保存に失敗しました');
    }
  }

  private validateForm(): boolean {
    if (!this.name.trim()) {
      this.error.set('問題名を入力してください');
      return false;
    }
    if (!this.description.trim()) {
      this.error.set('説明を入力してください');
      return false;
    }
    if (!this.flag.trim()) {
      this.error.set('フラグを入力してください');
      return false;
    }
    if (this.points <= 0) {
      this.error.set('ポイントは1以上を指定してください');
      return false;
    }
    if (!this.genre.trim()) {
      this.error.set('ジャンルを入力してください');
      return false;
    }
    return true;
  }

  onFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      const file = input.files[0];
      if (this.isValidTarFile(file)) {
        this.selectedFile.set(file);
        this.error.set(null);
      } else {
        this.error.set('tar, tar.gz, または tgz ファイルを選択してください');
      }
    }
  }

  private isValidTarFile(file: File): boolean {
    const validExtensions = ['.tar', '.tar.gz', '.tgz'];
    const fileName = file.name.toLowerCase();
    return validExtensions.some((ext) => fileName.endsWith(ext));
  }

  removeFile(): void {
    this.selectedFile.set(null);
    this.uploadProgress.set(0);
    // ファイル入力をリセット
    const fileInput = document.getElementById('imageFile') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
  }

  formatFileSize(bytes: number): string {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  cancel(): void {
    this.router.navigate(['/admin/challenges']);
  }
}
