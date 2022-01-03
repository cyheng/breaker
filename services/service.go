package services

import (
	"breaker/feature"
	"context"
	"fmt"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	Start(args interface{}, ctx context.Context) (err error)
	Stop(ctx context.Context)
}

type Instance struct {
	App Service

	Name string
}

var InstanceMap = make(map[string]*Instance)

func Register(name string, app Service) {
	InstanceMap[name] = &Instance{
		App: app,

		Name: name,
	}
}

func init() {
	Register(feature.FPortal, NewPortal())
	Register(feature.FBridge, NewBridge())
}

func Run(name string, args interface{}) error {
	instance, ok := InstanceMap[name]
	if !ok {
		err := fmt.Errorf("service %s not found", name)
		return err
	}
	defer func() {
		err := recover()
		if err != nil {
			e := err.(error)
			e = errors.WithStack(e)
			log.Fatalf("%s servcie crashed, ERR: %+v", name, e)
		}
	}()
	err := instance.App.Start(args, context.Background())
	if err != nil {
		log.Fatalf("%s servcie fail, ERR: %s", name, err)
		return err
	}
	return nil
}
