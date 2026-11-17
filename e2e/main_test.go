package e2e

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/distributed-programming-2026/go-sdk/pkg/integrationtests"
	"github.com/distributed-programming-2026/go-sdk/pkg/integrationtests/mockgrpc"
	_ "github.com/go-sql-driver/mysql"
	rabbit "github.com/rabbitmq/amqp091-go"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accessapi "access/api/server/internalapi"
	accesspublicapi "access/api/server/publicapi"
	authapi "auth/api/server/internalapi"
	registererapi "registerer/api/server/publicapi"
	sessionapi "session/api/server/internalapi"
	userapi "user/api/server/internalapi"
	usergroupapi "usergroup/api/server/publicapi"
)

const (
	databasePassword  = "integration:password@local"
	grpcContainerPort = "8081/tcp"
	httpContainerPort = "8082/tcp"
	serviceCommand    = "service"
)

const (
	accessModule     = "access"
	authModule       = "auth"
	registererModule = "registerer"
	rabbitmqName     = "rabbitmq"
	mysqlName        = "mysql"
	sessionModule    = "session"
	userModule       = "user"
	usergroupModule  = "usergroup"
)

var (
	env                *integrationtests.Environment
	prepareFanOutGate  = newFanOutGate(2)
	commitFanOutGate   = newFanOutGate(2)
	rollbackFanOutGate = newFanOutGate(2)
	authProxy          = &authRegistrationProxy{prepareGate: prepareFanOutGate, commitGate: commitFanOutGate, rollbackGate: rollbackFanOutGate}
	userProxy          = &userRegistrationProxy{prepareGate: prepareFanOutGate, commitGate: commitFanOutGate, rollbackGate: rollbackFanOutGate}
	accessProxy        = &accessRegistrationProxy{prepareGate: prepareFanOutGate, commitGate: commitFanOutGate, rollbackGate: rollbackFanOutGate}
	sessions           = &sessionProxy{}
	deps               struct {
		client     *registererapi.ClientWithResponses
		access     *accesspublicapi.ClientWithResponses
		groups     *usergroupapi.ClientWithResponses
		accessGRPC accessapi.ModuleAccessServiceClient
		dbs        map[string]*sql.DB
		events     <-chan rabbit.Delivery
	}
	registererServiceContainer testcontainers.Container
)

func TestMain(m *testing.M) {
	env = integrationtests.New(integrationtests.Config{
		Servers: []integrationtests.ServerSpec{
			mockgrpc.New("auth-proxy", func(server *grpc.Server) { authapi.RegisterAuthServiceServer(server, authProxy) }),
			mockgrpc.New("user-proxy", func(server *grpc.Server) { userapi.RegisterUserServiceServer(server, userProxy) }),
			mockgrpc.New("access-proxy", func(server *grpc.Server) { accessapi.RegisterModuleAccessServiceServer(server, accessProxy) }),
			mockgrpc.New("session-proxy", func(server *grpc.Server) { sessionapi.RegisterSessionServiceServer(server, sessions) }),
		},
		Containers: append(infrastructureSpecs(), serviceSpecs()...),
		Timeouts:   integrationtests.Timeouts{Prepare: 5 * time.Minute, Up: 3 * time.Minute, Tests: 5 * time.Minute, Down: time.Minute},
		Setup:      setupClients,
	})
	os.Exit(integrationtests.Run(m, env))
}

