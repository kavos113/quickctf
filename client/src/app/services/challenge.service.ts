import { inject, Injectable, signal } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import {
  GetChallengesRequestSchema,
  GetInstanceStatusRequestSchema,
  GetInstanceStatusResponse_Status,
  StartInstanceRequestSchema,
  StopInstanceRequestSchema,
  SubmitFlagRequestSchema,
} from '../../gen/api/server/v1/client_pb';
import { Challenge, SubmissionSchema } from '../../gen/api/server/v1/model_pb';
import { challengeClient } from './grpc-client';
import { AuthService } from './auth.service';

@Injectable({
  providedIn: 'root',
})
export class ChallengeService {
  private readonly authService = inject(AuthService);

  readonly challenges = signal<Challenge[]>([]);
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);

  async loadChallenges(): Promise<void> {
    this.isLoading.set(true);
    this.error.set(null);

    try {
      const request = create(GetChallengesRequestSchema, {});
      const response = await challengeClient.getChallenges(request);

      if (response.errorMessage) {
        this.error.set(response.errorMessage);
        return;
      }

      this.challenges.set(response.challenges);
    } catch (err) {
      console.error('Failed to load challenges:', err);
      this.error.set('問題の読み込みに失敗しました');
    } finally {
      this.isLoading.set(false);
    }
  }

  async submitFlag(
    challengeId: string,
    flag: string,
  ): Promise<{ success: boolean; correct?: boolean; points?: number; error?: string }> {
    try {
      const userId = this.authService.authState().userId || '';
      const submission = create(SubmissionSchema, {
        challengeId,
        userId,
        submittedFlag: flag,
        timestamp: BigInt(Date.now()),
      });

      const request = create(SubmitFlagRequestSchema, { submission });
      const response = await challengeClient.submitFlag(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return {
        success: true,
        correct: response.correct,
        points: response.pointsAwarded,
      };
    } catch (err) {
      console.error('Failed to submit flag:', err);
      return { success: false, error: 'フラグの送信に失敗しました' };
    }
  }

  async startInstance(challengeId: string): Promise<{
    success: boolean;
    host?: string;
    port?: number;
    error?: string;
  }> {
    try {
      const request = create(StartInstanceRequestSchema, { challengeId });
      const response = await challengeClient.startInstance(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true, host: response.host, port: response.port };
    } catch (err) {
      console.error('Failed to start instance:', err);
      return { success: false, error: 'インスタンスの起動に失敗しました' };
    }
  }

  async stopInstance(challengeId: string): Promise<{ success: boolean; error?: string }> {
    try {
      const request = create(StopInstanceRequestSchema, { challengeId });
      const response = await challengeClient.stopInstance(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true };
    } catch (err) {
      console.error('Failed to stop instance:', err);
      return { success: false, error: 'インスタンスの停止に失敗しました' };
    }
  }

  async getInstanceStatus(
    challengeId: string,
  ): Promise<{ success: boolean; status?: GetInstanceStatusResponse_Status; host?: string; port?: number; error?: string }> {
    try {
      const request = create(GetInstanceStatusRequestSchema, { challengeId });
      const response = await challengeClient.getInstanceStatus(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true, status: response.status, host: response.host, port: response.port };
    } catch (err) {
      console.error('Failed to get instance status:', err);
      return { success: false, error: 'ステータスの取得に失敗しました' };
    }
  }
}
