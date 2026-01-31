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

    let result;
    if (this.isEditMode()) {
      result = await this.adminService.updateChallenge(this.originalName, challengeData);
    } else {
      result = await this.adminService.createChallenge(challengeData);
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

  cancel(): void {
    this.router.navigate(['/admin/challenges']);
  }
}
