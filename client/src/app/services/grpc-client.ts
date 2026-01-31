import { Client, createClient } from '@connectrpc/connect';
import { createConnectTransport } from '@connectrpc/connect-web';
import { UserAuthService } from '../../gen/api/server/v1/client_pb';

const transport = createConnectTransport({
  baseUrl: '/api',
});

export const userAuthClient: Client<typeof UserAuthService> = createClient(
  UserAuthService,
  transport
) 