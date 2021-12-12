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
	return nil
}
