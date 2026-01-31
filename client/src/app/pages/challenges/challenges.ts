import { Component, inject, OnInit, signal } from '@angular/core';
import { Router } from '@angular/router';
import { Challenge } from '../../../gen/api/server/v1/model_pb';
import { AdminService } from '../../services/admin.service';
import { AuthService } from '../../services/auth.service';
import { ChallengeService } from '../../services/challenge.service';
import { ChallengeDetailComponent } from './challenge-detail/challenge-detail';

@Component({
  selector: 'app-challenges',
  imports: [ChallengeDetailComponent],
  templateUrl: './challenges.html',
  styleUrl: './challenges.css',
})
export class ChallengesComponent implements OnInit {
  private readonly authService = inject(AuthService);
  readonly adminService = inject(AdminService);
  readonly challengeService = inject(ChallengeService);
  private readonly router = inject(Router);

  readonly challenges = this.challengeService.challenges;
  readonly isLoading = this.challengeService.isLoading;
  readonly error = this.challengeService.error;
  readonly username = signal('');

  selectedChallenge = signal<Challenge | null>(null);
  selectedGenre = signal<string | null>(null);

  ngOnInit(): void {
    if (!this.authService.isAuthenticated()) {
      this.router.navigate(['/login']);
      return;
    }

    this.username.set(this.authService.authState().username || '');
    this.challengeService.loadChallenges();
  }

  get genres(): string[] {
    const genreSet = new Set(this.challenges().map((c) => c.genre));
    return Array.from(genreSet).sort();
  }

  get filteredChallenges(): Challenge[] {
    const genre = this.selectedGenre();
    if (!genre) {
      return this.challenges();
    }
    return this.challenges().filter((c) => c.genre === genre);
  }

  selectGenre(genre: string | null): void {
    this.selectedGenre.set(genre);
  }

  openChallenge(challenge: Challenge): void {
    this.selectedChallenge.set(challenge);
  }

  closeChallenge(): void {
    this.selectedChallenge.set(null);
  }

  goToAdmin(): void {
    if (this.adminService.isAdmin()) {
      this.router.navigate(['/admin/challenges']);
    } else {
      this.router.navigate(['/admin/activate']);
    }
  }

  async logout(): Promise<void> {
    await this.authService.logout();
    this.router.navigate(['/login']);
  }
}
