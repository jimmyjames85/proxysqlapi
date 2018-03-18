package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"runtime/debug"
	runtimepprof "runtime/pprof"
	"time"

	"github.com/go-chi/chi"
)

type Config struct {
	Port   int    `envconfig:"PORT" required:"false" default:"16032"` // port to run on
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

func (c *Config) ToJSON() string {
	b, _ := json.Marshal(c)
	return string(b)
}

type Server struct {
	cfg Config

	httpRouter    *chi.Mux
	httpServer    *http.Server
	httpEndpoints []Endpoint

	healthcheckRouter    *chi.Mux
	healthcheckServer    *http.Server
	healthcheckEndpoints []Endpoint
}

// Endpoint is leveraged in handler.go in rootHandler, which prints out registered routes.
type Endpoint struct {
	Path        string
	HandlerFunc http.HandlerFunc
	Method      string
}

// New creates a new server
func New(cfg Config) (*Server, error) {
	return &Server{cfg: cfg}, nil
}

// Serve starts http server running on the port set in srv
func (srv *Server) Serve() {
	defer srv.Close()
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", srv.cfg.Port))
	if err != nil {
		panic(fmt.Sprintf("unable to serve http - %v", err))
	}
	srv.listen(httpListener)
}

// listen starts a server on the given listeners. It allows for easier testability of the server.
func (srv *Server) listen(httpListener net.Listener) {

	srv.httpRouter = chi.NewRouter()

	srv.httpEndpoints = []Endpoint{
		// root and healthchecks
		{Method: "GET", Path: "/", HandlerFunc: srv.rootHandler},
		{Method: "GET", Path: "/config", HandlerFunc: srv.configHandler},
		// pprof
		{Method: "GET", Path: "/debug/pprof/cmdline", HandlerFunc: pprof.Cmdline},
		{Method: "GET", Path: "/debug/pprof/profile", HandlerFunc: pprof.Profile},
		{Method: "GET", Path: "/debug/pprof/symbol", HandlerFunc: pprof.Symbol},
		{Method: "GET", Path: "/debug/pprof/trace", HandlerFunc: pprof.Trace},
		{Method: "GET", Path: "/debug/pprof/", HandlerFunc: pprof.Index},
	}

	// runtime pprof endoints
	for _, p := range runtimepprof.Profiles() {
		srv.httpEndpoints = append(srv.httpEndpoints, Endpoint{Method: "GET", Path: "/debug/pprof/" + p.Name(), HandlerFunc: pprof.Index})
	}
	for _, ep := range srv.httpEndpoints {
		srv.httpRouter.MethodFunc(ep.Method, ep.Path, ep.HandlerFunc)
	}

	log.Printf("listening on %d", srv.cfg.Port)
	srv.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", srv.cfg.Port), Handler: Panic(srv.httpRouter)}
	srv.httpServer.WriteTimeout = 1 * time.Minute
	srv.httpServer.ReadTimeout = 1 * time.Minute

	if err := srv.httpServer.Serve(httpListener); err != nil {
		if err != http.ErrServerClosed {
			log.Printf("server crash: %v", err)
			os.Exit(1)
		}
	}

}

func Panic(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Println(rec, debug.Stack())
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// Close closes all db connections or any other clean up
func (srv *Server) Close() error {

	srv.healthcheckServer.Close() // ignoring err

	// close socket to stop new requests from coming in
	return srv.httpServer.Close()
}
