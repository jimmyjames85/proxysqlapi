package server

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/http/pprof"
	"runtime/debug"
	runtimepprof "runtime/pprof"
	"strings"
	"sync"
	"time"

	healthcheck "github.com/docker/go-healthcheck"
	"github.com/go-chi/chi"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	"github.com/jimmyjames85/proxysqlapi/pkg/common"
)

type Config struct {
	common.DBConfig
	Port               int                  `envconfig:"PORT" required:"false" default:"16032"` // port to run on
	ConsulWatchEnabled bool                 `envconfig:"CONSUL_WATCH_ENABLED" required:"false" default:"false"`
	ConsulWatches      WatchDefinitionSlice `envconfig:"CONSUL_WATCHES" required:"false" default:"[]"`
	ConsulAddr         string               `envconfig:"CONSUL_ADDR" required:"false" default:"localhost:8500"`
	ConsulDatacenters  []string             `envconfig:"CONSUL_DATACENTERS" required:"false" default:""`
}

func (c *Config) ToJSON() string {
	copy := *c
	copy.DBPswd = "****"
	b, _ := json.Marshal(copy)
	return string(b)
}

type Server struct {
	cfg Config

	httpRouter    *chi.Mux
	httpServer    *http.Server
	httpEndpoints []Endpoint

	hc *healthcheck.Registry

	healthcheckRouter    *chi.Mux
	healthcheckServer    *http.Server
	healthcheckEndpoints []Endpoint

	consulWatchers    []*consulServiceWatcher
	lastKnownBackends map[int][]admin.MysqlServer
	backendLock       sync.Mutex

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
	return &Server{
		cfg:               cfg,
		lastKnownBackends: make(map[int][]admin.MysqlServer),
	}, nil
}

// Serve starts http server running on the port set in srv
func (s *Server) Serve() error {
	dbcfg := mysql.Config{
		Addr:              fmt.Sprintf("%s:%d", s.cfg.DBHost, s.cfg.DBPort),
		Passwd:            s.cfg.DBPswd,
		User:              s.cfg.DBuser,
		Net:               "tcp",
		InterpolateParams: true,
	}

	var err error
	s.psqlAdminDb, err = sql.Open("mysql", dbcfg.FormatDSN())
	if err != nil {
		return err
	}
	s.hc = healthcheck.NewRegistry()
	s.hc.RegisterPeriodicFunc("proxysql", time.Second*5, s.proxysqlHealthcheck)

	if s.cfg.ConsulWatchEnabled {
		err = s.watchConsul()
		if err != nil {
			s.psqlAdminDb.Close()
			return err
		}
		s.hc.RegisterPeriodicFunc("consul", time.Second*5, s.consulHealthcheck)
	}

	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.cfg.Port))
	if err != nil {
		s.psqlAdminDb.Close()
		return fmt.Errorf("unable to serve http - %v", err)
	}

	defer s.Close()
	return s.listen(httpListener)
}

