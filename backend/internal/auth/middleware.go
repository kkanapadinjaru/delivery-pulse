package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Middleware validates Keycloak JWTs on every request and rejects those that
// are missing, malformed, or expired. In LOCAL_DEV mode it bypasses all
// validation and injects MockManager instead.
type Middleware struct {
	localDev  bool
	authority string // e.g. "https://keycloak.internal/realms/delivery-pulse"

	jwksURL  string
	mu       sync.RWMutex
	keyCache map[string]any // kid → public key
	cacheExp time.Time
}

// New returns a Middleware. If localDev is true, JWT validation is skipped.
func New(localDev bool, authority string) *Middleware {
	return &Middleware{
		localDev:  localDev,
		authority: authority,
		keyCache:  make(map[string]any),
	}
}

// Handler wraps the next handler, enforcing authentication.
func (m *Middleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if m.localDev {
			r = r.WithContext(WithUser(r.Context(), MockManager))
			next.ServeHTTP(w, r)
			return
		}

		user, err := m.authenticate(r)
		if err != nil {
			slog.Warn("authentication failed", "path", r.RequestURI, "remote", r.RemoteAddr, "error", err)
			http.Error(w, `{"error":"Unauthorized: `+err.Error()+`"}`, http.StatusUnauthorized)
			return
		}

		r = r.WithContext(WithUser(r.Context(), user))
		next.ServeHTTP(w, r)
	})
}

// RequireRole returns a middleware that enforces a minimum role level.
func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u := FromContext(r.Context())
			if u == nil || !u.HasRole(role) {
				http.Error(w, `{"error":"Forbidden: requires `+role+` role"}`, http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func (m *Middleware) authenticate(r *http.Request) (*User, error) {
	tokenStr := extractBearerToken(r)
	if tokenStr == "" {
		return nil, fmt.Errorf("missing Authorization header")
	}

	token, err := jwt.Parse(tokenStr, m.keyFunc, jwt.WithIssuedAt())
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return nil, fmt.Errorf("invalid claims")
	}

	user := &User{
		Email:       claimString(claims, "email"),
		DisplayName: claimString(claims, "preferred_username"),
	}
	if user.Email == "" {
		user.Email = claimString(claims, "sub")
	}

	// Keycloak stores realm roles in realm_access.roles.
	if ra, ok := claims["realm_access"].(map[string]any); ok {
		if roles, ok := ra["roles"].([]any); ok {
			for _, r := range roles {
				if s, ok := r.(string); ok {
					user.Roles = append(user.Roles, s)
				}
			}
		}
	}

	return user, nil
}

func (m *Middleware) keyFunc(token *jwt.Token) (any, error) {
	kid, _ := token.Header["kid"].(string)

	m.mu.RLock()
	key, cached := m.keyCache[kid]
	expired := time.Now().After(m.cacheExp)
	m.mu.RUnlock()

	if cached && !expired {
		return key, nil
	}

	// Refresh JWKS.
	if err := m.refreshJWKS(context.Background()); err != nil {
		slog.Error("JWKS refresh failed", "authority", m.authority, "error", err)
		return nil, fmt.Errorf("refreshing JWKS: %w", err)
	}

	m.mu.RLock()
	key, ok := m.keyCache[kid]
	m.mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("unknown key ID %q", kid)
	}
	return key, nil
}

func (m *Middleware) refreshJWKS(ctx context.Context) error {
	if m.jwksURL == "" {
		m.jwksURL = strings.TrimRight(m.authority, "/") + "/protocol/openid-connect/certs"
	}

	resp, err := http.Get(m.jwksURL)
	if err != nil {
		return fmt.Errorf("fetching JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks struct {
		Keys []struct {
			Kid string `json:"kid"`
			Kty string `json:"kty"`
			N   string `json:"n"`
			E   string `json:"e"`
		} `json:"keys"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("decoding JWKS: %w", err)
	}

	newCache := make(map[string]any, len(jwks.Keys))
	for _, k := range jwks.Keys {
		if k.Kty != "RSA" || k.Kid == "" || k.N == "" || k.E == "" {
			continue
		}
		pub, err := parseRSAJWK(k.N, k.E)
		if err != nil {
			continue
		}
		newCache[k.Kid] = pub
	}

	m.mu.Lock()
	m.keyCache = newCache
	m.cacheExp = time.Now().Add(10 * time.Minute)
	m.mu.Unlock()

	return nil
}

func extractBearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	if after, ok := strings.CutPrefix(h, "Bearer "); ok {
		return after
	}
	return ""
}

func claimString(claims jwt.MapClaims, key string) string {
	if v, ok := claims[key].(string); ok {
		return v
	}
	return ""
}

// parseRSAJWK reconstructs an RSA public key from base64url-encoded JWK n/e fields.
func parseRSAJWK(nB64, eB64 string) (*rsa.PublicKey, error) {
	nBytes, err := base64.RawURLEncoding.DecodeString(nB64)
	if err != nil {
		return nil, fmt.Errorf("decoding n: %w", err)
	}
	eBytes, err := base64.RawURLEncoding.DecodeString(eB64)
	if err != nil {
		return nil, fmt.Errorf("decoding e: %w", err)
	}
	e := new(big.Int).SetBytes(eBytes)
	if !e.IsInt64() {
		return nil, fmt.Errorf("RSA exponent too large")
	}
	return &rsa.PublicKey{N: new(big.Int).SetBytes(nBytes), E: int(e.Int64())}, nil
}
