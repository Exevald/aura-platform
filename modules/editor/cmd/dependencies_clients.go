package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	userinternalapi "user/api/server/internalapi"

	"editor/internal/editor/app"
	"editor/internal/editor/config"
	clientinfra "editor/internal/editor/infra/client"
)

func newUserClient(conf config.Env) (app.UserClient, *grpc.ClientConn, error) {
	c, err := grpc.NewClient(conf.Dependencies.UserAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return clientinfra.NewUser(userinternalapi.NewUserServiceClient(c), conf.Dependencies.Timeout), c, nil
}
