package main

import (
	"log"
	"net/http"

	"github.com/ras0q/goalie"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/config"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/database"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/injector"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run() (err error) {
	g := goalie.New()
	defer g.Collect(&err)

	var c config.Config
	c.Parse()

	// connect to and migrate database
	db, err := database.Setup(c.MySQLConfig())
	if err != nil {
		return err
	}
	defer g.Guard(db.Close)

	server, err := injector.InjectServer(db)
	if err != nil {
		return err
	}

	if err := http.ListenAndServe(c.AppAddr, server); err != nil {
		return err
	}

	return nil
}
