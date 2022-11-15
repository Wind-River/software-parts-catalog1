// go get github.com/pressly/goose@master
//
//go:generate go build -o goose *.go
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	// _ "wrs/highlander/migrate/goose/migrations/highlander"

	_ "wrs/tkdb/goose/migrations/tkdb"

	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
)

var (
	flags = flag.NewFlagSet("goose", flag.ExitOnError)
	dir   = flags.String("dir", ".", "directory with migration files")

	driver = flags.String("driver", "pgx", "database driver")
	user   = flags.String("user", "", "Database username")
	pass   = flags.String("pass", "", "Database pass")
	host   = flags.String("host", "localhost", "Database host")
	port   = flags.Int64("port", 5432, "Database port")
	dbname = flags.String("dbname", "postgres", "Database username")
)

type sqlConnectionInfo struct {
	driver string
	user   string
	pass   string
	host   string
	port   int64
	dbname string
}

func (info sqlConnectionInfo) Driver() string {
	switch info.driver {
	case "postgres":
		return "pgx"
	default:
		return info.driver
	}
}

func (info sqlConnectionInfo) String() string {
	connectionString := fmt.Sprintf("host='%s' dbname='%s' sslmode=disable port='%d'", info.host, info.dbname, info.port)
	if info.user != "" {
		connectionString = fmt.Sprintf("%s user='%s'", connectionString, info.user)
	}
	if info.pass != "" {
		connectionString = fmt.Sprintf("%s pass='%s'", connectionString, info.pass)
	}

	return connectionString
}

func main() {
	log.Printf("os.Args: %v\n", os.Args)
	flags.Parse(os.Args[1:])

	if flags.NArg() < 1 {
		flags.Usage()
		return
	}

	dbstring := sqlConnectionInfo{*driver, *user, *pass, *host, *port, *dbname}.String()
	command := flags.Arg(0)

	db, err := goose.OpenDBWithDriver(*driver, dbstring)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v\n", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v\n", err)
		}
	}()

	arguments := []string{}
	if flags.NArg() > 1 {
		arguments = append(arguments, flags.Args()[1:]...)
	}

	log.Printf("goose.Run(%s, %s, %s, %v)\n", command, dbstring, *dir, arguments)
	if err := goose.Run(command, db, *dir, arguments...); err != nil {
		log.Fatalf("goose %v: %v", command, err)
	}
}