func infrastructureSpecs() []integrationtests.ContainerSpec {
	result := make([]integrationtests.ContainerSpec, 0, 2)
	result = append(result, integrationtests.ContainerSpec{
		Name: rabbitmqName,
		Request: testcontainers.ContainerRequest{
			Image: "rabbitmq:4.1-alpine", ExposedPorts: []string{"5672/tcp"},
			Env:        map[string]string{"RABBITMQ_DEFAULT_USER": "integration", "RABBITMQ_DEFAULT_PASS": "integration-password"},
			WaitingFor: wait.ForLog("Server startup complete").WithStartupTimeout(time.Minute),
		},
	})
	result = append(result, integrationtests.ContainerSpec{
		Name: mysqlName,
		Request: testcontainers.ContainerRequest{
			Image: "mysql:8.4", ExposedPorts: []string{"3306/tcp"},
			Env:        map[string]string{"MYSQL_ROOT_PASSWORD": "root-password"},
			Files:      []testcontainers.ContainerFile{{HostFilePath: "docker/mysql/init.sql", ContainerFilePath: "/docker-entrypoint-initdb.d/init.sql", FileMode: 0o644}},
			WaitingFor: wait.ForLog("port: 3306  MySQL Community Server").WithStartupTimeout(time.Minute),
		},
	})
	return result
}

