package middleware

import (
	"net/http"
)
func EnsureAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		
	})
}