package app

import (
	"breaker/feature"
	"context"
)

type App struct {
	ctx      context.Context
	features []feature.Feature
}

func New(f ...feature.Feature) *App {
	return &App{
		ctx:      context.Background(),
		features: f,
	}
}

func (a *App) Run() error {
	for _, f := range a.features {
		server, ok := f.(feature.Server)
		if ok {
			err := server.Start(a.ctx)
			if err != nil {
				_ = server.Stop(a.ctx)
				return err
			}
		}
		client, ok := f.(feature.Client)
		if ok {
			err := client.Start()
			if err != nil {
				_ = client.Stop(a.ctx)
				return err
			}
		}

	}
	return nil
}
