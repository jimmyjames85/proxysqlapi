package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
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

func insertSomeMysqlServers(db *sql.DB, servers []*admin.MysqlServer) {
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
	var arr []*admin.MysqlUser

	for _, n := range names {
		u := admin.NewMysqlUser(n)
		p := strings.Repeat(n, 2)
		u.Password = &p
		arr = append(arr, u)
	}

	err = admin.InsertMysqlUsers(db, arr...)
	if err != nil {
		fmt.Printf("%s", err)
	}
	fmt.Printf("success\n")
}

func main() {

	c := server.Config{}
	envconfig.MustProcess("PROXYSQLAPI", c)

	dbcfg := mysql.Config{
		Addr:              fmt.Sprintf("%s:%d", c.DBHost, c.DBPort),
		Passwd:            c.DBPswd,
		User:              c.DBuser,
		Net:               "tcp",
		InterpolateParams: true,
	}

	db, err := sql.Open("mysql", dbcfg.FormatDSN())
	if err != nil {
		log.Fatal(err)

	}
	defer db.Close()

	psqlCfg, err := admin.LoadConfig("example.json")
	if err != nil {
		log.Fatalf("err loading config: %v", err)
	}

	err = psqlCfg.LoadToMemory(db)
	if err != nil {
		log.Fatal(err)
	}

}
