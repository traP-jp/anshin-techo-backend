package config

import (
	"net"
	"strconv"

	"github.com/alecthomas/kong"
	"github.com/go-sql-driver/mysql"
)

type Config struct {
	AppAddr string `env:"APP_ADDR" default:":8080"`
	DBUser  string `env:"NS_MARIADB_USER" default:"root"`
	DBPass  string `env:"NS_MARIADB_PASSWORD" default:"pass"`
	DBHost  string `env:"NS_MARIADB_HOSTNAME" default:"localhost"`
	DBPort  int    `env:"NS_MARIADB_PORT" default:"3306"`
	DBName  string `env:"NS_MARIADB_DATABASE" default:"app"`
}

func (c *Config) Parse() {
	kong.Parse(c)
}

func (c Config) MySQLConfig() *mysql.Config {
	mc := mysql.NewConfig()

	mc.User = c.DBUser
	mc.Passwd = c.DBPass
	mc.Net = "tcp"
	mc.Addr = net.JoinHostPort(c.DBHost, strconv.Itoa(c.DBPort))
	mc.DBName = c.DBName
	mc.Collation = "utf8mb4_general_ci"
	mc.AllowNativePasswords = true

	return mc
}
