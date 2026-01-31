import { Routes } from '@angular/router';
import { adminGuard, authGuard, guestGuard } from './guards/auth.guard';
import { AdminActivateComponent } from './pages/admin/admin-activate/admin-activate';
import { AdminChallengesComponent } from './pages/admin/admin-challenges/admin-challenges';
import { ChallengeFormComponent } from './pages/admin/challenge-form/challenge-form';
import { ChallengesComponent } from './pages/challenges/challenges';
import { LoginComponent } from './pages/login/login';
import { RegisterComponent } from './pages/register/register';

export const routes: Routes = [
  { path: 'login', component: LoginComponent, canActivate: [guestGuard] },
  { path: 'register', component: RegisterComponent, canActivate: [guestGuard] },
  { path: 'challenges', component: ChallengesComponent, canActivate: [authGuard] },
  { path: 'admin/activate', component: AdminActivateComponent, canActivate: [authGuard] },
  { path: 'admin/challenges', component: AdminChallengesComponent, canActivate: [adminGuard] },
  {
    path: 'admin/challenges/create',
    component: ChallengeFormComponent,
    canActivate: [adminGuard],
  },
  {
    path: 'admin/challenges/edit/:challengeId',
    component: ChallengeFormComponent,
    canActivate: [adminGuard],
  },
  { path: '', redirectTo: '/challenges', pathMatch: 'full' },
];
