import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { Attachment } from '../../../../gen/api/server/v1/model_pb';
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
  readonly attachments = signal<Attachment[]>([]);
  readonly selectedAttachmentFile = signal<File | null>(null);
  readonly isUploadingAttachment = signal(false);

  challengeId = '';
  name = '';
  description = '';
  flag = '';
  points = 100;
  genre = '';
  requiresInstance = false;

  ngOnInit(): void {
    this.route.params.subscribe((params) => {
      const challengeId = params['challengeId'];
      if (challengeId) {
        this.isEditMode.set(true);
        this.challengeId = challengeId;
        this.loadChallengeData(challengeId);
      }
    });
  }

  private loadChallengeData(challengeId: string): void {
    const challenges = this.adminService.challenges();
    const challenge = challenges.find((c) => c.challengeId === challengeId);

    if (challenge) {
      this.name = challenge.name;
      this.description = challenge.description;
      this.flag = challenge.flag;
      this.points = challenge.points;
      this.genre = challenge.genre;
      this.requiresInstance = challenge.requiresInstance;
      this.attachments.set([...challenge.attachments]);
    } else {
      this.adminService.loadChallenges().then(() => {
        const reloadedChallenges = this.adminService.challenges();
        const foundChallenge = reloadedChallenges.find((c) => c.challengeId === challengeId);
        if (foundChallenge) {
          this.name = foundChallenge.name;
          this.description = foundChallenge.description;
          this.flag = foundChallenge.flag;
          this.points = foundChallenge.points;
          this.genre = foundChallenge.genre;
          this.requiresInstance = foundChallenge.requiresInstance;
          this.attachments.set([...foundChallenge.attachments]);
        } else {
          this.error.set('問題が見つかりません');
        }
      });
    }
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
      requiresInstance: this.requiresInstance,
    };

    let challengeId: string | undefined;
    let result: { success: boolean; challengeId?: string; error?: string };

    if (this.isEditMode()) {
      const updateResult = await this.adminService.updateChallenge(this.challengeId, challengeData);
      result = updateResult;
      challengeId = this.challengeId;
    } else {
      const createResult = await this.adminService.createChallenge(challengeData);
      result = createResult;
      challengeId = createResult.challengeId;
    }

    if (result.success && challengeId && this.selectedFile()) {
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

      this.isLoading.set(false);
      // tarファイルをアップロードした場合は問題詳細ページに遷移してビルドログを表示
      this.router.navigate(['/admin/challenges', challengeId], {
        queryParams: { jobId: uploadResult.jobId },
      });
      return;
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
    const fileInput = document.getElementById('imageFile') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
  }

  formatFileSize(bytes: number | bigint): string {
    const numBytes = typeof bytes === 'bigint' ? Number(bytes) : bytes;
    if (numBytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(numBytes) / Math.log(k));
    return parseFloat((numBytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i];
  }

  cancel(): void {
    this.router.navigate(['/admin/challenges']);
  }

  onAttachmentFileSelected(event: Event): void {
    const input = event.target as HTMLInputElement;
    if (input.files && input.files.length > 0) {
      this.selectedAttachmentFile.set(input.files[0]);
      this.error.set(null);
    }
  }

  async uploadAttachment(): Promise<void> {
    const file = this.selectedAttachmentFile();
    if (!file) {
      this.error.set('ファイルを選択してください');
      return;
    }

    if (!this.challengeId) {
      this.error.set('問題を先に保存してください');
      return;
    }

    this.isUploadingAttachment.set(true);
    this.error.set(null);

    const result = await this.adminService.uploadAttachment(this.challengeId, file);

    this.isUploadingAttachment.set(false);

    if (result.success && result.attachment) {
      this.attachments.update((current) => [...current, result.attachment!]);
      this.selectedAttachmentFile.set(null);
      const fileInput = document.getElementById('attachmentFile') as HTMLInputElement;
      if (fileInput) {
        fileInput.value = '';
      }
    } else {
      this.error.set(result.error || '添付ファイルのアップロードに失敗しました');
    }
  }

  async deleteAttachment(attachmentId: string): Promise<void> {
    if (!confirm('この添付ファイルを削除しますか？')) {
      return;
    }

    this.isLoading.set(true);
    this.error.set(null);

    const result = await this.adminService.deleteAttachment(this.challengeId, attachmentId);

    this.isLoading.set(false);

    if (result.success) {
      this.attachments.update((current) => current.filter((a) => a.attachmentId !== attachmentId));
    } else {
      this.error.set(result.error || '添付ファイルの削除に失敗しました');
    }
  }

  removeAttachmentFile(): void {
    this.selectedAttachmentFile.set(null);
    const fileInput = document.getElementById('attachmentFile') as HTMLInputElement;
    if (fileInput) {
      fileInput.value = '';
    }
  }
}