// listen starts a server on the given listeners. It allows for easier testability of the server.
func (s *Server) listen(httpListener net.Listener) error {

	s.httpRouter = chi.NewRouter()
	s.httpEndpoints = []Endpoint{
		// root and healthchecks
		{Method: "GET", Path: "/", HandlerFunc: s.rootHandler},
		{Method: "GET", Path: "/healthcheck", HandlerFunc: s.healthcheckHandler},

		// load to memory
		{Method: "PUT", Path: "/load/config", HandlerFunc: s.loadConfigHandler},
		{Method: "PUT", Path: "/load/global_variables", HandlerFunc: s.loadGlobalVariablesHandler},
		{Method: "PUT", Path: "/load/mysql_query_rules", HandlerFunc: s.loadMysqlQueryRulesHanlder},
		{Method: "PUT", Path: "/load/mysql_servers", HandlerFunc: s.loadMysqlServersHandler},
		{Method: "PUT", Path: "/load/mysql_users", HandlerFunc: s.loadMysqlUsersHandler},

		// load to runtime
		{Method: "PUT", Path: "/load/runtime/config", HandlerFunc: s.loadConfigToRuntimeHandler},
		{Method: "PUT", Path: "/load/runtime/global_variables", HandlerFunc: s.loadGlobalVariablesToRuntimeHandler},
		{Method: "PUT", Path: "/load/runtime/mysql_query_rules", HandlerFunc: s.loadMysqlQueryRulesToRuntimeHanlder},
		{Method: "PUT", Path: "/load/runtime/mysql_servers", HandlerFunc: s.loadMysqlServersToRuntimeHandler},
		{Method: "PUT", Path: "/load/runtime/mysql_users", HandlerFunc: s.loadMysqlUsersToRuntimeHandler},

		// memory tables
		// {Method: "GET", Path: "/config", HandlerFunc: s.adminConfig},
		{Method: "GET", Path: "/global_variables", HandlerFunc: s.adminGlobalVariablesHandler},
		//{Method: "GET", Path: "/mysql_collations", HandlerFunc: s.adminMysqlCollationsHandler},
		//{Method: "GET", Path: "/mysql_group_replication_hostgroups", HandlerFunc: s.adminMysqlGroupReplicationHostgroupsHandler},
		{Method: "GET", Path: "/mysql_query_rules", HandlerFunc: s.adminMysqlQueryRulesHandler},
		//{Method: "GET", Path: "/mysql_query_rules_fast_routing", HandlerFunc: s.adminMysqlQueryRulesFastRoutingHandler},
		//{Method: "GET", Path: "/mysql_replication_hostgroups", HandlerFunc: s.adminMysqlReplicationHostgroupsHandler},
		{Method: "GET", Path: "/mysql_servers", HandlerFunc: s.adminMysqlServersHandler},
		{Method: "GET", Path: "/mysql_users", HandlerFunc: s.adminMysqlUsersHandler},
		//{Method: "GET", Path:"/proxysql_servers", HandlerFunc: s.adminProxysqlServersHandler},
		//{Method: "GET", Path:"/scheduler", HandlerFunc: s.adminSchedulerHandler},

		// runtime tables
		// {Method: "GET", Path: "/runtime/config", HandlerFunc: s.adminRuntimeConfigHandler},
		//{Method: "GET", Path: "/runtime/checksums_values", HandlerFunc: s.adminRuntimeChecksumsValuesHandler},
		{Method: "GET", Path: "/runtime/global_variables", HandlerFunc: s.adminRuntimeGlobalVariablesHandler},
		//{Method: "GET", Path: "/runtime/mysql_group_replication_hostgroups", HandlerFunc: s.adminRuntimeMysqlGroupReplicationHostgroupsHandler},
		{Method: "GET", Path: "/runtime/mysql_query_rules", HandlerFunc: s.adminRuntimeMysqlQueryRulesHandler},
		//{Method: "GET", Path: "/runtime/mysql_query_rules_fast_routing", HandlerFunc: s.adminRuntimeMysqlQueryRulesFastRoutingHandler},
		//{Method: "GET", Path: "/runtime/mysql_replication_hostgroups", HandlerFunc: s.adminRuntimeMysqlReplicationHostgroupsHandler},
		{Method: "GET", Path: "/runtime/mysql_servers", HandlerFunc: s.adminRuntimeMysqlServersHandler},
		{Method: "GET", Path: "/runtime/mysql_users", HandlerFunc: s.adminRuntimeMysqlUsersHandler},
		//{Method: "GET", Path: "/runtime/proxysql_servers", HandlerFunc: s.adminRuntimeProxysqlServersHandler},
		//{Method: "GET", Path: "/runtime/scheduler", HandlerFunc: s.adminRuntimeSchedulerHandler},

		// stats tables
		//{Method: "GET", Path: "/stats/global_variables", HandlerFunc: s.statsGlobalVariablesHandler},
		//{Method: "GET", Path: "/stats/memory_metrics", HandlerFunc: s.statsMemoryMetricsHandler},
		//{Method: "GET", Path: "/stats/mysql_commands_counters", HandlerFunc: s.statsMysqlCommandsCountersHandler},
		{Method: "GET", Path: "/stats/mysql_connection_pool", HandlerFunc: s.statsMysqlConnectionPoolHandler},
		//{Method: "GET", Path: "/stats/mysql_connection_pool_reset", HandlerFunc: s.statsMysqlConnectionPoolResetHandler},
		{Method: "GET", Path: "/stats/mysql_global", HandlerFunc: s.statsMysqlGlobalHandler},
		//{Method: "GET", Path: "/stats/mysql_prepared_statements_info", HandlerFunc: s.statsMysqlPreparedStatementsInfoHandler},
		//{Method: "GET", Path: "/stats/mysql_processlist", HandlerFunc: s.statsMysqlProcesslistHandler},
		{Method: "GET", Path: "/stats/mysql_query_digest", HandlerFunc: s.statsMysqlQueryDigestHandler},
		//{Method: "GET", Path: "/stats/mysql_query_digest_reset", HandlerFunc: s.statsMysqlQueryDigestResetHandler},
		{Method: "GET", Path: "/stats/mysql_query_rules", HandlerFunc: s.statsMysqlQueryRulesHandler},
		{Method: "GET", Path: "/stats/mysql_users", HandlerFunc: s.statsMysqlUsersHandler},
		//{Method: "GET", Path: "/stats/proxysql_servers_checksums", HandlerFunc: s.statsProxysqlServersChecksumsHandler},
		//{Method: "GET", Path: "/stats/proxysql_servers_metrics", HandlerFunc: s.statsProxysqlServersMetricsHandler},
		//{Method: "GET", Path: "/stats/proxysql_servers_status", HandlerFunc: s.statsProxysqlServersStatusHandler},

		// monitor tables
		//{Method: "GET", Path: "/monitor/mysql_server_connect_log", HandlerFunc: s.monitorMysqlServerConnectLogHandler},
		//{Method: "GET", Path: "/monitor/mysql_server_group_replication_log", HandlerFunc: s.monitorMysqlServerGroupReplicationLogHandler},
		{Method: "GET", Path: "/monitor/mysql_server_ping_log", HandlerFunc: s.monitorMysqlServerPingLogHandler},
		//{Method: "GET", Path: "/monitor/mysql_server_read_only_log", HandlerFunc: s.monitorMysqlServerReadOnlyLogHandler},
		//{Method: "GET", Path: "/monitor/mysql_server_replication_lag_log", HandlerFunc: s.monitorMysqlServerReplicationLagLogHandler},

		// pprof
		{Method: "GET", Path: "/debug/config", HandlerFunc: s.configHandler},
		{Method: "GET", Path: "/debug/pprof/cmdline", HandlerFunc: pprof.Cmdline},
		{Method: "GET", Path: "/debug/pprof/profile", HandlerFunc: pprof.Profile},
		{Method: "GET", Path: "/debug/pprof/symbol", HandlerFunc: pprof.Symbol},
		{Method: "GET", Path: "/debug/pprof/trace", HandlerFunc: pprof.Trace},
		{Method: "GET", Path: "/debug/pprof/", HandlerFunc: pprof.Index},
	}

	// service endpoints
	for _, ep := range s.httpEndpoints {
		s.httpRouter.MethodFunc(ep.Method, ep.Path, ep.HandlerFunc)
	}

	// runtime pprof endoints
	for _, p := range runtimepprof.Profiles() {
		s.httpEndpoints = append(s.httpEndpoints, Endpoint{Method: "GET", Path: "/debug/pprof/" + p.Name(), HandlerFunc: pprof.Index})
	}

	log.Printf("listening on %d", s.cfg.Port)
	s.httpServer = &http.Server{Addr: fmt.Sprintf(":%d", s.cfg.Port), Handler: panic(s.httpRouter)}
	s.httpServer.WriteTimeout = 1 * time.Minute
	s.httpServer.ReadTimeout = 1 * time.Minute

	if err := s.httpServer.Serve(httpListener); err != nil {
		if err != http.ErrServerClosed {
			return err
		}
	}
	return nil
}

