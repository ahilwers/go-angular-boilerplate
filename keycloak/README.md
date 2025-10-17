# Keycloak Setup and Integration

This guide provides step-by-step instructions for integrating Keycloak with the boilerplate application.

## Prerequisites

- Docker and Docker Compose installed
- Go backend running

## Step 1: Start Keycloak

```bash
# Start Keycloak and MongoDB
docker-compose up -d keycloak mongodb

# Follow logs (optional)
docker-compose logs -f keycloak
```

Keycloak is ready when you see this message:
```
Keycloak 23.0 - Started in Xms
```

## Step 2: Open Keycloak Admin Console

1. Open http://localhost:8081 in your browser
2. Click on "Administration Console"
3. Login:
   - **Username**: `admin`
   - **Password**: `admin`

## Step 3: Create Realm

A realm is an isolated environment for users, clients, and roles.

1. Click on the dropdown menu at the top left (showing "master")
2. Click on **"Create Realm"**
3. Enter:
   - **Realm name**: `boilerplate`
   - **Enabled**: ✓ (checked)
4. Click **"Create"**

## Step 4: Create Client

A client represents your application.

1. In the `boilerplate` realm, go to **Clients** → **Create client**

### Basic Settings:
- **Client type**: `OpenID Connect`
- **Client ID**: `boilerplate-client`
- Click **Next**

### Capability config:
- **Client authentication**: ✓ **ON** (important!)
- **Authorization**: ✗ OFF
- **Authentication flow**:
  - ✓ Standard flow
  - ✓ Direct access grants
  - ✗ Implicit flow (deprecated)
  - ✗ Service accounts roles
- Click **Next**

### Login settings:
- **Root URL**: `http://localhost:8080`
- **Valid redirect URIs**:
  ```
  http://localhost:8080/*
  http://localhost:4200/*
  ```
- **Valid post logout redirect URIs**: `http://localhost:8080/*`
- **Web origins**: `+` (means: all allowed redirect URIs)

4. Click **Save**

## Step 5: Get Client Secret

