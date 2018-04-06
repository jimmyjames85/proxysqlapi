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

	_ "github.com/cyberdelia/go-metrics-graphite"
	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	m "github.com/jimmyjames85/proxysqlapi/pkg/metrics"
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

func insertNewPerson(db *sql.DB) error {
	stmt := `INSERT into gotham.person (first, middle, last) VALUES (?,?,?) -- gotham`
	_, err := db.Exec(stmt, "ji'my", "bo", "slice")
	if err != nil {
		return err
	}
	return nil
}

func printGotham(db *sql.DB) error {
	stmt := `SELECT
		 first,
		 middle,
		 last
		 FROM gotham.person -- gotham`

	rows, err := db.Query(stmt)
	if err != nil {
		return err
	}

	defer rows.Close()
	for rows.Next() {
		var f, m, l string
		err = rows.Scan(&f, &m, &l)
		if err != nil {
			return err
		}
		fmt.Printf("%10s %15s %10s\n", f, m, l)
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	return nil
}

func insertNewServer(psqlDB *sql.DB) error {
	// stmt := `INSERT into mysql_servers (hostname, comment) VALUES ('fakehost','host\'s comment goes here')`
	// _, err := psqlDB.Exec(stmt)
	stmt := `INSERT into mysql_servers (hostname, comment) VALUES (?,?)` // ('fakehost','host's comment goes here')`
	_, err := psqlDB.Exec(stmt, "fakehost", "host's info goes here")
	if err != nil {
		return err
	}
	return nil
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

	return

	// go graphite.Graphite(metrics.DefaultRegistry, (10 * time.Second), fmt.Sprintf("psql.%s", hostname), addr)

	// registry := metrics.NewPrefixedRegistry()

	// keepGoing := true
	// go func() {
	// 	time.Sleep(time.Minute)
	// 	fmt.Printf("STOP\n\n")
	// 	keepGoing = false
	// }()

	// rand.Seed(time.Now().Unix())
	// for keepGoing {
	// 	v := rand.Int63n(100)
	// 	if rand.Int()%2 == 0 {
	// 		v *= -1
	// 		metrics.GetOrRegisterGauge("foo.bar.baz", nil).Update(v)
	// 	} else {
	// 		metrics.GetOrRegisterGauge("foo.bar.baz", nil).Update(v)
	// 	}
	// 	fmt.Printf("%d\n", v)
	// 	time.Sleep(9 * time.Second)
	// }
	// fmt.Printf("END\n\n")
	// return

	// psqlDBcfg := mysql.Config{
	// 	Addr:              "127.0.0.1:6032",
	// 	Passwd:            "",
	// 	User:              "root",
	// 	Net:               "tcp",
	// 	InterpolateParams: true,
	// }
	// psqlDB, err := sql.Open("mysql", psqlDBcfg.FormatDSN())
	// if err != nil {
	// 	log.Fatalf("%s", err)
	// }
	// defer psqlDB.Close()

	// err = insertNewServer(psqlDB)
	// if err != nil {
	// 	log.Fatalf("%s\n", err)
	// }

	host := "127.0.0.1"
	port := 6033
	user := "newyork"
	pswd := "newyork"
	dbcfg := mysql.Config{
		Addr:              fmt.Sprintf("%s:%d", host, port),
		Passwd:            pswd,
		User:              user,
		Net:               "tcp",
		InterpolateParams: true,
	}
	db, err := sql.Open("mysql", dbcfg.FormatDSN())
	if err != nil {
		log.Fatalf("%s", err)
	}

	err = insertNewPerson(db)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	err = printGotham(db)
	if err != nil {
		log.Fatalf("%s\n", err)
	}

	defer db.Close()

}
