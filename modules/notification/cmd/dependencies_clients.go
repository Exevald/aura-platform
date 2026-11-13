package main

import (
	"errors"
	mailerapi "mailer/api/server/internalapi"
	"notification/internal/notification/app"
	"notification/internal/notification/config"
	clientinfra "notification/internal/notification/infra/client"
	realtimeapi "realtime/api/server/internalapi"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type serviceClients struct {
	realtime     app.RealtimePublisher
	mailer       app.MailerSender
	realtimeConn *grpc.ClientConn
	mailerConn   *grpc.ClientConn
}

func newServiceClients(c config.Env) (*serviceClients, error) {
	realtimeConn, err := grpc.NewClient(c.Dependencies.RealtimeAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	mailerConn, err := grpc.NewClient(c.Dependencies.MailerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = realtimeConn.Close()
		return nil, err
	}
	return &serviceClients{
		realtime:     clientinfra.NewRealtime(realtimeapi.NewRealtimeServiceClient(realtimeConn), c.Dependencies.Timeout),
		mailer:       clientinfra.NewMailer(mailerapi.NewMailerServiceClient(mailerConn), c.Dependencies.Timeout),
		realtimeConn: realtimeConn,
		mailerConn:   mailerConn,
	}, nil
}

func (c *serviceClients) Close() error {
	return errors.Join(c.realtimeConn.Close(), c.mailerConn.Close())
}
