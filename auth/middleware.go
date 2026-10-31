package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"anthology-live/config"
)

const (
	CookieSessionName = "anthology_live_session"
	sessionSecret     = "anthology_live_secret_salt_2026"
)

// GenerateSessionToken creates a signed token for web session cookies.
func GenerateSessionToken(username string) string {
	ts := strconv.FormatInt(time.Now().Add(7*24*time.Hour).Unix(), 10)
	payload := username + ":" + ts
	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(payload))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return base64.RawURLEncoding.EncodeToString([]byte(payload)) + "." + sig
}

// ValidateSessionToken verifies a web session token and returns the username.
func ValidateSessionToken(token string) (string, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 2 {
		return "", false
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return "", false
	}
	payload := string(payloadBytes)
	pParts := strings.Split(payload, ":")
	if len(pParts) != 2 {
		return "", false
	}
	username := pParts[0]
	expStr := pParts[1]
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return "", false
	}

	mac := hmac.New(sha256.New, []byte(sessionSecret))
	mac.Write([]byte(payload))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(parts[1]), []byte(expectedSig)) {
		return "", false
	}
	return username, true
}

// ExtractCredentials checks query params, basic auth, or cookie.
func ExtractCredentials(r *http.Request) (username, password, token string) {
	q := r.URL.Query()

	// 1. Query parameters (Xtream Codes / M3U)
	username = q.Get("username")
	if username == "" {
		username = q.Get("u")
	}
	password = q.Get("password")
	if password == "" {
		password = q.Get("p")
	}
	token = q.Get("token")
	if token == "" {
		token = q.Get("key")
	}

	// 2. HTTP Basic Auth
	if username == "" && password == "" {
		u, p, ok := r.BasicAuth()
		if ok {
			username = u
			password = p
		}
	}

	// 3. Bearer Token
	if token == "" {
		authHeader := r.Header.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			token = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// 4. Session Cookie (Web)
	if username == "" {
		if c, err := r.Cookie(CookieSessionName); err == nil && c != nil {
			if u, ok := ValidateSessionToken(c.Value); ok {
				username = u
			}
		}
	}

	return username, password, token
}

// AuthMiddleware wraps an http.Handler with multi-layer authentication.
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg := config.Get()
		db := DB()

		// 1. Check Master Secret Key bypass if configured
		username, password, token := ExtractCredentials(r)
		if cfg.SecretKey != "" && (token == cfg.SecretKey || password == cfg.SecretKey) {
			ctx := WithUser(r.Context(), &User{ID: 0, Username: "admin"})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 2. If DB has no users, allow access (first-time setup friendly)
		if db == nil || !db.HasUsers() {
			ctx := WithUser(r.Context(), &User{ID: 1, Username: "guest"})
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		// 3. If credentials provided, authenticate against SQLite
		if username != "" {
			if password != "" {
				user, err := db.Authenticate(username, password)
				if err == nil {
					ctx := WithUser(r.Context(), user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			} else {
				// Checked via valid session cookie
				user, err := db.GetUserByUsername(username)
				if err == nil {
					ctx := WithUser(r.Context(), user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
		}

		// Unauthorized response
		w.Header().Set("WWW-Authenticate", `Basic realm="Anthology Live"`)
		http.Error(w, fmt.Sprintf("Yetkisiz erişim. Lütfen kullanıcı adı ve şifrenizi girin."), http.StatusUnauthorized)
	})
}
