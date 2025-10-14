package middleware

import (
	"net/http"
)

func SetAccessControlAllowOrigin(next http.Handler, val string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", val)
		next.ServeHTTP(w, r)
	})
}

func SetAccessControlAllowMethods(next http.Handler, val string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Methods", val)
		next.ServeHTTP(w, r)
	})
}

func SetAccessControlAllowHeaders(next http.Handler, val string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Headers", val)
		next.ServeHTTP(w, r)
	})
}
