package middleware

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
)

func Logger(next http.Handler, logger *log.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s := fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		body, err := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		if err != nil {
			logger.Printf("%s body=%s\n", s, string(body))
		} else {
			logger.Println(s)
		}
		next.ServeHTTP(w, r)
	})
}
