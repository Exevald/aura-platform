package main

import (
	"errors"
	"uuid"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	accessinternalapi "access/api/server/internalapi"
	authinternalapi "auth/api/server/internalapi"
	"registerer/internal/registerer/app"
	"registerer/internal/registerer/config"
	clientinfra "registerer/internal/registerer/infra/client"
	sessioninternalapi "session/api/server/internalapi"
	userinternalapi "user/api/server/internalapi"
	workspaceinternalapi "workspace/api/server/internalapi"
)

type serviceClients struct {
	auth        app.AuthAdapter
	users       app.UserAdapter
	sessions    app.SessionAdapter
	workspaces  app.WorkspaceAdapter
	access      app.AccessAdapter
	connections []*grpc.ClientConn
}

func (c *serviceClients) Close() error {
	result := make([]error, 0, len(c.connections))
	for _, connection := range c.connections {
		result = append(result, connection.Close())
	}
	return errors.Join(result...)
}

func newServiceClients(conf config.Env) (*serviceClients, error) {
	addresses := []string{conf.Dependencies.AuthAddress, conf.Dependencies.UserAddress, conf.Dependencies.SessionAddress, conf.Dependencies.WorkspaceAddress, conf.Dependencies.AccessAddress}
	connections := make([]*grpc.ClientConn, 0, len(addresses))
	for _, address := range addresses {
		connection, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			for _, opened := range connections {
				_ = opened.Close()
			}
			return nil, err
		}
		connections = append(connections, connection)
	}
	return &serviceClients{
		auth:        clientinfra.NewAuth(authinternalapi.NewAuthServiceClient(connections[0]), conf.Dependencies.Timeout),
		users:       clientinfra.NewUser(userinternalapi.NewUserServiceClient(connections[1]), conf.Dependencies.Timeout),
		sessions:    clientinfra.NewSession(sessioninternalapi.NewSessionServiceClient(connections[2]), conf.Dependencies.Timeout),
		workspaces:  clientinfra.NewWorkspace(workspaceinternalapi.NewWorkspaceServiceClient(connections[3]), conf.Dependencies.Timeout),
		access:      clientinfra.NewAccess(accessinternalapi.NewModuleAccessServiceClient(connections[4]), conf.Dependencies.Timeout, uuid.MustParse(conf.Dependencies.DefaultModuleID)),
		connections: connections,
	}, nil
}
