package servers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/pkg/common"
)

// ServeHTTP will run service on specific port.
func ServeHTTP(port int, router chi.Router) {
	common.Info("ServeHTTP", "Port:", port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
	}

	common.Fatal(server.ListenAndServe())
}
