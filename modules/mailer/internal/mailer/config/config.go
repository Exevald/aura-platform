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
	AMQP  AMQP  `envconfig:"amqp"`
	Relay Relay `envconfig:"relay"`
	SMTP  SMTP  `envconfig:"smtp"`
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
	Database              string        `envconfig:"database" default:"mailer"`
	User                  string        `envconfig:"user" default:"mailer"`
	Password              string        `envconfig:"password" default:"mailer"`
	MaxConnections        int           `envconfig:"max_connections" default:"10"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_lifetime" default:"30m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"5m"`
	ConnectTimeout        time.Duration `envconfig:"connect_timeout" default:"5s"`
	ReadTimeout           time.Duration `envconfig:"read_timeout" default:"10s"`
	WriteTimeout          time.Duration `envconfig:"write_timeout" default:"10s"`
}

type AMQP struct {
	URL            string        `envconfig:"url" default:"amqp://local:local@localhost:5672/"`
	Queue          string        `envconfig:"queue" default:"mailer.delivery-events"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout" default:"5s"`
	PrefetchCount  int           `envconfig:"prefetch_count" default:"10"`
}

type Relay struct {
	Interval       time.Duration `envconfig:"interval" default:"1s"`
	BatchSize      int           `envconfig:"batch_size" default:"100"`
	PublishTimeout time.Duration `envconfig:"publish_timeout" default:"10s"`
	LockTimeout    time.Duration `envconfig:"lock_timeout" default:"1s"`
}

type SMTP struct {
	Address string `envconfig:"address" default:"localhost:1025"`
	From    string `envconfig:"from" default:"Aura <noreply@aura.local>"`
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
	if c.MySQL.Address == "" || c.MySQL.Database == "" || c.MySQL.User == "" || c.AMQP.URL == "" || c.AMQP.Queue == "" || c.HTTP.Address == "" || c.GRPC.Address == "" || c.SMTP.Address == "" || c.SMTP.From == "" {
		return errors.New("database, broker and server configuration must not be empty")
	}
	return nil
}
