//go:build !tinygo

package device

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// HTTP Basic Authentication middleware
func (s *server) _basicAuth(w http.ResponseWriter, r *http.Request) bool {

	// skip basic authentication if no user
	if s.user == "" {
		return true
	}

	ruser, rpasswd, ok := r.BasicAuth()

	if ok {
		userHash := sha256.Sum256([]byte(s.user))
		passHash := sha256.Sum256([]byte(s.passwd))
		ruserHash := sha256.Sum256([]byte(ruser))
		rpassHash := sha256.Sum256([]byte(rpasswd))

		// https://www.alexedwards.net/blog/basic-authentication-in-go
		userMatch := (subtle.ConstantTimeCompare(userHash[:], ruserHash[:]) == 1)
		passMatch := (subtle.ConstantTimeCompare(passHash[:], rpassHash[:]) == 1)

		if userMatch && passMatch {
			return true
		}
	}

	w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
	http.Error(w, "Unauthorized", http.StatusUnauthorized)

	return false
}

// basicAuth middleware function for http.Handler
func (s *server) basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !s._basicAuth(w, r) {
			return
		}
		// Call the next handler if the credentials are valid
		next.ServeHTTP(w, r)
	})
}
