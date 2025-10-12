package main

import (
	"log/slog"
	"net/http"
	"os"

	. "github.com/it-ankka/battleline/internal/app/context"
	"github.com/it-ankka/battleline/internal/app/middleware"
	"github.com/it-ankka/battleline/internal/app/router"
)

func main() {
	address := ":8080"

	// Add contextual information here
	defaultAttrs := []slog.Attr{}
	handlerOptions := slog.HandlerOptions{
		AddSource: true,
	}
	// Debug mode enabled
	if _, debug := os.LookupEnv("DEBUG"); debug {
		handlerOptions.Level = slog.LevelDebug
	}
	slogHandler := slog.NewJSONHandler(os.Stdout, &handlerOptions).WithAttrs(defaultAttrs)
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
