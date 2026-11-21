package config

import (
	"errors"
	"time"
)

type Env struct {
	App          App          `envconfig:"app"`
	HTTP         HTTP         `envconfig:"http"`
	MySQL        MySQL        `envconfig:"mysql"`
	AMQP         AMQP         `envconfig:"amqp"`
	Temporal     Temporal     `envconfig:"temporal"`
	Dependencies Dependencies `envconfig:"dependencies"`
}

type App struct {
	Debug        bool          `envconfig:"debug"`
	GraceTimeout time.Duration `envconfig:"grace_timeout" default:"30s"`
}

type HTTP struct {
	Address string `envconfig:"address" default:":8082"`
}

type MySQL struct {
	Address               string        `envconfig:"address" default:"localhost:3306"`
	Database              string        `envconfig:"database" default:"publication"`
	User                  string        `envconfig:"user" default:"publication"`
	Password              string        `envconfig:"password" default:"publication"`
	MaxConnections        int           `envconfig:"max_connections" default:"10"`
	ConnectionMaxLifeTime time.Duration `envconfig:"connection_max_lifetime" default:"30m"`
	ConnectionMaxIdleTime time.Duration `envconfig:"connection_max_idle_time" default:"5m"`
	ConnectTimeout        time.Duration `envconfig:"connect_timeout" default:"5s"`
	ReadTimeout           time.Duration `envconfig:"read_timeout" default:"10s"`
	WriteTimeout          time.Duration `envconfig:"write_timeout" default:"10s"`
}

type AMQP struct {
	URL            string        `envconfig:"url" default:"amqp://local:local@localhost:5672/"`
	Queue          string        `envconfig:"queue" default:"publication.events"`
	ConnectTimeout time.Duration `envconfig:"connect_timeout" default:"5s"`
	PrefetchCount  int           `envconfig:"prefetch_count" default:"10"`
}

type Temporal struct {
	Address   string `envconfig:"address" default:"localhost:7233"`
	Namespace string `envconfig:"namespace" default:"default"`
}
type Dependencies struct {
	EditorAddress       string `envconfig:"editor_address" default:"localhost:8081"`
	ShortlinkAddress    string `envconfig:"shortlink_address" default:"localhost:8081"`
	WorkspaceAddress    string `envconfig:"workspace_address" default:"localhost:8081"`
	NotificationAddress string `envconfig:"notification_address" default:"localhost:8081"`
	FilestorageAddress  string `envconfig:"filestorage_address" default:"localhost:8081"`
	PublicBaseURL       string `envconfig:"public_base_url" default:"http://aura-platform.home"`
}

func (c Env) Validate() error {
	if c.App.GraceTimeout <= 0 || c.AMQP.ConnectTimeout <= 0 {
		return errors.New("timeouts must be positive")
	}
	if c.MySQL.ConnectTimeout <= 0 || c.MySQL.ReadTimeout <= 0 || c.MySQL.WriteTimeout <= 0 || c.MySQL.ConnectionMaxLifeTime < 0 || c.MySQL.ConnectionMaxIdleTime < 0 {
		return errors.New("MySQL timeouts must be positive; connection lifetimes must not be negative")
	}
	if c.MySQL.MaxConnections <= 0 || c.AMQP.PrefetchCount <= 0 {
		return errors.New("connection, prefetch and batch counts must be positive")
	}
	if c.MySQL.Address == "" || c.MySQL.Database == "" || c.MySQL.User == "" || c.AMQP.URL == "" || c.AMQP.Queue == "" || c.HTTP.Address == "" {
		return errors.New("database, broker and server configuration must not be empty")
	}
	return c.validateIntegrations()
}

func (c Env) validateIntegrations() error {
	if c.Temporal.Address == "" || c.Temporal.Namespace == "" {
		return errors.New("Temporal configuration must not be empty")
	}
	if c.Dependencies.EditorAddress == "" || c.Dependencies.ShortlinkAddress == "" || c.Dependencies.WorkspaceAddress == "" || c.Dependencies.NotificationAddress == "" || c.Dependencies.FilestorageAddress == "" || c.Dependencies.PublicBaseURL == "" {
		return errors.New("publication dependencies must not be empty")
	}
	return nil
}
