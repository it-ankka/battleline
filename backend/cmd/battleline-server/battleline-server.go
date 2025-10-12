package main

import (
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	. "github.com/it-ankka/battleline/internal/app/context"
	"github.com/it-ankka/battleline/internal/app/middleware"
	"github.com/it-ankka/battleline/internal/app/router"
)

func main() {
	address := ":8080"

	// Add contextual information here
	defaultAttrs := []slog.Attr{}

	// Use short filename
	handlerOptions := slog.HandlerOptions{
		AddSource: true,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
				}
			}
			return a
		},
	}
	// Debug mode enabled
	if mode, _ := os.LookupEnv("MODE"); strings.ToLower(mode) == "debug" {
		handlerOptions.Level = slog.LevelDebug
	}
	var slogHandler slog.Handler
	if logHandler, _ := os.LookupEnv("LOG_FORMAT"); strings.ToLower(logHandler) == "json" {
		slogHandler = slog.NewJSONHandler(os.Stdout, &handlerOptions).WithAttrs(defaultAttrs)
	} else {
		slogHandler = slog.NewTextHandler(os.Stdout, &handlerOptions).WithAttrs(defaultAttrs)
	}
	logger := slog.New(slogHandler)
	slog.SetDefault(logger)

	a := NewAppContext()
	r := router.NewAppRouter(a)

	slog.Info("Server started.", slog.String("address", address))

	corsMiddleware := func(h http.Handler) http.Handler { return middleware.SetAccessControlAllowOrigin(h, "*") }
	loggerMiddleware := func(h http.Handler) http.Handler { return middleware.Logger(h, slog.Default()) }
	stack := middleware.CreateStack(loggerMiddleware, corsMiddleware)

	http.ListenAndServe(address, stack(r))
}
