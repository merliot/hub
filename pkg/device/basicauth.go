//go:build !tinygo

package device

import (
	"crypto/sha256"
	"crypto/subtle"
	"net/http"
)

// HTTP Basic Authentication middleware
func _basicAuth(w http.ResponseWriter, r *http.Request) bool {
	var user = Getenv("USER", "")
	var passwd = Getenv("PASSWD", "")

	// skip basic authentication if no user
	if user == "" {
		return true
	}

	ruser, rpasswd, ok := r.BasicAuth()

	if ok {
		userHash := sha256.Sum256([]byte(user))
		passHash := sha256.Sum256([]byte(passwd))
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
func basicAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !_basicAuth(w, r) {
			return
		}
		// Call the next handler if the credentials are valid
		next.ServeHTTP(w, r)
	})
}
