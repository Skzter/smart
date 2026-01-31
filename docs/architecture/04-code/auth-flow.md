# Level 4: Authentication Flow

This document describes the detailed code-level authentication flows in S.M.A.R.T, covering both frontend user authentication (Auth0) and internal service-to-service authentication.

## Diagrams

![Auth Flow](diagrams/auth-flow.mmd.svg)
See [auth-flow.mmd](diagrams/auth-flow.mmd) for the Mermaid source.

## Overview

S.M.A.R.T implements two distinct authentication mechanisms:
1. **Frontend User Authentication** - Auth0 OAuth for HTWK users
2. **Internal Service Authentication** - JWT tokens for service-to-service (MCP ↔ Autotester)

## 1. Frontend User Authentication (Auth0)

### Initial Load Flow

**Actors:**
- User (application user)
- Frontend Application
- Auth Service
- Auth0
- Autotester API

**Steps:**

1. **Application Load**
   - User navigates to S.M.A.R.T URL
   - Frontend loads index.html
   - Svelte app initializes

2. **Auth Initialization**
   - `auth.initAuth()` called on app mount
   - Fetches `/auth_config.json`:
     ```json
     {
       "domain": "jokresner.eu.auth0.com",
       "clientId": "abc123..."
     }
     ```

3. **Create Auth0 Client**
   - `createAuth0Client()` from Auth0 SDK
   - Configures:
     - Domain: Application Auth0 tenant
     - Client ID
     - Redirect URI: window.location.origin
     - Cache: localStorage

4. **Check Authentication State**
   - `auth0Client.isAuthenticated()`
   - Checks localStorage for cached session
   - If authenticated: retrieve user profile
   - If not: user sees login screen

5. **Redirect Check**
   - Check URL for `?code=` and `?state=` params
   - If present: user returned from Auth0
   - Call `handleRedirectCallback()`
   - Parse authorization code
   - Exchange for tokens
   - Store in cache

6. **Update Application State**
   - `auth` store updated:
     ```typescript
     {
       isAuthenticated: true,
       user: { sub: "auth0|123", email: "user@htwk.de" },
       auth0Client: client
     }
     ```
   - App re-renders with authenticated state

---

### Login Flow

**Steps:**

1. **User Clicks Login**
   - Login button triggers `auth.login()`

2. **Initiate OAuth Flow**
   - `auth0Client.loginWithRedirect()` called
   - Redirects to Auth0 login page
   - URL: `https://jokresner.eu.auth0.com/authorize?...`
   - Parameters:
     - `client_id`: App client ID
     - `redirect_uri`: App URL
     - `response_type`: code
     - `scope`: openid profile email
     - `state`: CSRF token

3. **Auth0 Authentication**
   - User enters application credentials
   - Auth0 validates with HTWK SSO
   - Auth0 may show consent screen (first time)

4. **Authorization Code Grant**
   - Auth0 redirects back to app
   - URL: `https://sp21.imn.htwk-leipzig.de/?code=abc&state=xyz`

5. **Token Exchange**
   - `handleRedirectCallback()` called automatically
   - SDK exchanges code for tokens:
     - Access token
     - ID token (JWT)
     - Refresh token
   - Tokens stored in localStorage

6. **User Profile Retrieval**
   - `auth0Client.getUser()` called
   - Parses ID token
   - Returns user profile

7. **Authenticated State**
   - App redirected to main interface
   - User ID stored in shared state
   - Subsequent API calls include user ID

---

### Logout Flow

**Steps:**

1. **User Clicks Logout**
   - Logout button triggers `auth.logout()`

2. **Clear Local Session**
   - `auth0Client.logout()` called
   - Clears localStorage
   - Clears cookies

3. **Auth0 Logout**
   - Redirects to Auth0 logout endpoint
   - URL: `https://jokresner.eu.auth0.com/v2/logout?...`
   - Parameters:
     - `client_id`
     - `returnTo`: App URL

4. **Session Termination**
   - Auth0 clears SSO session
   - Redirects back to app

5. **Unauthenticated State**
   - App shows login screen
   - User must re-authenticate to access

---

## 2. Internal Service Authentication (MCP ↔ Autotester)

**Purpose:** MCP authenticates to Autotester by sending a **bearer token in the `Authorization` header**. MCP does **not** call `/auth/generate`; the token is obtained by an authenticated user (e.g. via Frontend) and provided to MCP at startup.

**Actors:**
- Authenticated user (obtains token via Frontend)
- MCP Service (receives token at startup, sends it on every request)
- Autotester API (validates token)

### Token Provision (user → MCP)

