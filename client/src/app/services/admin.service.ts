import { Injectable, signal } from '@angular/core';
import { create } from '@bufbuild/protobuf';
import {
  CreateChallengeRequestSchema,
  DeleteChallengeRequestSchema,
  ListChallengesRequestSchema,
  UpdateChallengeRequestSchema,
} from '../../gen/api/server/v1/admin_pb';
import { Challenge, ChallengeSchema } from '../../gen/api/server/v1/model_pb';
import { adminAuthClient, adminClient } from './grpc-client';

const ADMIN_KEY = 'is_admin';

@Injectable({
  providedIn: 'root',
})
export class AdminService {
  readonly isAdmin = signal(false);
  readonly challenges = signal<Challenge[]>([]);
  readonly isLoading = signal(false);
  readonly error = signal<string | null>(null);

  constructor() {
    this.loadAdminStatus();
  }

  private loadAdminStatus(): void {
    if (typeof window === 'undefined') return;
    const isAdmin = localStorage.getItem(ADMIN_KEY) === 'true';
    this.isAdmin.set(isAdmin);
  }

  private saveAdminStatus(isAdmin: boolean): void {
    if (isAdmin) {
      localStorage.setItem(ADMIN_KEY, 'true');
    } else {
      localStorage.removeItem(ADMIN_KEY);
    }
    this.isAdmin.set(isAdmin);
  }

  async activateAdmin(activationCode: string): Promise<{ success: boolean; error?: string }> {
    try {
      await adminAuthClient.adminLogin({ password: activationCode });
      this.saveAdminStatus(true);
      return { success: true };
    } catch (error) {
      console.error('Admin activation error:', error);
      return { success: false, error: 'アクティベーションに失敗しました' };
    }
  }

  async deactivateAdmin(): Promise<void> {
    try {
      await adminAuthClient.adminLogout({});
    } catch (error) {
      console.error('Admin logout error:', error);
    } finally {
      this.saveAdminStatus(false);
    }
  }

  async loadChallenges(): Promise<void> {
    this.isLoading.set(true);
    this.error.set(null);

    try {
      const request = create(ListChallengesRequestSchema, {});
      const response = await adminClient.listChallenges(request);

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

  async createChallenge(challenge: {
    name: string;
    description: string;
    flag: string;
    points: number;
    genre: string;
  }): Promise<{ success: boolean; challengeId?: string; error?: string }> {
    try {
      const challengeMsg = create(ChallengeSchema, {
        name: challenge.name,
        description: challenge.description,
        flag: challenge.flag,
        points: challenge.points,
        genre: challenge.genre,
      });

      const request = create(CreateChallengeRequestSchema, { challenge: challengeMsg });
      const response = await adminClient.createChallenge(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true, challengeId: response.challengeId };
    } catch (err) {
      console.error('Failed to create challenge:', err);
      return { success: false, error: '問題の作成に失敗しました' };
    }
  }

  async updateChallenge(
    challengeId: string,
    challenge: {
      name: string;
      description: string;
      flag: string;
      points: number;
      genre: string;
    },
  ): Promise<{ success: boolean; error?: string }> {
    try {
      const challengeMsg = create(ChallengeSchema, {
        name: challenge.name,
        description: challenge.description,
        flag: challenge.flag,
        points: challenge.points,
        genre: challenge.genre,
      });

      const request = create(UpdateChallengeRequestSchema, {
        challengeId,
        challenge: challengeMsg,
      });
      const response = await adminClient.updateChallenge(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true };
    } catch (err) {
      console.error('Failed to update challenge:', err);
      return { success: false, error: '問題の更新に失敗しました' };
    }
  }

  async deleteChallenge(challengeId: string): Promise<{ success: boolean; error?: string }> {
    try {
      const request = create(DeleteChallengeRequestSchema, { challengeId });
      const response = await adminClient.deleteChallenge(request);

      if (response.errorMessage) {
        return { success: false, error: response.errorMessage };
      }

      return { success: true };
    } catch (err) {
      console.error('Failed to delete challenge:', err);
      return { success: false, error: '問題の削除に失敗しました' };
    }
  }
}
