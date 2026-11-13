package config

import (
	"errors"
	"time"
)

type Env struct {
	App          App          `envconfig:"app"`
	GRPC         GRPC         `envconfig:"grpc"`
	HTTP         HTTP         `envconfig:"http"`
	MySQL        MySQL        `envconfig:"mysql"`
	AMQP         AMQP         `envconfig:"amqp"`
	Relay        Relay        `envconfig:"relay"`
	Dependencies Dependencies `envconfig:"dependencies"`
	Recovery     Recovery     `envconfig:"recovery"`
}

type Recovery struct {
	Interval       time.Duration `envconfig:"interval" default:"1s"`
	AttemptTimeout time.Duration `envconfig:"attempt_timeout" default:"5s"`
	MaxAttempts    int           `envconfig:"max_attempts" default:"3"`
}

type Dependencies struct {
	UserAddress        string        `envconfig:"user_address" default:"user-service:8081"`
	FilestorageAddress string        `envconfig:"filestorage_address" default:"filestorage-service:8081"`
	ModuleID           string        `envconfig:"module_id" default:"b8547c76-51fd-4f7c-9c68-8f30c84f61cb"`
	Timeout            time.Duration `envconfig:"timeout" default:"5s"`
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
	Database              string        `envconfig:"database" default:"editor"`
	User                  string        `envconfig:"user" default:"editor"`
	Password              string        `envconfig:"password" default:"editor"`
	MaxConnections        int           `envconfig:"max_connections" default:"10"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_lifetime" default:"30m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"5m"`
	ConnectTimeout        time.Duration `envconfig:"connect_timeout" default:"5s"`
	ReadTimeout           time.Duration `envconfig:"read_timeout" default:"10s"`
	WriteTimeout          time.Duration `envconfig:"write_timeout" default:"10s"`
}

type AMQP struct {
	URL            string        `envconfig:"url" default:"amqp://local:local@localhost:5672/"`
	Queue          string        `envconfig:"queue" default:"editor.document-events"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout" default:"5s"`
	PrefetchCount  int           `envconfig:"prefetch_count" default:"10"`
}

type Relay struct {
	Interval       time.Duration `envconfig:"interval" default:"1s"`
	BatchSize      int           `envconfig:"batch_size" default:"100"`
	PublishTimeout time.Duration `envconfig:"publish_timeout" default:"10s"`
	LockTimeout    time.Duration `envconfig:"lock_timeout" default:"1s"`
}

func (c Env) Validate() error {
	if c.App.GraceTimeout <= 0 || c.AMQP.ConnectTimeout <= 0 || c.Relay.Interval <= 0 || c.Relay.PublishTimeout <= 0 || c.Relay.LockTimeout < 0 {
		return errors.New("timeouts and relay interval must be positive; lock timeout must not be negative")
	}
	if c.MySQL.ConnectTimeout <= 0 || c.MySQL.ReadTimeout <= 0 || c.MySQL.WriteTimeout <= 0 || c.MySQL.ConnectionMaxLifeTime < 0 || c.MySQL.ConnectionMaxIdleTime < 0 {
		return errors.New("MySQL timeouts must be positive; connection lifetimes must not be negative")
	}
	if c.MySQL.MaxConnections <= 0 || c.AMQP.PrefetchCount <= 0 || c.Relay.BatchSize <= 0 {
		return errors.New("connection, prefetch and batch counts must be positive")
	}
	if c.MySQL.Address == "" || c.MySQL.Database == "" || c.MySQL.User == "" || c.AMQP.URL == "" || c.AMQP.Queue == "" || c.HTTP.Address == "" || c.GRPC.Address == "" {
		return errors.New("database, broker and server configuration must not be empty")
	}
	if c.Dependencies.UserAddress == "" || c.Dependencies.FilestorageAddress == "" || c.Dependencies.ModuleID == "" || c.Dependencies.Timeout <= 0 {
		return errors.New("user dependency is invalid")
	}
	if c.Recovery.Interval <= 0 || c.Recovery.AttemptTimeout <= 0 || c.Recovery.MaxAttempts <= 0 {
		return errors.New("share recovery configuration is invalid")
	}
	return nil
}
