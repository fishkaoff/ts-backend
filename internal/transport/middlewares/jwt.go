package middlewares

import (
	"context"
	"net/http"
	"strings"

	jwtclient "github.com/fishkaoff/ts-backend/internal/domain/lib/jwt"
)

type JwtChecker interface {
	IsValidJWT(jwtString string) (bool, error)
	ExtractClaims(tokenString string) (*jwtclient.CustomClaims, error)
}

type JWTMiddleware struct {
	checker JwtChecker
}

func NewJWTMiddleware(checker JwtChecker) *JWTMiddleware {
	return &JWTMiddleware{
		checker: checker,
	}
}

func (m *JWTMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		const prefix = "Bearer "
		if !strings.HasPrefix(authHeader, prefix) {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, prefix)

		claims, err := m.checker.ExtractClaims(tokenString)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.Id)
		r = r.WithContext(ctx)

		next.ServeHTTP(w, r)
	})
}
