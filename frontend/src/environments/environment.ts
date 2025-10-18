export const environment = {
  production: false,
  apiUrl: 'http://localhost:8080/api/v1',
  auth: {
    enabled: true,
    issuer: 'http://localhost:8081/realms/boilerplate',
    clientId: 'boilerplate-client',
    redirectUri: 'http://localhost:4200/auth/callback',
    scope: 'openid profile email'
  }
};
