import { Routes } from '@angular/router';
import { authGuard, guestGuard } from './guards/auth.guard';
import { ChallengesComponent } from './pages/challenges/challenges';
import { LoginComponent } from './pages/login/login';
import { RegisterComponent } from './pages/register/register';

export const routes: Routes = [
  { path: 'login', component: LoginComponent, canActivate: [guestGuard] },
  { path: 'register', component: RegisterComponent, canActivate: [guestGuard] },
  { path: 'challenges', component: ChallengesComponent, canActivate: [authGuard] },
  { path: '', redirectTo: '/challenges', pathMatch: 'full' },
];
