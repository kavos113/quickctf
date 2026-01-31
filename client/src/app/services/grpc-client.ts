import { createPromiseClient, type PromiseClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { UserAuthService } from '../../gen/api/server/v1/client_connect';

const transport = createConnectTransport({
  baseUrl: '/api',
});

export const userAuthClient: PromiseClient<typeof UserAuthService> = createPromiseClient(
  UserAuthService,
  transport,
);
