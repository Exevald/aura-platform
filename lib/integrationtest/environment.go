package integrationtest

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"slices"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
	"github.com/testcontainers/testcontainers-go/network"
	"github.com/testcontainers/testcontainers-go/wait"
)

type LogBuffer struct {
	mu  sync.Mutex
	buf bytes.Buffer
}

func (b *LogBuffer) Accept(log testcontainers.Log) { _, _ = b.Write(log.Content) }

func (b *LogBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.Write(p)
}

func (b *LogBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.buf.String()
}

// Environment owns test containers and closes them in reverse startup order.
type Environment struct {
	Network *testcontainers.DockerNetwork
	Logs    LogBuffer
	closers []func() error
}

func New() *Environment { return &Environment{} }

func (e *Environment) AddCloser(fn func() error) { e.closers = append(e.closers, fn) }

func (e *Environment) StartNetwork(ctx context.Context) error {
	dockerNetwork, err := network.New(ctx)
	if err != nil {
		return err
	}
	e.Network = dockerNetwork
	e.AddCloser(func() error {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return dockerNetwork.Remove(ctx)
	})
	return nil
}

func (e *Environment) StartMySQL(ctx context.Context, module, alias, password string) (string, error) {
	container, err := tcmysql.Run(ctx, "mysql:8.4",
		tcmysql.WithDatabase(module+"_test"),
		tcmysql.WithUsername(module),
		tcmysql.WithPassword(password),
		network.WithNetwork([]string{alias}, e.Network),
	)
	if container != nil {
		e.AddCloser(func() error { return testcontainers.TerminateContainer(container) })
	}
	if err != nil {
		return "", err
	}
	return container.ConnectionString(ctx, "parseTime=true")
}

func (e *Environment) StartContainer(ctx context.Context, request testcontainers.ContainerRequest) (testcontainers.Container, error) {
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: request,
		Started:          true,
	})
	if container != nil {
		e.AddCloser(func() error { return testcontainers.TerminateContainer(container) })
	}
	return container, err
}

func (e *Environment) StartBroker(ctx context.Context, user, password string) (string, error) {
	broker, err := e.StartContainer(ctx, testcontainers.ContainerRequest{
		Image:        "rabbitmq:4.1-alpine",
		ExposedPorts: []string{"5672/tcp"},
		Networks:     []string{e.Network.Name},
		NetworkAliases: map[string][]string{
			e.Network.Name: {"rabbitmq"},
		},
		Env: map[string]string{
			"RABBITMQ_DEFAULT_USER": user,
			"RABBITMQ_DEFAULT_PASS": password,
		},
		WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(time.Minute),
	})
	if err != nil {
		return "", err
	}
	address, err := Address(ctx, broker, "5672/tcp")
	if err != nil {
		return "", err
	}
	return (&url.URL{Scheme: "amqp", User: url.UserPassword(user, password), Host: address, Path: "/"}).String(), nil
}

func (e *Environment) StartService(ctx context.Context, image, command string, variables map[string]string) (testcontainers.Container, error) {
	return e.StartContainer(ctx, testcontainers.ContainerRequest{
		Image:        image,
		Cmd:          []string{command},
		ExposedPorts: []string{"8081/tcp", "8082/tcp"},
		Networks:     []string{e.Network.Name},
		Env:          variables,
		LogConsumerCfg: &testcontainers.LogConsumerConfig{
			Consumers: []testcontainers.LogConsumer{&e.Logs},
		},
		WaitingFor: wait.ForHTTP("/healthz").WithPort("8082/tcp").WithStatusCodeMatcher(func(status int) bool {
			return status == http.StatusNoContent
		}).WithStartupTimeout(time.Minute),
	})
}

func Address(ctx context.Context, container testcontainers.Container, port string) (string, error) {
	host, err := container.Host(ctx)
	if err != nil {
		return "", err
	}
	mappedPort, err := container.MappedPort(ctx, port)
	if err != nil {
		return "", err
	}
	return net.JoinHostPort(host, mappedPort.Port()), nil
}

func (e *Environment) Close() error {
	var errs []error
	for _, v := range slices.Backward(e.closers) {
		errs = append(errs, v())
	}
	return errors.Join(errs...)
}

func Run(m *testing.M, env *Environment, prepare func(context.Context) error) (exitCode int) {
	defer func() {
		if err := env.Close(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			exitCode = 1
		}
		if exitCode != 0 {
			fmt.Fprintln(os.Stderr, env.Logs.String())
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	if err := prepare(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return m.Run()
}