1. **User obtains token**
   - Authenticated user (from allowed networks) calls `/api/v1/auth/generate` via Frontend (see [1. Frontend Authentication](#1-frontend-authentication) for user auth).
   - Nginx and Autotester enforce IP restriction (internal only).
   - Autotester returns JWT; Frontend (or user) receives it.

2. **Token provided to MCP at startup**
   - Bearer token is passed to MCP via configuration (e.g. IDE/CLI env or config).
   - MCP stores the token in memory for the lifetime of the process.

3. **MCP does not call `/auth/generate`**
   - No token request is made by MCP to Autotester.
   - No automatic refresh by MCP; if the token expires, the user must obtain a new token and restart/reconfigure MCP.

---

### API Request with Token

**Steps:**

1. **MCP Tool Invocation**
   - LLM calls MCP tool (e.g., `generate_test`)
   - Tool needs to call Autotester API

2. **Repository Prepares Request**
   - Autotester API Repository is called
   - Uses the bearer token provided at startup
   - Adds to request headers:
     ```go
     req.Header.Set("Authorization", "Bearer "+token)
     ```

3. **Request to Autotester**
   - POST to `/api/v1/chat` (or other endpoint)
   - Includes `Authorization: Bearer <token>` header

4. **Token Validation**
   - Autotester validates token:
     ```go
     tokenString := extractToken(req)
     token, err := jwt.Parse(tokenString, keyFunc)
     if err != nil || !token.Valid {
         return ErrUnauthorized
     }
     ```
   - Checks:
     - Signature valid
     - Not expired
     - Issuer correct

5. **Authorization Success**
   - Request proceeds to handler
   - Normal processing
   - Response returned

(MCP does not perform token refresh; the user must obtain a new token via Frontend and reconfigure MCP if the token expires.)

---

## Security Considerations

### Frontend Authentication

1. **OAuth 2.0 / OIDC**
   - Industry-standard authentication
   - Secure token exchange
   - No password handling in app

2. **Token Storage**
   - localStorage (acceptable for SPAs)
   - HttpOnly cookies (more secure, future enhancement)

3. **CSRF Protection**
   - `state` parameter in OAuth flow
   - Validates on callback

4. **Token Expiry**
   - Short-lived access tokens
   - Refresh tokens for renewal
   - Automatic re-authentication

### Internal Authentication

1. **Network Isolation**
   - Two-layer protection (Nginx + app)
   - CIDR-based IP validation
   - No external access possible

2. **JWT Tokens**
   - Signed with secret key
   - Tamper-proof
   - Time-limited (24h)

3. **Token Storage**
   - Database storage enables revocation
   - Audit trail of token generation
   - Can blacklist compromised tokens

4. **Defense in Depth**
   - Nginx blocks external IPs
   - Application validates again
   - Both layers must be bypassed for attack

---

## Code Examples

### Frontend: Auth Initialization
```typescript
// lib/authService.ts
const initAuth = async () => {
    const response = await fetch("/auth_config.json");
    const authConfig = await response.json();
    
    const auth0Client = await createAuth0Client({
        domain: authConfig.domain,
        clientId: authConfig.clientId,
        cacheLocation: "localstorage",
        authorizationParams: {
            redirect_uri: window.location.origin,
        },
    });
    
    // Handle redirect callback
    if (window.location.search.includes("code=")) {
        await auth0Client.handleRedirectCallback();
        window.history.replaceState({}, document.title, "/");
    }
    
    const isAuthenticated = await auth0Client.isAuthenticated();
    const user = await auth0Client.getUser();
    
    set({ isAuthenticated, user, auth0Client });
};
```

### Backend: Internal-Only Middleware
```go
// internal/autotester/application/router.go
func internalOnlyMiddleware(logger *slog.Logger) gin.HandlerFunc {
    allowedCIDRs := []string{
        "127.0.0.0/8", "::1/128",
        "172.16.0.0/12", "192.168.0.0/16",
    }
    
    allowedNets := make([]*net.IPNet, 0)
    for _, cidr := range allowedCIDRs {
        _, ipNet, _ := net.ParseCIDR(cidr)
        allowedNets = append(allowedNets, ipNet)
    }
    
    return func(c *gin.Context) {
        clientIP := net.ParseIP(c.ClientIP())
        
        for _, ipNet := range allowedNets {
            if ipNet.Contains(clientIP) {
                c.Next()
                return
            }
        }
        
        logger.Warn("Blocked external access", "ip", clientIP)
        c.AbortWithStatusJSON(403, gin.H{
            "error": "Forbidden: Internal networks only",
        })
    }
}
```

### Backend: Token Generation
```go
// internal/autotester/domain/service/auth.go
func (s *AuthService) GenerateToken(userId string) (string, error) {
    claims := jwt.MapClaims{
        "sub": userId,
        "iat": time.Now().Unix(),
        "exp": time.Now().Add(24 * time.Hour).Unix(),
        "iss": "autotester",
    }
    
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    signed, err := token.SignedString([]byte(s.secretKey))
    if err != nil {
        return "", err
    }
    
    // Store in database
    if err := s.tokenRepo.SaveToken(userId, signed, time.Now().Add(24*time.Hour)); err != nil {
        return "", err
    }
    
    return signed, nil
}
```

---

## Monitoring and Logging

### Authentication Events
- Log all authentication attempts
- Track token generation
- Monitor blocked access attempts
- Alert on unusual patterns

### Metrics
- Login success/failure rates
- Token generation frequency
- Blocked request counts
- Token validation latency

### Datadog Integration
- Trace authentication flows
- APM spans for each step
- Error tracking
- Performance monitoring
