package main

import (
	"database/sql"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"time"

	_ "github.com/cyberdelia/go-metrics-graphite"
	graphite "github.com/cyberdelia/go-metrics-graphite"
	"github.com/go-sql-driver/mysql"
	metrics "github.com/rcrowley/go-metrics"
)

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

	hostname, _ := os.Hostname()

	ghost := "127.0.0.1"
	gport := 2003
	// todo s/tcp/udp
	gaddr := fmt.Sprintf("%s:%d", ghost, gport)
	fmt.Printf("%s\n", gaddr)
	addr, err := net.ResolveTCPAddr("tcp", gaddr)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	fmt.Printf("no err: %s\n", gaddr)
	go graphite.Graphite(metrics.DefaultRegistry, (10 * time.Second), fmt.Sprintf("psql.%s", hostname), addr)

	// registry := metrics.NewPrefixedRegistry()

	keepGoing := true
	go func() {
		time.Sleep(time.Minute)
		fmt.Printf("STOP\n\n")
		keepGoing = false
	}()

	rand.Seed(time.Now().Unix())
	for keepGoing {
		v := rand.Int63n(100)
		if rand.Int()%2 == 0 {
			v *= -1
			metrics.GetOrRegisterGauge("foo.bar.baz", nil).Update(v)
		} else {
			metrics.GetOrRegisterGauge("foo.bar.baz", nil).Update(v)
		}
		fmt.Printf("%d\n", v)
		time.Sleep(9 * time.Second)
	}
	fmt.Printf("END\n\n")
	return

	psqlDBcfg := mysql.Config{
		Addr:              "127.0.0.1:6032",
		Passwd:            "",
		User:              "root",
		Net:               "tcp",
		InterpolateParams: true,
	}
	psqlDB, err := sql.Open("mysql", psqlDBcfg.FormatDSN())
	if err != nil {
		log.Fatalf("%s", err)
	}
	defer psqlDB.Close()

	err = insertNewServer(psqlDB)
	if err != nil {
		log.Fatalf("%s\n", err)
	}
	return

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
