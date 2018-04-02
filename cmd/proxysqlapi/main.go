package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	m "github.com/jimmyjames85/proxysqlapi/pkg/metrics"
	"github.com/jimmyjames85/proxysqlapi/pkg/server"
	"github.com/kelseyhightower/envconfig"
)

func printHeader(h string) {
	fmt.Printf("----------------------------------------\n%s\n\n", h)
}

func printGlobalVariable(db *sql.DB, name string, runtime bool) {
	var vars map[string]string
	var err error
	if runtime {
		printHeader("PrintRuntimeGlobalVariable")
		vars, err = admin.SelectRuntimeGlobalVariables(db)
	} else {
		printHeader("PrintGlobalVariable")
		vars, err = admin.SelectGlobalVariables(db)
	}

	if err != nil {
		log.Fatal(err)
	}

	value, ok := vars[name]
	if !ok {
		value = fmt.Sprintf("no such global_variable: %q", name)
	}

	fmt.Printf("%s\n", value)
}

func printGlobalVariables(db *sql.DB, runtime bool) {
	var vars map[string]string
	var err error
	if runtime {
		printHeader("SelectRuntimeGlobalVariables")
		vars, err = admin.SelectRuntimeGlobalVariables(db)
	} else {
		printHeader("SelectGlobalVariables")
		vars, err = admin.SelectGlobalVariables(db)
	}
	if err != nil {
		log.Fatal(err)
	}
	for name, value := range vars {
		fmt.Printf("%55s:\t%s\n", name, value)
	}
}

func printMysqlServers(db *sql.DB, runtime bool) {

	var servers []admin.MysqlServer
	var err error
	if runtime {
		printHeader("SelectRuntimeMysqlServers")
		servers, err = admin.SelectRuntimeMysqlServers(db)
	} else {
		printHeader("SelectMysqlServers")
		servers, err = admin.SelectMysqlServers(db)
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, s := range servers {
		fmt.Printf("%s\n", s.ToJSON())
	}
}

func insertSomeMysqlServers(db *sql.DB, servers []admin.MysqlServer) {
	printHeader("InsertMysqlServers")
	err := admin.InsertMysqlServers(db, servers...)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("success\n")
}

func printMysqlUsers(db *sql.DB, runtime bool) {

	var users []admin.MysqlUser
	var err error
	if runtime {
		printHeader("SelectRuntimeMysqlUsers")
		users, err = admin.SelectRuntimeMysqlUsers(db)
	} else {
		printHeader("SelectMysqlUsers")
		users, err = admin.SelectMysqlUsers(db)
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range users {
		fmt.Printf("%s\n", u.ToJSON())
	}
}

func printMysqlQueryRules(db *sql.DB, runtime bool) {

	var rules []admin.MysqlQueryRule
	var err error
	if runtime {
		printHeader("SelectRuntimeMysqlQueryRules")
		rules, err = admin.SelectRuntimeMysqlQueryRules(db)
	} else {
		printHeader("SelectMysqlQueryRules")
		rules, err = admin.SelectMysqlQueryRules(db)
	}
	if err != nil {
		log.Fatal(err)
	}
	for _, u := range rules {
		fmt.Printf("%s\n", u.ToJSON())
	}
}

func insertSomeMysqlUsers(db *sql.DB) {
	printHeader("InsertMysqlUsers")

	err := admin.DropMysqlUsers(db)
	if err != nil {
		log.Fatal(err)
	}

	names := []string{"jim", "ron", "dharmik", "anthony", "eric", "david", "dustin"}
	var arr []admin.MysqlUser

	for _, n := range names {
		u := admin.NewMysqlUser(n)
		p := strings.Repeat(n, 2)
		u.Password = &p
		arr = append(arr, *u)
	}

	err = admin.InsertMysqlUsers(db, arr...)
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("success\n")
}

func main() {

	metricsCfg := m.Config{}
	envconfig.MustProcess("PROXYSQLAPI", &metricsCfg) // todo get better with this config maybe pass it in
	b, _ := json.Marshal(metricsCfg)
	fmt.Printf("metricsDbCfg: %s\n", string(b))
	hostname, _ := os.Hostname()
	ghost := "127.0.0.1"
	gport := 2003
	gaddr := fmt.Sprintf("%s:%d", ghost, gport)
	addr, err := net.ResolveTCPAddr("tcp", gaddr)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	emitter := m.NewEmitter(metricsCfg, nil, 10*time.Second, fmt.Sprintf("fix.this.%s", hostname), addr)
	go func() {
		eerr := emitter.Serve()
		if eerr != nil {
			log.Fatalf("%s", eerr)
		}
	}()

	cfg := server.Config{}
	envconfig.MustProcess("PROXYSQLAPI", &cfg)
	srv, err := server.New(cfg)

	//psqlCfg, err := admin.LoadConfig("example.json")
	if err != nil {
		log.Fatalf("err loading config: %v", err)
	}

	err = srv.Serve()

	if err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
