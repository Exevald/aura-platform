package config

import (
	"errors"
	"time"
)

type Env struct {
	App   App   `envconfig:"app"`
	GRPC  GRPC  `envconfig:"grpc"`
	HTTP  HTTP  `envconfig:"http"`
	MySQL MySQL `envconfig:"mysql"`
}

type App struct {
	Debug        bool          `envconfig:"debug"`
	GraceTimeout time.Duration `envconfig:"grace_timeout" default:"30s"`
}

type GRPC struct {
	Address string `envconfig:"address" default:":8081"`
}

type HTTP struct {
	Address string `envconfig:"address" default:":8082"`
}

type MySQL struct {
	Address               string        `envconfig:"address" default:"localhost:3306"`
	Database              string        `envconfig:"database" default:"auditor"`
	User                  string        `envconfig:"user" default:"auditor"`
	Password              string        `envconfig:"password" default:"auditor"`
	MaxConnections        int           `envconfig:"max_connections" default:"10"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_lifetime" default:"30m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"5m"`
	ConnectTimeout        time.Duration `envconfig:"connect_timeout" default:"5s"`
	ReadTimeout           time.Duration `envconfig:"read_timeout" default:"10s"`
	WriteTimeout          time.Duration `envconfig:"write_timeout" default:"10s"`
}

func (c Env) Validate() error {
	if c.App.GraceTimeout <= 0 {
		return errors.New("grace timeout must be positive")
	}
	if c.MySQL.ConnectTimeout <= 0 || c.MySQL.ReadTimeout <= 0 || c.MySQL.WriteTimeout <= 0 || c.MySQL.ConnectionMaxLifeTime < 0 || c.MySQL.ConnectionMaxIdleTime < 0 {
		return errors.New("MySQL timeouts must be positive; connection lifetimes must not be negative")
	}
	if c.MySQL.MaxConnections <= 0 {
		return errors.New("connection count must be positive")
	}
	if c.MySQL.Address == "" || c.MySQL.Database == "" || c.MySQL.User == "" || c.HTTP.Address == "" || c.GRPC.Address == "" {
		return errors.New("database and server configuration must not be empty")
	}
	return nil
}
