package server

import (
	"database/sql"
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
	"github.com/go-sql-driver/mysql"
)

type Config struct {
	Port   int    `envconfig:"PORT" required:"false" default:"16032"` // port to run on
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

func (c *Config) ToJSON() string {
	// TODO redact sensitive information
	b, _ := json.Marshal(c)
	return string(b)
}

type Server struct {
	cfg Config

	httpRouter    *chi.Mux
	httpServer    *http.Server
	httpEndpoints []Endpoint

	healthcheckRouter    *chi.Mux
	healthcheckEndpoints []Endpoint

	psqlAdminDb *sql.DB
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
func (s *Server) Serve() error {
	defer s.Close()
	var err error

	dbcfg := mysql.Config{
		Addr:              fmt.Sprintf("%s:%d", s.cfg.DBHost, s.cfg.DBPort),
		Passwd:            s.cfg.DBPswd,
		User:              s.cfg.DBuser,
		Net:               "tcp",
		InterpolateParams: true,
	}
	s.psqlAdminDb, err = sql.Open("mysql", dbcfg.FormatDSN())
	if err != nil {
		log.Fatal(err)
	}
	defer s.psqlAdminDb.Close()

	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		panic(fmt.Sprintf("unable to serve http - %v", err))
	}
	s.listen(httpListener)
	return nil
}

// listen starts a server on the given listeners. It allows for easier testability of the server.
func (s *Server) listen(httpListener net.Listener) {

	s.httpRouter = chi.NewRouter()

	s.httpEndpoints = []Endpoint{
		// root and healthchecks
		{Method: "GET", Path: "/", HandlerFunc: s.rootHandler},
		{Method: "GET", Path: "/load/config/{userID}", HandlerFunc: s.loadHandler},
		{Method: "GET", Path: "/load/config/to/runtime/{userID}", HandlerFunc: s.loadToRuntimeHandler},

		// tables
		{Method: "GET", Path: "/admin/mysql_servers", HandlerFunc: s.adminMysqlServersHandler},
		{Method: "GET", Path: "/admin/runtime/mysql_servers", HandlerFunc: s.adminRuntimeMysqlServersHandler},

		{Method: "GET", Path: "/admin/mysql_users", HandlerFunc: s.adminMysqlUsersHandler},
		{Method: "GET", Path: "/admin/runtime/mysql_users", HandlerFunc: s.adminRuntimeMysqlUsersHandler},

		{Method: "GET", Path: "/admin/mysql_query_rules", HandlerFunc: s.adminMysqlQueryRulesHandler},
		{Method: "GET", Path: "/admin/runtime/mysql_query_rules", HandlerFunc: s.adminRuntimeMysqlQueryRulesHandler},

		// pprof
		{Method: "GET", Path: "/debug/config", HandlerFunc: s.configHandler},
		{Method: "GET", Path: "/debug/pprof/cmdline", HandlerFunc: pprof.Cmdline},
		{Method: "GET", Path: "/debug/pprof/profile", HandlerFunc: pprof.Profile},
		{Method: "GET", Path: "/debug/pprof/symbol", HandlerFunc: pprof.Symbol},
		{Method: "GET", Path: "/debug/pprof/trace", HandlerFunc: pprof.Trace},
		{Method: "GET", Path: "/debug/pprof/", HandlerFunc: pprof.Index},
	}

	// runtime pprof endoints
	for _, p := range runtimepprof.Profiles() {
		s.httpEndpoints = append(s.httpEndpoints, Endpoint{Method: "GET", Path: "/debug/pprof/" + p.Name(), HandlerFunc: pprof.Index})
	}
	for _, ep := range s.httpEndpoints {
		s.httpRouter.MethodFunc(ep.Method, ep.Path, ep.HandlerFunc)
	}

	log.Printf("listening on %d", s.cfg.Port)
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", s.cfg.Port), Handler: Panic(s.httpRouter)}
	s.httpServer.WriteTimeout = 1 * time.Minute
	s.httpServer.ReadTimeout = 1 * time.Minute

	if err := s.httpServer.Serve(httpListener); err != nil {
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
				log.Println(rec, string(debug.Stack()))
			}
		}()
		h.ServeHTTP(w, r)
	})
}

// Close closes all db connections or any other clean up
func (srv *Server) Close() error {

	// close socket to stop new requests from coming in
	return srv.httpServer.Close()
}
