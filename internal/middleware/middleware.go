package middleware

import (
	"context"
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
)

func AuthMiddleware(authClient *auth.Client, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Token não fornecido", http.StatusUnauthorized)
			return
		}

		splitToken := strings.Split(authHeader, "Bearer ")
		if len(splitToken) != 2 {
			http.Error(w, "Formato de token inválido", http.StatusUnauthorized)
			return
		}
		tokenString := splitToken[1]

		token, err := authClient.VerifyIDToken(r.Context(), tokenString)
		if err != nil {
			http.Error(w, "Token inválido ou expirado", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), "userUID", token.UID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}