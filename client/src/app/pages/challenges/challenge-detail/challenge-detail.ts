import { Component, EventEmitter, inject, Input, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Challenge } from '../../../../gen/api/server/v1/model_pb';
import { ChallengeService } from '../../../services/challenge.service';

@Component({
  selector: 'app-challenge-detail',
  imports: [FormsModule],
  templateUrl: './challenge-detail.html',
  styleUrl: './challenge-detail.css',
})
export class ChallengeDetailComponent {
  private readonly challengeService = inject(ChallengeService);

  @Input({ required: true }) challenge!: Challenge;
  @Output() closeModal = new EventEmitter<void>();

  flag = signal('');
  submitResult = signal<{
    type: 'success' | 'error' | 'wrong';
    message: string;
  } | null>(null);
  isSubmitting = signal(false);
  instanceStatus = signal<string | null>(null);
  isInstanceLoading = signal(false);

  async submitFlag(): Promise<void> {
    if (!this.flag() || this.isSubmitting()) return;

    this.isSubmitting.set(true);
    this.submitResult.set(null);

    const result = await this.challengeService.submitFlag(this.challenge.challengeId, this.flag());

    this.isSubmitting.set(false);

    if (!result.success) {
      this.submitResult.set({ type: 'error', message: result.error || 'エラーが発生しました' });
      return;
    }

    if (result.correct) {
      this.submitResult.set({
        type: 'success',
        message: `正解！ ${result.points} ポイント獲得`,
      });
    } else {
      this.submitResult.set({ type: 'wrong', message: '不正解です' });
    }
  }

  async startInstance(): Promise<void> {
    this.isInstanceLoading.set(true);
    const result = await this.challengeService.startInstance(this.challenge.challengeId);
    this.isInstanceLoading.set(false);

    if (result.success) {
      await this.checkInstanceStatus();
    }
  }

  async stopInstance(): Promise<void> {
    this.isInstanceLoading.set(true);
    const result = await this.challengeService.stopInstance(this.challenge.challengeId);
    this.isInstanceLoading.set(false);

    if (result.success) {
      this.instanceStatus.set(null);
    }
  }

  async checkInstanceStatus(): Promise<void> {
    const result = await this.challengeService.getInstanceStatus(this.challenge.challengeId);
    if (result.success) {
      this.instanceStatus.set(result.status || null);
    }
  }

  updateFlag(event: Event): void {
    const input = event.target as HTMLInputElement;
    this.flag.set(input.value);
  }

  onOverlayClick(event: MouseEvent): void {
    if ((event.target as HTMLElement).classList.contains('modal-overlay')) {
      this.closeModal.emit();
    }
  }

  onClose(): void {
    this.closeModal.emit();
  }
}
