import { Component, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { BuildLogSummary } from '../../../../gen/api/server/v1/admin_pb';
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

  challenge = signal<Challenge | null>(null);
  buildLogs = signal<BuildLogSummary[]>([]);
  selectedLog = signal<{ jobId: string; content: string; status: string } | null>(null);
  isLoading = signal(true);
  isLoadingLogs = signal(false);
  isLoadingLogContent = signal(false);
  error = signal<string | null>(null);
  buildLogsExpanded = signal(false);
  isPolling = signal(false);

  private pollingInterval: ReturnType<typeof setInterval> | null = null;
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
    this.stopPolling();
  }

  private async loadChallenge(challengeId: string): Promise<void> {
    this.isLoading.set(true);
    this.error.set(null);

    try {
      const result = await this.adminService.getChallenge(challengeId);
      if (result.success && result.challenge) {
        this.challenge.set(result.challenge);
        await this.loadBuildLogs(challengeId);

        // jobIdクエリパラメータがある場合は該当のログを開いてポーリング開始
        if (this.targetJobId) {
          this.buildLogsExpanded.set(true);
          await this.viewBuildLog(this.targetJobId);
          this.startPolling(this.targetJobId);
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
      this.stopPolling();
      return;
    }

    this.isLoadingLogContent.set(true);
    try {
      const result = await this.adminService.getBuildLog(jobId);
      if (result.success) {
        this.selectedLog.set({
          jobId,
          content: result.logContent || '',
          status: result.status || '',
        });

        // ステータスがpendingまたはbuildingの場合はポーリング開始
        if (result.status === 'pending' || result.status === 'building') {
          this.startPolling(jobId);
        } else {
          this.stopPolling();
        }
      }
    } catch (error) {
      console.error('Failed to load build log:', error);
    } finally {
      this.isLoadingLogContent.set(false);
    }
  }

  private startPolling(jobId: string): void {
    this.stopPolling();
    this.isPolling.set(true);

    this.pollingInterval = setInterval(async () => {
      try {
        const result = await this.adminService.getBuildLog(jobId);
        if (result.success) {
          this.selectedLog.set({
            jobId,
            content: result.logContent || '',
            status: result.status || '',
          });

          // ビルドログ一覧も更新
          const challenge = this.challenge();
          if (challenge) {
            await this.loadBuildLogs(challenge.challengeId);
          }

          // ステータスがsuccessまたはfailedになったらポーリング停止
          if (result.status === 'success' || result.status === 'failed') {
            this.stopPolling();
          }
        }
      } catch (error) {
        console.error('Polling error:', error);
      }
    }, 2000);
  }

  private stopPolling(): void {
    if (this.pollingInterval) {
      clearInterval(this.pollingInterval);
      this.pollingInterval = null;
    }
    this.isPolling.set(false);
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

  getStatusClass(status: string): string {
    switch (status) {
      case 'success':
        return 'status-success';
      case 'failed':
        return 'status-failed';
      case 'building':
        return 'status-building';
      case 'pending':
        return 'status-pending';
      default:
        return '';
    }
  }

  getStatusLabel(status: string): string {
    switch (status) {
      case 'success':
        return '成功';
      case 'failed':
        return '失敗';
      case 'building':
        return 'ビルド中';
      case 'pending':
        return '待機中';
      default:
        return status;
    }
  }
}
