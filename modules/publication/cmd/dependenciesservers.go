package main

import (
	service2 "publication/internal/publication/app/service"

	"github.com/distributed-programming-2026/lib/runtime/http"

	"publication/internal/publication/config"
	publichandler "publication/internal/publication/infra/handlers/public"
)

func newHTTPServer(conf config.Env, releases *service2.Releases) (*http.Server, error) {
	register, err := publichandler.Register(releases)
	if err != nil {
		return nil, err
	}

	return http.NewServer(http.Config{
		Address:      conf.HTTP.Address,
		GraceTimeout: conf.App.GraceTimeout,
	}, register)
}