func serviceSpecs() []integrationtests.ContainerSpec {
	var result []integrationtests.ContainerSpec
	for _, module := range []string{userModule, accessModule, sessionModule} {
		dependsOn := []string{mysqlName, rabbitmqName}
		if module == accessModule {
			dependsOn = append(dependsOn, "usergroup-service")
		}
		result = append(result, integrationtests.ContainerSpec{
			Name: module + "-service", DependsOn: dependsOn,
			Request: testcontainers.ContainerRequest{Image: "distributed-programming-" + module + ":e2e", Cmd: []string{serviceCommand}, ExposedPorts: []string{grpcContainerPort, httpContainerPort}},
			Configure: func(_ context.Context, testEnv *integrationtests.Environment, request *testcontainers.ContainerRequest) error {
				request.Env = moduleEnvironment(module, testEnv)
				if module == accessModule {
					request.Env["ACCESS_DEPENDENCIES_USERGROUP_ADDRESS"], _ = testEnv.NetworkEndpoint("usergroup-service", grpcContainerPort)
				}
				request.WaitingFor = healthWait()
				return nil
			},
			AfterReady: func(ctx context.Context, testEnv *integrationtests.Environment, _ testcontainers.Container) error {
				address, err := testEnv.HostEndpoint(ctx, module+"-service", grpcContainerPort)
				if err != nil {
					return err
				}
				connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
				if err != nil {
					return err
				}
				if err = testEnv.AddCleanup(module+" proxy client", func(context.Context) error { return connection.Close() }); err != nil {
					return err
				}
				switch module {
				case userModule:
					userProxy.setClient(userapi.NewUserServiceClient(connection))
				case accessModule:
					deps.accessGRPC = accessapi.NewModuleAccessServiceClient(connection)
					accessProxy.setClient(deps.accessGRPC)
				default:
					sessions.setClient(sessionapi.NewSessionServiceClient(connection))
				}
				return nil
			},
		})
	}
	result = append(result, integrationtests.ContainerSpec{
		Name: "usergroup-service", DependsOn: []string{mysqlName, rabbitmqName, "user-service"},
		Request: testcontainers.ContainerRequest{Image: "distributed-programming-usergroup:e2e", Cmd: []string{serviceCommand}, ExposedPorts: []string{grpcContainerPort, httpContainerPort}},
		Configure: func(_ context.Context, testEnv *integrationtests.Environment, request *testcontainers.ContainerRequest) error {
			request.Env = moduleEnvironment(usergroupModule, testEnv)
			request.Env["USERGROUP_DEPENDENCIES_USER_ADDRESS"], _ = testEnv.NetworkEndpoint("user-service", grpcContainerPort)
			request.WaitingFor = healthWait()
			return nil
		},
	})
	result = append(result, integrationtests.ContainerSpec{
		Name: "access-message-handler", DependsOn: []string{mysqlName, rabbitmqName, "access-service"},
		Request: testcontainers.ContainerRequest{Image: "distributed-programming-access:e2e", Cmd: []string{"message-handler"}, ExposedPorts: []string{httpContainerPort}},
		Configure: func(_ context.Context, testEnv *integrationtests.Environment, request *testcontainers.ContainerRequest) error {
			request.Env = moduleEnvironment(accessModule, testEnv)
			request.WaitingFor = healthWait()
			return nil
		},
	})
	result = append(result, integrationtests.ContainerSpec{
		Name: "auth-service", DependsOn: []string{mysqlName, rabbitmqName, "session-service"},
		Request: testcontainers.ContainerRequest{Image: "distributed-programming-auth:e2e", Cmd: []string{serviceCommand}, ExposedPorts: []string{grpcContainerPort, httpContainerPort}},
		Configure: func(_ context.Context, testEnv *integrationtests.Environment, request *testcontainers.ContainerRequest) error {
			request.Env = moduleEnvironment(authModule, testEnv)
			request.Env["AUTH_DEPENDENCIES_SESSION_ADDRESS"], _ = testEnv.NetworkEndpoint("session-service", grpcContainerPort)
			request.Env["AUTH_COOKIE_SECURE"] = "false"
			request.WaitingFor = healthWait()
			return nil
		},
		AfterReady: func(ctx context.Context, testEnv *integrationtests.Environment, _ testcontainers.Container) error {
			address, err := testEnv.HostEndpoint(ctx, "auth-service", grpcContainerPort)
			if err != nil {
				return err
			}
			connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				return err
			}
			if err = testEnv.AddCleanup("auth proxy client", func(context.Context) error { return connection.Close() }); err != nil {
				return err
			}
			authProxy.setClient(authapi.NewAuthServiceClient(connection))
			return nil
		},
	})
	result = append(result, integrationtests.ContainerSpec{
		Name:      "registerer-service",
		DependsOn: []string{mysqlName, rabbitmqName, "auth-service", "user-service", "access-service", "session-service", "auth-proxy", "user-proxy", "access-proxy", "session-proxy"},
		Request: testcontainers.ContainerRequest{
			Image: "alpine:3.22", Entrypoint: []string{"/registerer-supervisor.sh"}, ExposedPorts: []string{httpContainerPort},
			Files: []testcontainers.ContainerFile{
				{HostFilePath: "../modules/registerer/dist/registerer", ContainerFilePath: "/registerer", FileMode: 0o755},
				{HostFilePath: "registerer-supervisor.sh", ContainerFilePath: "/registerer-supervisor.sh", FileMode: 0o755},
			},
		},
		Configure: func(_ context.Context, testEnv *integrationtests.Environment, request *testcontainers.ContainerRequest) error {
			request.Env = moduleEnvironment(registererModule, testEnv)
			request.Env["REGISTERER_DEPENDENCIES_AUTH_ADDRESS"], _ = testEnv.NetworkEndpoint("auth-proxy")
			request.Env["REGISTERER_DEPENDENCIES_USER_ADDRESS"], _ = testEnv.NetworkEndpoint("user-proxy")
			request.Env["REGISTERER_DEPENDENCIES_ACCESS_ADDRESS"], _ = testEnv.NetworkEndpoint("access-proxy")
			request.Env["REGISTERER_DEPENDENCIES_SESSION_ADDRESS"], _ = testEnv.NetworkEndpoint("session-proxy")
			request.Env["REGISTERER_COOKIE_SECURE"] = "false"
			request.Env["REGISTERER_RECOVERY_INTERVAL"] = "50ms"
			request.WaitingFor = healthWait()
			return nil
		},
		AfterReady: func(_ context.Context, _ *integrationtests.Environment, container testcontainers.Container) error {
			registererServiceContainer = container
			return nil
		},
	})
	return result
}

