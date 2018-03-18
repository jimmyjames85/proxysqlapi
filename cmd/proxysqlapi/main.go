package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/jimmyjames85/proxysqlapi/pkg/admin"
	"github.com/kelseyhightower/envconfig"
	"github.com/spf13/viper"
)

type config struct {
	Port   int    `envconfig:"PORT" required:"false" default:"16032"` // port to run on
	DBuser string `envconfig:"DB_USER" default:"root"`
	DBPswd string `envconfig:"DB_PASS" default:""`
	DBHost string `envconfig:"DB_HOST" default:"localhost"`
	DBPort int    `envconfig:"DB_PORT" default:"6032"`
}

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

func loadViperCfg(filename string) error {

	f, err := os.Stat(filename)
	if err != nil {
		return err
	}

	if f.IsDir() {
		return fmt.Errorf("%q is a directory: I was expecting a config file", filename)
	}

	viper.SetConfigFile(filename)
	viper.SetConfigType("json")
	err = viper.ReadInConfig()
	if err != nil {
		return err
	}

	fmt.Printf("%q\n", viper.GetString("mysql_servers.hostgroup_id"))

	servers := viper.GetStringSlice("mysql_servers")
	for k, v := range servers {
		fmt.Printf("%s %s\n", k, v)
	}

	return nil
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

	psqlCfg, err := admin.LoadConfig("example.json")
	if err != nil {
		log.Fatal(err)
	}

	for k, v := range psqlCfg.GlobalVariables {
		fmt.Printf("%s = %s\n", k, v)
	}

	for _, s := range psqlCfg.MysqlServers {
		fmt.Printf("%s\n", s.ToJSON())
	}

	for _, u := range psqlCfg.MysqlUsers {
		fmt.Printf("%s\n", u.ToJSON())
	}

	err = psqlCfg.LoadToRuntime(db)
	if err != nil {
		log.Fatal(err)
	}

	return
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
