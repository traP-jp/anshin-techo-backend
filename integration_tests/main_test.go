package integrationtests

import (
	"fmt"
	"log"
	"net/http"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/ory/dockertest/v3"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/config"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/database"
	"github.com/traP-jp/anshin-techo-backend/infrastructure/injector"
	"github.com/traP-jp/anshin-techo-backend/internal/service/bot"
)

var globalServer http.Handler

func TestMain(m *testing.M) {
	if err := run(m); err != nil {
		log.Fatalf("runtime error: %+v", err)
	}
}

func run(m *testing.M) error {
	c := config.Config{
		DBUser: "root",
		DBPass: "pass",
		DBHost: "localhost",
		DBPort: 3306,
		DBName: "app",
	}

	pool, err := dockertest.NewPool("")
	if err != nil {
		return fmt.Errorf("connect to docker: %w", err)
	}

	if err := pool.Client.Ping(); err != nil {
		return fmt.Errorf("ping docker: %w", err)
	}

	mysqlConfig := c.MySQLConfig()

	resource, err := pool.Run("mysql", "latest", []string{
		"MYSQL_ROOT_PASSWORD=" + mysqlConfig.Passwd,
		"MYSQL_DATABASE=" + mysqlConfig.DBName,
		"MYSQL_ROOT_HOST=%",
	})
	if err != nil {
		return fmt.Errorf("start mysql docker: %w", err)
	}

	mysqlConfig.Addr = "localhost:" + resource.GetPort("3306/tcp")

	log.Println("wait for database container")

	var db *sqlx.DB
	if err := pool.Retry(func() error {
		_db, err := database.Setup(mysqlConfig)
		if err != nil {
			return err
		}

		db = _db

		return nil
	}); err != nil {
		return fmt.Errorf("connect to database container: %w", err)
	}

	mockBot := bot.NewMockService()

	server, err := injector.InjectServer(injector.Dependencies{
		DB:  db,
		Bot: mockBot,
	})
	if err != nil {
		return fmt.Errorf("inject server: %w", err)
	}

	globalServer = server

	m.Run()

	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("purge mysql docker: %w", err)
	}

	return nil
}
