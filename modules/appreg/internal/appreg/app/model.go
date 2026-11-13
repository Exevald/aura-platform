package app

import (
	"errors"
	"uuid"
)

var ErrNotFound = errors.New("application not found")

type App struct {
	ID             uuid.UUID
	Code           string
	Name           string
	FrontendRoute  string
	BackendTarget  string
	DefaultEnabled bool
}

func Apps(catalogue []App, defaultOnly bool) []App {
	out := make([]App, 0, len(catalogue))
	for _, v := range catalogue {
		if !defaultOnly || v.DefaultEnabled {
			out = append(out, v)
		}
	}
	return out
}

func Find(catalogue []App, id uuid.UUID) (App, error) {
	for _, v := range catalogue {
		if v.ID == id {
			return v, nil
		}
	}
	return App{}, ErrNotFound
}
