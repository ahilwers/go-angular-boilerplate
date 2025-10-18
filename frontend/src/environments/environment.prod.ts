export const environment = {
  production: true,
  apiUrl: '/api/v1',
  auth: {
    enabled: true,
    issuer: 'https://auth.example.com/realms/boilerplate',
    clientId: 'boilerplate-client',
    redirectUri: 'https://app.example.com/auth/callback',
    scope: 'openid profile email'
  }
};
