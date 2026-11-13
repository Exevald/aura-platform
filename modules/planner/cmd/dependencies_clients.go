package main

import (
	api "calendar/api/server/internalapi"
	"planner/internal/planner/app"
	"planner/internal/planner/config"
	clientinfra "planner/internal/planner/infra/client"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type serviceClients struct {
	calendar   app.CalendarAdapter
	connection *grpc.ClientConn
}

func newServiceClients(c config.Env) (*serviceClients, error) {
	conn, e := grpc.NewClient(c.Dependencies.CalendarAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if e != nil {
		return nil, e
	}
	return &serviceClients{calendar: clientinfra.NewCalendar(api.NewCalendarServiceClient(conn), c.Dependencies.Timeout), connection: conn}, nil
}

func (c *serviceClients) Close() error {
	return c.connection.Close()
}
