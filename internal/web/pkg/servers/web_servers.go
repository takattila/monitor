package servers

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi"
	"github.com/takattila/monitor/pkg/common"
	"github.com/takattila/settings-manager"
	"golang.org/x/crypto/acme/autocert"
)

type Server struct {
	Port       int
	Domain     string
	Router     chi.Router
	RoutePath  string
	ProgramDir string
	FilesDir   string
	Cfg        *settings.Settings
}

func (s *Server) Start() {
	if s.Port == 443 {
		s.ServeTLS()
	} else {
		s.ServeHTTP()
	}
}

// ServeHTTP will run service on specific port.
func (s *Server) ServeHTTP() {
	common.Info("Port:", s.Port)

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.Port),
		Handler: s.Router,
	}

	common.Fatal(server.ListenAndServe())
}

// ServeTLS runs service with TLS config on a specific domain.
func (s *Server) ServeTLS() {
	common.Info("Port:", s.Port)
	common.Info("Domain:", s.Domain)

	certManager := autocert.Manager{
		Prompt:     autocert.AcceptTOS,
		HostPolicy: autocert.HostWhitelist(s.Domain),
	}

	// optionally use a cache dir
	dir := cacheDir()
	if dir != "" {
		certManager.Cache = autocert.DirCache(dir)
	}

	// create the server itself
	server := &http.Server{
		Addr: fmt.Sprintf(":%d", s.Port),
		TLSConfig: &tls.Config{
			GetCertificate: certManager.GetCertificate,
		},
		Handler: s.Router,
	}

	go func() {
		common.Fatal(http.ListenAndServe(":http", certManager.HTTPHandler(nil)))
	}()

	common.Fatal(server.ListenAndServeTLS("", "")) // Key and cert are coming from Let's Encrypt
}

// Files conveniently sets up a http.FileServer handler to serve
// static files from a http.FileSystem.
func (s *Server) Files() {
	notAllowed := "{}*"
	if strings.ContainsAny(s.RoutePath, notAllowed) {
		common.Warning("Does not permit any URL parameters:", notAllowed)
		return
	}

	if s.RoutePath != "/" && s.RoutePath[len(s.RoutePath)-1] != '/' {
		s.Router.Get(s.RoutePath, http.RedirectHandler(s.RoutePath+"/", 301).ServeHTTP)
		s.RoutePath += "/"
	}
	s.RoutePath += "*"

	s.Router.Get(s.RoutePath, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(http.Dir(filepath.Join(s.ProgramDir, s.FilesDir))))

		common.Info("r.URL.Path", r.URL.Path)
		fs.ServeHTTP(w, r)
	})
}

// cacheDir makes a consistent cache directory inside /tmp. Returns "" on error.
func cacheDir() (dir string) {
	if u, _ := user.Current(); u != nil {
		dir = filepath.Join(os.TempDir(), "cache-golang-autocert-"+u.Username)
		if err := os.MkdirAll(dir, 0700); err == nil {
			return dir
		}
	}
	return ""
}
