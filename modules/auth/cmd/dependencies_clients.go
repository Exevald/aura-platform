package main

import (
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"auth/internal/auth/app"
	"auth/internal/auth/config"
	clientinfra "auth/internal/auth/infra/client"
	sessioninternalapi "session/api/server/internalapi"
)

type serviceClients struct {
	sessions    app.SessionClient
	connections []*grpc.ClientConn
}

func (c *serviceClients) Close() error {
	errs := make([]error, 0, len(c.connections))
	for _, connection := range c.connections {
		errs = append(errs, connection.Close())
	}
	return errors.Join(errs...)
}

func newServiceClients(conf config.Env) (*serviceClients, error) {
	sessionConnection, err := grpc.NewClient(conf.Dependencies.SessionAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &serviceClients{
		sessions: clientinfra.NewSession(
			sessioninternalapi.NewSessionServiceClient(sessionConnection),
			conf.Dependencies.Timeout,
		),
		connections: []*grpc.ClientConn{sessionConnection},
	}, nil
}
