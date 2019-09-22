package routing

import (
	"net/http"
)

// Authorize is a middleware that verifies the authorization token.
func Authorize(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")

		if token != "" {
			// TODO: Implement Auth
			next.ServeHTTP(w, r)
			return
		}

		http.Error(w, http.StatusText(403), 403)
	})
}

// CheckPermissions checks the example role-based access permissions.
func CheckPermissions(role string) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if token != "" {
				// TODO: Implement role-based authorization
				next.ServeHTTP(w, r)
				return
			}

			http.Error(w, http.StatusText(403), 403)
		})
	}
}