func moduleEnvironment(module string, testEnv *integrationtests.Environment) map[string]string {
	prefix := map[string]string{authModule: "AUTH_", registererModule: "REGISTERER_", userModule: "USER_", accessModule: "ACCESS_", sessionModule: "SESSION_", usergroupModule: "USERGROUP_"}[module]
	databaseAddress, _ := testEnv.NetworkEndpoint(mysqlName, "3306/tcp")
	brokerAddress, _ := testEnv.NetworkEndpoint(rabbitmqName, "5672/tcp")
	return map[string]string{
		prefix + "MYSQL_ADDRESS": databaseAddress, prefix + "MYSQL_DATABASE": module + "_test", prefix + "MYSQL_USER": module, prefix + "MYSQL_PASSWORD": databasePassword,
		prefix + "AMQP_URL": "amqp://integration:integration-password@" + brokerAddress + "/", prefix + "HTTP_ADDRESS": ":8082", prefix + "GRPC_ADDRESS": ":8081",
		prefix + "RELAY_INTERVAL": "50ms", prefix + "RELAY_PUBLISH_TIMEOUT": "1s", prefix + "APP_GRACE_TIMEOUT": "1s",
	}
}

func healthWait() wait.Strategy {
	return wait.ForHTTP("/healthz").WithPort(httpContainerPort).WithStatusCodeMatcher(func(status int) bool { return status == http.StatusNoContent }).WithStartupTimeout(time.Minute)
}

func setupClients(ctx context.Context, testEnv *integrationtests.Environment) error {
	address, err := testEnv.HostEndpoint(ctx, "registerer-service", httpContainerPort)
	if err != nil {
		return err
	}
	deps.client, err = registererapi.NewClientWithResponses("http://" + address)
	if err != nil {
		return err
	}
	accessAddress, err := testEnv.HostEndpoint(ctx, "access-service", httpContainerPort)
	if err != nil {
		return err
	}
	deps.access, err = accesspublicapi.NewClientWithResponses("http://" + accessAddress)
	if err != nil {
		return err
	}
	groupAddress, err := testEnv.HostEndpoint(ctx, "usergroup-service", httpContainerPort)
	if err != nil {
		return err
	}
	deps.groups, err = usergroupapi.NewClientWithResponses("http://" + groupAddress)
	if err != nil {
		return err
	}
	deps.dbs = make(map[string]*sql.DB)
	for _, module := range []string{authModule, userModule, accessModule, sessionModule, registererModule, usergroupModule} {
		endpoint, endpointErr := testEnv.HostEndpoint(ctx, mysqlName, "3306/tcp")
		if endpointErr != nil {
			return endpointErr
		}
		database, openErr := sql.Open("mysql", fmt.Sprintf("%s:%s@tcp(%s)/%s_test?parseTime=true", module, databasePassword, endpoint, module))
		if openErr != nil {
			return openErr
		}
		if addErr := testEnv.AddCleanup(module+" database", func(context.Context) error { return database.Close() }); addErr != nil {
			return addErr
		}
		if pingErr := database.PingContext(ctx); pingErr != nil {
			return pingErr
		}
		deps.dbs[module] = database
	}
	brokerEndpoint, err := testEnv.HostEndpoint(ctx, rabbitmqName, "5672/tcp")
	if err != nil {
		return err
	}
	brokerURL := url.URL{Scheme: "amqp", User: url.UserPassword("integration", "integration-password"), Host: brokerEndpoint, Path: "/"}
	connection, err := rabbit.Dial(brokerURL.String())
	if err != nil {
		return err
	}
	if err = testEnv.AddCleanup("rabbit connection", func(context.Context) error { return connection.Close() }); err != nil {
		return err
	}
	channel, err := connection.Channel()
	if err != nil {
		return err
	}
	if err = testEnv.AddCleanup("rabbit channel", func(context.Context) error { return channel.Close() }); err != nil {
		return err
	}
	queue, err := channel.QueueDeclare("", false, true, true, false, nil)
	if err != nil {
		return err
	}
	if err = channel.QueueBind(queue.Name, "#", "domain-event", false, nil); err != nil {
		return err
	}
	deps.events, err = channel.Consume(queue.Name, "", true, true, false, false, nil)
	return err
}
