import {
  Component,
  ElementRef,
  inject,
  OnDestroy,
  OnInit,
  signal,
  ViewChild,
} from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { BuildLogSummary, BuildStatus } from '../../../../gen/api/server/v1/admin_pb';
import { Challenge } from '../../../../gen/api/server/v1/model_pb';
import { AdminService } from '../../../services/admin.service';
import { ThemeService } from '../../../services/theme.service';

@Component({
  selector: 'app-challenge-detail',
  imports: [],
  templateUrl: './challenge-detail.html',
  styleUrl: './challenge-detail.css',
})
export class ChallengeDetailComponent implements OnInit, OnDestroy {
  private readonly router = inject(Router);
  private readonly route = inject(ActivatedRoute);
  private readonly adminService = inject(AdminService);
  readonly themeService = inject(ThemeService);

  @ViewChild('logOutput') logOutputRef?: ElementRef<HTMLPreElement>;

  challenge = signal<Challenge | null>(null);
  buildLogs = signal<BuildLogSummary[]>([]);
  selectedLog = signal<{ jobId: string; content: string; status: BuildStatus } | null>(null);
  isLoading = signal(true);
  isLoadingLogs = signal(false);
  isLoadingLogContent = signal(false);
  error = signal<string | null>(null);
  buildLogsExpanded = signal(false);
  isStreaming = signal(false);

  private abortController: AbortController | null = null;
  private targetJobId: string | null = null;

  ngOnInit(): void {
    const challengeId = this.route.snapshot.paramMap.get('challengeId');
    this.targetJobId = this.route.snapshot.queryParamMap.get('jobId');

    if (challengeId) {
      this.loadChallenge(challengeId);
    } else {
      this.error.set('Challenge ID が指定されていません');
      this.isLoading.set(false);
    }
  }

  ngOnDestroy(): void {
    this.stopStreaming();
  }

  private scrollToBottom(): void {
    if (this.logOutputRef?.nativeElement) {
      const element = this.logOutputRef.nativeElement;
      element.scrollTop = element.scrollHeight;
    }
  }

  private async loadChallenge(challengeId: string): Promise<void> {
    this.isLoading.set(true);
    this.error.set(null);

    try {
      const result = await this.adminService.getChallenge(challengeId);
      if (result.success && result.challenge) {
        this.challenge.set(result.challenge);
        await this.loadBuildLogs(challengeId);

        if (this.targetJobId) {
          this.buildLogsExpanded.set(true);
          await this.viewBuildLog(this.targetJobId);
        }
      } else {
        this.error.set(result.error || '問題が見つかりません');
      }
    } catch (err) {
      console.error('Failed to load challenge:', err);
      this.error.set('問題の読み込みに失敗しました');
    } finally {
      this.isLoading.set(false);
    }
  }

  async loadBuildLogs(challengeId: string): Promise<void> {
    this.isLoadingLogs.set(true);
    try {
      const result = await this.adminService.listBuildLogs(challengeId);
      if (result.success && result.logs) {
        this.buildLogs.set(result.logs);
      }
    } catch (error) {
      console.error('Failed to load build logs:', error);
    } finally {
      this.isLoadingLogs.set(false);
    }
  }

  toggleBuildLogs(): void {
    this.buildLogsExpanded.update((v) => !v);
  }

  async viewBuildLog(jobId: string): Promise<void> {
    if (this.selectedLog()?.jobId === jobId) {
      this.selectedLog.set(null);
      this.stopStreaming();
      return;
    }

    this.stopStreaming();
    this.isLoadingLogContent.set(true);

    const logs = this.buildLogs();
    const logInfo = logs.find((l) => l.jobId === jobId);
    const status = logInfo?.status ?? BuildStatus.UNSPECIFIED;

    if (status === BuildStatus.PENDING || status === BuildStatus.BUILDING) {
      this.selectedLog.set({
        jobId,
        content: '',
        status: status,
      });
      this.isLoadingLogContent.set(false);
      await this.startStreaming(jobId);
    } else {
      try {
        const result = await this.adminService.getBuildLog(jobId);
        if (result.success) {
          this.selectedLog.set({
            jobId,
            content: result.logContent || '',
            status: result.status ?? BuildStatus.UNSPECIFIED,
          });
        }
      } catch (error) {
        console.error('Failed to load build log:', error);
      } finally {
        this.isLoadingLogContent.set(false);
      }
    }
  }

  private async startStreaming(jobId: string): Promise<void> {
    this.stopStreaming();
    this.isStreaming.set(true);
    this.isLoading.set(false);
    this.abortController = new AbortController();

    try {
      for await (const data of this.adminService.streamBuildLog(jobId)) {
        console.log('Received log data:', data);

        if (this.abortController?.signal.aborted) break;

        if (data.logLine) {
          this.selectedLog.update((current) => {
            if (!current) return current;
            return {
              ...current,
              content: current.content + data.logLine,
              status: data.status,
            };
          });
          this.scrollToBottom();
        }

        if (data.isComplete) {
          this.selectedLog.update((current) => {
            if (!current) return current;
            return {
              ...current,
              status: data.status,
            };
          });

          const challenge = this.challenge();
          if (challenge) {
            await this.loadBuildLogs(challenge.challengeId);
          }

          this.stopStreaming();
          break;
        }
      }
    } catch (error) {
      if (this.abortController?.signal.aborted) return;
      console.error('Streaming error:', error);
    } finally {
      this.isStreaming.set(false);
    }
  }

  private stopStreaming(): void {
    if (this.abortController) {
      this.abortController.abort();
      this.abortController = null;
    }
    this.isStreaming.set(false);
  }

  editChallenge(): void {
    const challenge = this.challenge();
    if (challenge) {
      this.router.navigate(['/admin/challenges/edit', challenge.challengeId]);
    }
  }

  async deleteChallenge(): Promise<void> {
    const challenge = this.challenge();
    if (!challenge) return;

    if (!confirm(`「${challenge.name}」を削除してもよろしいですか？`)) {
      return;
    }

    const result = await this.adminService.deleteChallenge(challenge.challengeId);
    if (result.success) {
      this.router.navigate(['/admin/challenges']);
    } else {
      alert(result.error || '削除に失敗しました');
    }
  }

  goBack(): void {
    this.router.navigate(['/admin/challenges']);
  }

  getStatusClass(status: BuildStatus): string {
    switch (status) {
      case BuildStatus.SUCCESS:
        return 'status-success';
      case BuildStatus.FAILED:
        return 'status-failed';
      case BuildStatus.BUILDING:
        return 'status-building';
      case BuildStatus.PENDING:
        return 'status-pending';
      default:
        return '';
    }
  }

  getStatusLabel(status: BuildStatus): string {
    switch (status) {
      case BuildStatus.SUCCESS:
        return '成功';
      case BuildStatus.FAILED:
        return '失敗';
      case BuildStatus.BUILDING:
        return 'ビルド中';
      case BuildStatus.PENDING:
        return '待機中';
      default:
        return '不明';
    }
  }
}
