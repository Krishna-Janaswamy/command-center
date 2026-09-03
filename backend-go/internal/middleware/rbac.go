package middleware

import (
	"net/http"
	"strings"

	authpkg "github.com/example/service-virtualization-go/internal/auth"
)

// RequireRoles returns a middleware that enforces the presence of one of the allowed roles
func RequireRoles(jwtSvc *authpkg.JwtService, allowedRoles ...string) func(http.Handler) http.Handler {
	rolesMap := make(map[string]struct{}, len(allowedRoles))
	for _, r := range allowedRoles {
		rolesMap[r] = struct{}{}
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
				return
			}
			token := strings.TrimPrefix(authHeader, "Bearer ")
			claims, err := jwtSvc.ValidateToken(token)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			if _, ok := rolesMap[claims.Role]; !ok {
				http.Error(w, "forbidden: insufficient role", http.StatusForbidden)
				return
			}

			// Role is allowed — pass through
			next.ServeHTTP(w, r)
		})
	}
}
