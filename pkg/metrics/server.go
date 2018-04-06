package metrics

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"strings"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	cache "github.com/patrickmn/go-cache"
	metrics "github.com/rcrowley/go-metrics"
)

type Config struct {
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

type PsqlMetricEmitter struct {
	cache                *cache.Cache
	cacheRefreshInterval time.Duration
	psqlCfg              Config
	graphiteCfg          graphite.Config
	db                   *sql.DB
}

func NewEmitter(psqlCfg Config, r metrics.Registry, flushInterval time.Duration, prefix string, addr *net.TCPAddr) *PsqlMetricEmitter {
	if r == nil {
		r = metrics.DefaultRegistry
	}

	return &PsqlMetricEmitter{
		graphiteCfg: graphite.Config{
			Addr:          addr,
			Registry:      r,
			FlushInterval: flushInterval,
			DurationUnit:  time.Nanosecond,
			Prefix:        prefix,
		},
		psqlCfg:              psqlCfg,
		cache:                cache.New(cache.NoExpiration, 0*time.Second), // the second parameter is the purge interval, and we never want to purge
		cacheRefreshInterval: 1 * time.Second,                              //TODO
	}
}

func (e *PsqlMetricEmitter) handleError(err error) {
	log.Printf("metric_emitter_err: %s\n", err) //TODO
}

// blocking function
func (e *PsqlMetricEmitter) emit() {
	for _ = range time.Tick(e.graphiteCfg.FlushInterval) {
		fmt.Printf("emit to graphite\n")
		if err := graphite.Once(e.graphiteCfg); nil != err {
			e.handleError(err)
		}
	}
}

func (e *PsqlMetricEmitter) Serve() error {

	dbcfg := mysql.Config{
		Addr:              fmt.Sprintf("%s:%d", e.psqlCfg.DBHost, e.psqlCfg.DBPort),
		Passwd:            e.psqlCfg.DBPswd,
		User:              e.psqlCfg.DBuser,
		Net:               "tcp",
		InterpolateParams: true,
	}
	fmt.Printf("emitter connected: %v\n", dbcfg.FormatDSN())
	var err error
	e.db, err = sql.Open("mysql", dbcfg.FormatDSN())
	if err != nil {
		return err
	}
	defer e.Close()
	go e.emit()

	// todo make 10 sec configurable
	for _ = range time.Tick(1 * time.Second) {
		fmt.Printf("local emit\n")
		e.connectionPool()
	}
	return nil
}

func (e *PsqlMetricEmitter) connectionPool() {
	connPool, err := admin.SelectStatsMysqlConnectionPool(e.db)
	if err != nil {
		e.handleError(err)
		fmt.Printf("here: %v\n", e.db)
		return
	}
	for _, c := range connPool {
		namespace := fmt.Sprintf("%d.%s_%d", c.Hostgroup, c.SrvHost, c.SrvPort)
		fmt.Printf("emitting namespace: %s\n", namespace)
		metrics.GetOrRegisterGauge(fmt.Sprintf("%s.conn_used", namespace), e.graphiteCfg.Registry).Update(int64(c.ConnUsed))
		metrics.GetOrRegisterGauge(fmt.Sprintf("%s.conn_free", namespace), e.graphiteCfg.Registry).Update(int64(c.ConnFree))
		metrics.GetOrRegisterGauge(fmt.Sprintf("%s.queries", namespace), e.graphiteCfg.Registry).Update(int64(c.Queries))
		metrics.GetOrRegisterGauge(fmt.Sprintf("%s.latency_us", namespace), e.graphiteCfg.Registry).Update(int64(c.LatencyUS))
	}
}

func (e *PsqlMetricEmitter) updateStatsMysqlConnectionPool() {
	connPool, err := admin.SelectStatsMysqlConnectionPool(e.db)
	if err != nil {
		e.handleError(err)
		return
	}
	for _, c := range connPool {
		host := strings.Replace(c.SrvHost, ".", "-", -1)
		connUsed := mFmt("%d.%s_%d.conn_used", c.Hostgroup, host, c.SrvPort)
		// connFree := mFmt("%d.%s_%d.conn_free", c.Hostgroup, host, c.SrvPort)
		metrics.GetOrRegisterGauge(connUsed, e.graphiteCfg.Registry).Update(int64(c.ConnUsed))
		// metrics.GetOrRegisterGauge(fmt.Sprintf("%s.conn_free", namespace), e.graphiteCfg.Registry).Update(int64(c.ConnFree))
		// metrics.GetOrRegisterGauge(fmt.Sprintf("%s.queries", namespace), e.graphiteCfg.Registry).Update(int64(c.Queries))
		// metrics.GetOrRegisterGauge(fmt.Sprintf("%s.latency_us", namespace), e.graphiteCfg.Registry).Update(int64(c.LatencyUS))
	}
}

// func mFmtHostname(hostname string) string {
// 	return strings.Replace(strings.Replace(hostname, ".sendgrid.net", "", -1), ".", "-", -1)
// }

// mFmt creates a metrics compatible formatted string (ie, no spaces or slashes)
// e.g. metrics.GetOrRegisterGauge(mFmt("stats_mysql_query_rules.%d.count_star", qR.RuleID), s.metricsRegistry).Update(int64(qR.Hits))
func mFmt(format string, a ...interface{}) string {
	f := fmt.Sprintf(format, a...)
	f = strings.Replace(f, " ", "_", -1)
	f = strings.Replace(f, "/", "__", -1)
	return f
}

func (e *PsqlMetricEmitter) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil

	// TODO
}
