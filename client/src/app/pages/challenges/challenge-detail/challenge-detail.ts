import { Component, EventEmitter, inject, Input, OnInit, Output, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { Challenge } from '../../../../gen/api/server/v1/model_pb';
import { ChallengeService } from '../../../services/challenge.service';
import { GetInstanceStatusResponse_Status } from '../../../../gen/api/server/v1/client_pb';

interface InstanceConnectionInfo {
  host: string;
  port: number;
}

@Component({
  selector: 'app-challenge-detail',
  imports: [FormsModule],
  templateUrl: './challenge-detail.html',
  styleUrl: './challenge-detail.css',
})
export class ChallengeDetailComponent implements OnInit {
  private readonly challengeService = inject(ChallengeService);

  @Input({ required: true }) challenge!: Challenge;
  @Output() closeModal = new EventEmitter<void>();

  flag = signal('');
  submitResult = signal<{
    type: 'success' | 'error' | 'wrong';
    message: string;
  } | null>(null);
  isSubmitting = signal(false);
  instanceStatus = signal<GetInstanceStatusResponse_Status | null>(null);
  isInstanceLoading = signal(false);
  instanceConnectionInfo = signal<InstanceConnectionInfo | null>(null);

  async ngOnInit(): Promise<void> {
    await this.checkInstanceStatus();
  }

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
      this.instanceConnectionInfo.set({ host: result.host || '', port: result.port || 0 });
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
      if (result.status && result.host && result.port) {
        this.instanceConnectionInfo.set({ host: result.host, port: result.port });
      } else {
        this.instanceConnectionInfo.set(null);
      }
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

  isRunning(): boolean {
    return this.instanceStatus() === GetInstanceStatusResponse_Status.RUNNING;
  }
}
