package client

import (
	"access/internal/access/app"
	api "appreg/api/server/publicapi"
	"context"
	"errors"
	"time"
	"uuid"
)

type AppRegistry interface {
	List(context.Context) ([]app.Module, error)
}

type appRegistry struct {
	client  *api.ClientWithResponses
	timeout time.Duration
}

func NewAppRegistry(c *api.ClientWithResponses, t time.Duration) AppRegistry {
	return &appRegistry{c, t}
}

func (c *appRegistry) List(ctx context.Context) ([]app.Module, error) {
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()
	response, e := c.client.ListAppsWithResponse(ctx)
	if e != nil {
		return nil, e
	}
	if response.JSON200 == nil {
		return nil, errors.New("app registry unavailable")
	}
	out := make([]app.Module, 0, len(*response.JSON200))
	for _, v := range *response.JSON200 {
		id, err := uuid.Parse(v.Id.String())
		if err != nil {
			return nil, err
		}
		out = append(out, app.Module{ID: id, Code: v.Code, Name: v.Name})
	}
	return out, nil
}
