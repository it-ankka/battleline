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
	http.ListenAndServe(address, middleware.Logger(r, logger))
}