func (s *Server) proxysqlHealthcheck() error {
	return s.psqlAdminDb.Ping()
}

func (s *Server) consulHealthcheck() error {
	errs := []string{}
	for _, w := range s.consulWatchers {
		err := w.CheckHealth()
		if err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		hcMsg := "Error(s) communicating with consul: " + strings.Join(errs, "; ")
		return fmt.Errorf(hcMsg)
	}
	return nil
}

func panic(h http.Handler) http.Handler {
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
func (s *Server) Close() error {

	defer s.psqlAdminDb.Close()
	// close socket to stop new requests from coming in
	return s.httpServer.Close()
}

func (s *Server) updateBackends(update hostgroupUpdate) error {
	err := admin.DropMysqlServerHostgroup(s.psqlAdminDb, update.hostgroupID)
	if err != nil {
		return err
	}

	err = admin.InsertMysqlServers(s.psqlAdminDb, update.mysqlServers...)
	if err != nil {
		return err
	}

	err = admin.LoadMysqlServersToRuntime(s.psqlAdminDb)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) watchConsul() error {

	ch := make(chan hostgroupUpdate)

	for _, def := range s.cfg.ConsulWatches {
		w, err := newConsulServiceWatcher(s.cfg.ConsulAddr, s.cfg.ConsulDatacenters, def, ch)
		if err != nil {
			return err
		}
		w.Start()
	}

	go func() {
		for {
			update := <-ch
			if update.primary && len(update.mysqlServers) > 1 {
				log.Printf("detected multiple primaries for hostgroup %d: aborting update", update.hostgroupID) // todo update logger
				continue
			}
			s.backendLock.Lock()
			s.lastKnownBackends[update.hostgroupID] = update.mysqlServers
			err := s.updateBackends(update)
			s.backendLock.Unlock()
			if err != nil {
				log.Println(err)
			}
		}
	}()
	return nil

}
