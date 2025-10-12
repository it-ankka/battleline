package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
)

func logBody(r *http.Request, logger *slog.Logger) {
	if r.Body != http.NoBody {
		body, err := io.ReadAll(r.Body)
		r.Body = io.NopCloser(bytes.NewBuffer(body))
		if err != nil {
			logger.Error("Failed to read request body", slog.Any("error", err.Error()))
			return
		}

		var bodyData any = body

		if strings.ToLower(r.Header.Get("Content-Type")) == "application/json" {
			err := json.Unmarshal(body, &bodyData)
			if err != nil {
				logger.Error("Failed to read request body", slog.Any("error", err.Error()))
				return
			}
		}

		logger.Debug("HTTP Request body", slog.Any("body", bodyData))
	}
}

func Logger(next http.Handler, logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger.Info("HTTP Request", slog.String("method", r.Method), slog.String("path", r.URL.Path))
		logger.Debug("HTTP Headers", slog.Any("headers", r.Header))

		logBody(r, logger)
		next.ServeHTTP(w, r)
	})
}