1. Go to **Clients** → **boilerplate-client**
2. Select the **Credentials** tab
3. **Copy** the **Client Secret** (you'll need it soon)

Example:
```
Client Secret: a1b2c3d4-e5f6-7890-abcd-ef1234567890
```

## Step 6: Create Test User

1. Go to **Users** → **Add user**
2. Enter:
   - **Username**: `testuser`
   - **Email**: `test@example.com`
   - **First name**: `Test`
   - **Last name**: `User`
   - **Email verified**: ✓ ON
3. Click **Create**

### Set password:
1. In the newly created user, select the **Credentials** tab
2. Click on **Set password**
3. Enter:
   - **Password**: `testpass123`
   - **Password confirmation**: `testpass123`
   - **Temporary**: ✗ **OFF** (important!)
4. Click **Save**

## Step 7: Configure Backend

### Option A: Via Configuration File

Edit `config/local.yaml`:

```yaml
auth:
  enabled: true  # IMPORTANT: enable!
  issuer: "http://localhost:8081/realms/boilerplate"
  client_id: "boilerplate-client"
  client_secret: "YOUR_CLIENT_SECRET_HERE"  # from Step 5
  jwks_url: "http://localhost:8081/realms/boilerplate/protocol/openid-connect/certs"
```

### Option B: Via Environment Variables

Create/edit `.env`:

```bash
AUTH_ENABLED=true
AUTH_ISSUER=http://localhost:8081/realms/boilerplate
AUTH_CLIENT_ID=boilerplate-client
AUTH_CLIENT_SECRET=a1b2c3d4-e5f6-7890-abcd-ef1234567890
AUTH_JWKS_URL=http://localhost:8081/realms/boilerplate/protocol/openid-connect/certs
```

Then start the backend:
```bash
source .env
cd backend && go run main.go
```

## Step 8: Testing

### 1. Get Access Token

```bash
curl -X POST http://localhost:8081/realms/boilerplate/protocol/openid-connect/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "client_id=boilerplate-client" \
  -d "client_secret=YOUR_CLIENT_SECRET" \
  -d "grant_type=password" \
  -d "username=testuser" \
  -d "password=testpass123"
```

Response (shortened):
```json
{
  "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI...",
  "expires_in": 300,
  "refresh_expires_in": 1800,
  "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI...",
  "token_type": "Bearer"
}
```

### 2. Call API with Token

```bash
# Store token in variable (replace with your token)
TOKEN="eyJhbGciOiJSUzI1NiIsInR5cCI..."

# Call API
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/v1/projects
```

✅ **Success**: You receive the project list
❌ **Error**: `Unauthorized` → Token invalid or expired

### 3. Test without Token (should fail)

```bash
curl http://localhost:8080/api/v1/projects
# Expected: 401 Unauthorized
```

## Troubleshooting

### Problem: "Failed to refresh JWKS"

**Cause**: Backend cannot reach Keycloak

**Solution**:
```bash
# Check if Keycloak is running
docker-compose ps keycloak

# Test JWKS URL manually
curl http://localhost:8081/realms/boilerplate/protocol/openid-connect/certs
```

### Problem: "Invalid issuer"

**Cause**: Issuer in config doesn't match the token

**Solution**: Check the issuer URL:
```bash
# Decode token (without verification)
echo "YOUR_TOKEN" | cut -d. -f2 | base64 -d 2>/dev/null | jq .

# Output should contain:
{
  "iss": "http://localhost:8081/realms/boilerplate",
  ...
}
```

### Problem: "Token expired"

**Cause**: Access token is only valid for 5 minutes (default)

**Solution**: Get new token (see Step 8.1)

**Or**: Increase token lifetime:
1. Keycloak Admin → **Realm Settings** → **Tokens**
2. **Access Token Lifespan**: from `5 Minutes` to e.g. `30 Minutes`
3. **Save**

### Problem: "Client secret incorrect"

**Solution**:
1. Keycloak → **Clients** → **boilerplate-client** → **Credentials**
2. Click on **Regenerate** for new secret
3. Copy the new secret into your config

## Advanced Configuration

### Adding Roles

1. **Realm Roles** → **Create role**
2. **Role name**: e.g. `admin`, `user`
3. **Save**

4. Assign to user:
   - **Users** → `testuser` → **Role mapping**
   - **Assign role** → Select role

### Using Roles in Backend

```go
import "boilerplate/internal/auth"

func (h *Handler) AdminOnly(w http.ResponseWriter, r *http.Request) {
    claims, ok := auth.GetUserClaims(r.Context())
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Check if user has admin role
    hasAdminRole := false
    for _, role := range claims.Roles {
        if role == "admin" {
            hasAdminRole = true
            break
        }
    }

    if !hasAdminRole {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Perform admin operation
    w.Write([]byte("Admin access granted"))
}
```

### Creating Groups

1. **Groups** → **Create group**
2. **Name**: e.g. `developers`, `managers`
3. **Save**

4. Add user to group:
   - **Users** → `testuser` → **Groups**
   - **Join Group** → Select group

### Exporting Realm

For version control or other environments:

```bash
docker exec -it boilerplate-keycloak /opt/keycloak/bin/kc.sh export \
  --dir /tmp/export \
  --realm boilerplate \
  --users realm_file

docker cp boilerplate-keycloak:/tmp/export/boilerplate-realm.json ./keycloak/
```

### Importing Realm

Modify `docker-compose.yaml`:

```yaml
keycloak:
  # ... other configuration
  command: start-dev --import-realm
  volumes:
    - keycloak_data:/opt/keycloak/data
    - ./keycloak/boilerplate-realm.json:/opt/keycloak/data/import/realm.json
```

## Production Notes

⚠️ **Must change for production**:

1. **Admin password** change (not `admin/admin`)
2. **HTTPS** use instead of HTTP
3. **Client Secret** store securely (e.g. Secrets Manager)
4. **Token rotation** implement (Refresh Tokens)
5. **Realm backup** create regularly
6. **Rate limiting** enable
7. **Keycloak** run behind a reverse proxy

## Further Documentation

- [Keycloak Documentation](https://www.keycloak.org/documentation)
- [OpenID Connect Specs](https://openid.net/connect/)
- [JWT.io - Token Debugger](https://jwt.io/)

## Quick Reference

```bash
# Keycloak URLs
Admin Console:  http://localhost:8081
Realm:          http://localhost:8081/realms/boilerplate
JWKS:           http://localhost:8081/realms/boilerplate/protocol/openid-connect/certs
Token Endpoint: http://localhost:8081/realms/boilerplate/protocol/openid-connect/token

# Get token (one-liner)
curl -s -X POST http://localhost:8081/realms/boilerplate/protocol/openid-connect/token \
  -d "client_id=boilerplate-client" \
  -d "client_secret=YOUR_SECRET" \
  -d "grant_type=password" \
  -d "username=testuser" \
  -d "password=testpass123" \
  | jq -r .access_token

# Use token
TOKEN=$(curl -s -X POST ... | jq -r .access_token)
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/v1/projects
```
