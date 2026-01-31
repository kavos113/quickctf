import { Component, inject, OnInit, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Challenge } from '../../../../gen/api/server/v1/model_pb';
import { AdminService } from '../../../services/admin.service';
import { ThemeService } from '../../../services/theme.service';

@Component({
  selector: 'app-admin-challenges',
  imports: [],
  templateUrl: './admin-challenges.html',
  styleUrl: './admin-challenges.css',
})
export class AdminChallengesComponent implements OnInit {
  readonly adminService = inject(AdminService);
  readonly themeService = inject(ThemeService);
  private readonly router = inject(Router);

  readonly selectedChallenge = signal<Challenge | null>(null);

  ngOnInit(): void {
    this.adminService.loadChallenges();
  }

  createChallenge(): void {
    this.router.navigate(['/admin/challenges/create']);
  }

  editChallenge(challenge: Challenge): void {
    this.router.navigate(['/admin/challenges/edit'], {
      queryParams: {
        name: challenge.name,
        description: challenge.description,
        flag: challenge.flag,
        points: challenge.points,
        genre: challenge.genre,
      },
    });
  }

  async deleteChallenge(challenge: Challenge): Promise<void> {
    if (!confirm(`「${challenge.name}」を削除してもよろしいですか？`)) {
      return;
    }

    const result = await this.adminService.deleteChallenge(challenge.name);
    if (result.success) {
      await this.adminService.loadChallenges();
    } else {
      alert(result.error || '削除に失敗しました');
    }
  }

  goBack(): void {
    this.router.navigate(['/challenges']);
  }

  deactivateAdmin(): void {
    this.adminService.deactivateAdmin();
    this.router.navigate(['/challenges']);
  }
}
