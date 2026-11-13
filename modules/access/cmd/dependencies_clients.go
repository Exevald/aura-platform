package main

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"access/internal/access/config"
	clientinfra "access/internal/access/infra/client"
	appregapi "appreg/api/server/publicapi"
	usergroupapi "usergroup/api/server/internalapi"
)

func newUserGroupClient(conf config.Env) (clientinfra.UserGroup, *grpc.ClientConn, error) {
	connection, err := grpc.NewClient(conf.Dependencies.UserGroupAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, err
	}
	return clientinfra.NewUserGroup(usergroupapi.NewUsergroupServiceClient(connection), conf.Dependencies.Timeout), connection, nil
}

func newAppRegistryClient(conf config.Env) (clientinfra.AppRegistry, error) {
	client, err := appregapi.NewClientWithResponses("http://" + conf.Dependencies.AppRegAddress)
	if err != nil {
		return nil, err
	}
	return clientinfra.NewAppRegistry(client, conf.Dependencies.Timeout), nil
}
