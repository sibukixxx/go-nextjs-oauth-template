package middleware

import (
	"net/http"
	"strings"
)

// CORS returns a CORS middleware with the given allowed origins
func CORS(allowedOrigins []string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is in the allowed list
			if isOriginAllowed(origin, allowedOrigins) {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Handle preflight requests
			if r.Method == http.MethodOptions {
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
				w.Header().Set("Access-Control-Max-Age", "86400") // 24 hours
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// isOriginAllowed checks if the origin is in the allowed list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if len(allowedOrigins) == 0 {
		return false
	}

	for _, allowed := range allowedOrigins {
		// Exact match
		if origin == allowed {
			return true
		}

		// Wildcard matching (e.g., *.example.com)
		// *.example.com matches sub.example.com but not example.com
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*")
			// Origin must end with .example.com (subdomain required)
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
	}

	return false
}
