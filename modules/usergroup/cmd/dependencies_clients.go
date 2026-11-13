package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userinternalapi "user/api/server/internalapi"
	"usergroup/internal/usergroup/app"
	"usergroup/internal/usergroup/config"
	clientinfra "usergroup/internal/usergroup/infra/client"
)

func newUserClient(conf config.Env) (app.UserClient, *grpc.ClientConn, error) {
	connection, err := grpc.NewClient(conf.Dependencies.UserAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return clientinfra.NewUser(userinternalapi.NewUserServiceClient(connection), conf.Dependencies.Timeout), connection, nil
}
