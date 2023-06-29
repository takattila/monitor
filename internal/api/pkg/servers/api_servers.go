package servers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/pkg/logger"
)

var L logger.Logger

// ServeHTTP will run service on specific port.
func ServeHTTP(port int, router chi.Router) {
	L.Info("ServeHTTP", "Port:", port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	L.Fatal(server.ListenAndServe())
}
