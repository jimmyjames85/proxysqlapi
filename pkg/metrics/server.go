package metrics

import (
	"database/sql"
	"fmt"
	"log"
	"net"
	"time"

	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	metrics "github.com/rcrowley/go-metrics"
)

type Config struct {
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

type PsqlMetricEmitter struct {
	psqlCfg Config
	gcfg    graphite.Config
	db      *sql.DB
}

func NewEmitter(psqlCfg Config, r metrics.Registry, d time.Duration, prefix string, addr *net.TCPAddr) *PsqlMetricEmitter {
	if r == nil {
		r = metrics.DefaultRegistry
	}

	return &PsqlMetricEmitter{
		gcfg: graphite.Config{
			Addr:          addr,
			Registry:      r,
			FlushInterval: d,
			DurationUnit:  time.Nanosecond,
			Prefix:        prefix,
		},
		psqlCfg: psqlCfg,
	}
}

func (e *PsqlMetricEmitter) handleError(err error) {
	log.Printf("metric_emitter_err: %s\n", err) //TODO
}

// blocking function
func (e *PsqlMetricEmitter) emit() {
	for _ = range time.Tick(e.gcfg.FlushInterval) {
		fmt.Printf("emit to graphite\n")
		if err := graphite.Once(e.gcfg); nil != err {
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
		connPool, err := admin.SelectStatsMysqlConnectionPool(e.db)
		if err != nil {
			e.handleError(err)
			fmt.Printf("here: %v\n", e.db)
		}
		for _, c := range connPool {
			namespace := fmt.Sprintf("%d.%s_%d", c.Hostgroup, c.SrvHost, c.SrvPort)
			fmt.Printf("emitting namespace: %s\n", namespace)
			metrics.GetOrRegisterGauge(fmt.Sprintf("%s.conn_used", namespace), nil).Update(int64(c.ConnUsed))
			metrics.GetOrRegisterGauge(fmt.Sprintf("%s.conn_free", namespace), nil).Update(int64(c.ConnFree))
			metrics.GetOrRegisterGauge(fmt.Sprintf("%s.queries", namespace), nil).Update(int64(c.Queries))
			metrics.GetOrRegisterGauge(fmt.Sprintf("%s.latency_us", namespace), nil).Update(int64(c.LatencyUS))
		}
	}
	return nil
}
func (e *PsqlMetricEmitter) Close() error {
	if e.db != nil {
		return e.db.Close()
	}
	return nil

	// TODO
}