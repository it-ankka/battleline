package main

import (
	"log"
	"net/http"
	"os"

	. "github.com/it-ankka/battleline/internal/app/context"
	"github.com/it-ankka/battleline/internal/app/middleware"
	"github.com/it-ankka/battleline/internal/app/router"
)

func main() {
	address := ":8080"
	logger := log.New(os.Stdout, "", log.LstdFlags|log.Lshortfile)
	a := NewAppContext(logger)
	r := router.NewAppRouter(a)

	logger.Printf("Server started. Listening on %s\n", address)

	corsMiddleware := func(h http.Handler) http.Handler { return middleware.SetAccessControlAllowOrigin(h, "*") }
	loggerMiddleware := func(h http.Handler) http.Handler { return middleware.Logger(h, logger) }
	stack := middleware.CreateStack(loggerMiddleware, corsMiddleware)

	http.ListenAndServe(address, stack(r))
}
