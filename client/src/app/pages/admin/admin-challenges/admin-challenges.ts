import { Component, inject, OnInit } from '@angular/core';
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

  ngOnInit(): void {
    this.adminService.loadChallenges();
  }

  createChallenge(): void {
    this.router.navigate(['/admin/challenges/create']);
  }

  openDetail(challenge: Challenge): void {
    this.router.navigate(['/admin/challenges', challenge.challengeId]);
  }

  goBack(): void {
    this.router.navigate(['/challenges']);
  }

  deactivateAdmin(): void {
    this.adminService.deactivateAdmin();
    this.router.navigate(['/challenges']);
  }
}
