package shared

import (
	"log"
	"net/http"
	"net/url"
	"strings"
)

func CORS(allowedHosts []string) func(http.Handler) http.Handler {
	allowed := make(map[string]struct{}, len(allowedHosts))
	allowAll := false
	for _, h := range allowedHosts {
		h = strings.TrimSpace(h)
		if h == "" {
			continue
		}
		if h == "*" {
			allowAll = true
			continue
		}
		allowed[normalizeAllowedKey(h)] = struct{}{}
	}

	if len(allowedHosts) > 0 {
		log.Printf("CORS configured with hosts: %v", allowedHosts)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")
			if origin != "" {
				if allowAll || hostAllowed(origin, allowed) {
					w.Header().Set("Access-Control-Allow-Origin", origin)
					w.Header().Add("Vary", "Origin")
					w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PATCH,DELETE,OPTIONS")
					w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")
				}
			}

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func hostAllowed(origin string, allowed map[string]struct{}) bool {
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}
	if u.Host != "" {
		_, ok := allowed[u.Host]
		return ok
	}
	_, ok := allowed[origin]
	return ok
}

func normalizeAllowedKey(entry string) string {
	entry = strings.TrimSpace(entry)
	if u, err := url.Parse(entry); err == nil && u.Host != "" {
		return u.Host
	}
	return entry
}
