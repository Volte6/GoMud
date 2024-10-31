package web

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/volte6/gomud/users"
)

var (
	authCache = map[string]time.Time{}
)

func handlerToHandlerFunc(h http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		h.ServeHTTP(w, r)
	}
}

func doBasicAuth(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		authHeader := r.Header.Get("Authorization")

		if t, ok := authCache[authHeader]; ok {

			if t.After(time.Now()) {
				next.ServeHTTP(w, r)
				return
			}

			delete(authCache, authHeader)
		}

		// Extract the username and password from the request
		// Authorization header. If no Authentication header is present
		// or the header value is invalid, then the 'ok' return value
		// will be false.
		username, password, ok := r.BasicAuth()
		if ok {

			// Authorize against actual user record
			uRecord, err := users.LoadUser(username, true)
			if err == nil {

				if uRecord.PasswordMatches(password) {

					if uRecord.Permission == users.PermissionAdmin || uRecord.Permission == users.PermissionMod {

						slog.Warn("ADMIN LOGIN", "username", username, "success", true)

						// Cache auth for 30 minutes to avoid re-auth every load
						authCache[authHeader] = time.Now().Add(time.Minute * 30)

						next.ServeHTTP(w, r)
						return

					} else {

						slog.Error("ADMIN LOGIN", "username", username, "success", false, "error", `Permissions=`+uRecord.Permission)

					}
				}

			} else {
				slog.Error("ADMIN LOGIN", "username", username, "success", false, "error", err)
			}
		}

		// If the Authentication header is not present, is invalid, or the
		// username or password is wrong, then set a WWW-Authenticate
		// header to inform the client that we expect them to use basic
		// authentication and send a 401 Unauthorized response.
		w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	})
}
