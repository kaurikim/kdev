package server

import (
	"net/http"
	"strings"
)

func JwtAuthentication(next http.Handler, s *ExampleServer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v1/example/login" || r.URL.Path == "/v1/example/createuser" {
			next.ServeHTTP(w, r)
			return
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		s.mu.Lock()
		_, tokenValid := s.tokens[tokenString]
		s.mu.Unlock()

		if !tokenValid {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
