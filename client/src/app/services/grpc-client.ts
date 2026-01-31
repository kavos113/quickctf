import { Client, createClient, Interceptor } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { ClientChallengeService, UserAuthService } from '../../gen/api/server/v1/client_pb';
import { AdminAuthService, AdminService } from '../../gen/api/server/v1/admin_pb';

const AUTH_TOKEN_KEY = 'auth_token';

function getAuthToken(): string | null {
  if (typeof window === 'undefined') return null;
  return localStorage.getItem(AUTH_TOKEN_KEY);
}

const authInterceptor: Interceptor = (next) => async (req) => {
  const token = getAuthToken();
  if (token) {
    req.header.set('Authorization', `Bearer ${token}`);
  }
  return next(req);
};

const transport = createConnectTransport({
  baseUrl: '/api',
  interceptors: [authInterceptor],
});

export const userAuthClient: Client<typeof UserAuthService> = createClient(
  UserAuthService,
  transport,
);

export const challengeClient: Client<typeof ClientChallengeService> = createClient(
  ClientChallengeService,
  transport,
);

export const adminAuthClient: Client<typeof AdminAuthService> = createClient(
  AdminAuthService,
  transport,
);

export const adminClient: Client<typeof AdminService> = createClient(AdminService, transport);
