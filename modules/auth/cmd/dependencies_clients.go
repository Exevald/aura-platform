package main

import (
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accessinternalapi "access/api/server/internalapi"
	"auth/internal/auth/app"
	"auth/internal/auth/config"
	clientinfra "auth/internal/auth/infra/client"
	sessioninternalapi "session/api/server/internalapi"
	userinternalapi "user/api/server/internalapi"
)

type serviceClients struct {
	users       app.UserClient
	sessions    app.SessionClient
	access      app.AccessClient
	connections []*grpc.ClientConn
}

func newServiceClients(conf config.Env) (*serviceClients, error) {
	userConnection, err := grpc.NewClient(conf.Dependencies.UserAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	sessionConnection, err := grpc.NewClient(conf.Dependencies.SessionAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConnection.Close()
		return nil, err
	}
	accessConnection, err := grpc.NewClient(conf.Dependencies.AccessAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = userConnection.Close()
		_ = sessionConnection.Close()
		return nil, err
	}
	return &serviceClients{
		users: clientinfra.NewUser(userinternalapi.NewUserServiceClient(userConnection), conf.Dependencies.Timeout),
		sessions: clientinfra.NewSession(
			sessioninternalapi.NewSessionServiceClient(sessionConnection),
			conf.Dependencies.Timeout,
		),
		access:      clientinfra.NewAccess(accessinternalapi.NewAccessServiceClient(accessConnection), conf.Dependencies.Timeout, conf.Dependencies.EditorModuleID),
		connections: []*grpc.ClientConn{userConnection, sessionConnection, accessConnection},
	}, nil
}

func (c *serviceClients) Close() error {
	errs := make([]error, 0, len(c.connections))
	for _, connection := range c.connections {
		errs = append(errs, connection.Close())
	}
	return errors.Join(errs...)
}
