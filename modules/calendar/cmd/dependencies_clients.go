package main

import (
	"calendar/internal/calendar/app"
	"calendar/internal/calendar/config"
	clientinfra "calendar/internal/calendar/infra/client"
	api "notification/api/server/internalapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type serviceClients struct {
	notifications app.NotificationSender
	connection    *grpc.ClientConn
}

func newServiceClients(c config.Env) (*serviceClients, error) {
	conn, err := grpc.NewClient(c.Dependencies.NotificationAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &serviceClients{notifications: clientinfra.NewNotification(api.NewNotificationServiceClient(conn), c.Dependencies.Timeout), connection: conn}, nil
}

func (c *serviceClients) Close() error {
	return c.connection.Close()
}
