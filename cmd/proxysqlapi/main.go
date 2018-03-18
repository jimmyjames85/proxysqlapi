package main

import (
	"database/sql"
	"fmt"
	"log"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	Port   int    `envconfig:"PORT" required:"false" default:"16032"` // port to run on
	DBuser string `envconfig:"DB_USER" default:"admin"`
	DBPswd string `envconfig:"DB_PASS" default:"admin"`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

func printHeader(h string) {
	fmt.Printf("----------------------------------------\n%s\n\n", h)
}

func printGlobalVariable(db *sql.DB, name string, runtime bool) {
	var vars []admin.GlobalVariable
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

	value := fmt.Sprintf("no such global_variable: %q", name)
	for _, v := range vars {
		if strings.ToLower(v.Name) == name {
			value = v.ToJSON()
		}
	}
	fmt.Printf("%s\n", value)
}

func printGlobalVariables(db *sql.DB, runtime bool) {
	var vars []admin.GlobalVariable
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
	for _, v := range vars {
		fmt.Printf("%55s:\t%s\n", v.Name, v.Value)
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

func insertSomeMysqlServers(db *sql.DB) {
	printHeader("InsertMysqlServers")

	hostgroupID := 5

	err := admin.DropMysqlServerHostgroup(db, hostgroupID)
	if err != nil {
		log.Fatal(err)
	}

	hosts := []string{"foo", "bar"}
	var arr []*admin.MysqlServer

	for _, h := range hosts {
		s := admin.NewMysqlServer(h)
		s.Comment = "rw " + h
		s.HostgroupID = hostgroupID
		arr = append(arr, s)
	}

	err = admin.InsertMysqlServers(db, arr...)
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

	c := &config{}
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

	// insertSomeMysqlServers(db)
	// printMysqlServers(db, false)
	// printMysqlServers(db, true)

	// admin.LoadMysqlUsersToRuntime(db)
	// insertSomeMysqlUsers(db)
	// printMysqlUsers(db, false)
	// printMysqlUsers(db, true)

	someVariable := "mysql-shun_recovery_time_sec"
	err = admin.SetGlobalVariable(db, someVariable, "15")
	if err != nil {
		log.Printf(err.Error())
	}
	err = admin.LoadMysqlVariablesToRuntime(db)
	if err != nil {
		log.Printf(err.Error())
	}

	printGlobalVariable(db, someVariable, false)
	printGlobalVariable(db, someVariable, true)

}
