package breaker

import (
	log "github.com/sirupsen/logrus"
	"runtime/debug"
)

func RecoverMiddleware() MiddlewareFunc {
	return func(next HandlerFunc) HandlerFunc {
		return func(c Context) {
			defer func() {
				if r := recover(); r != nil {
					log.WithField("trace id", c.Session().ID()).Errorf("PANIC | %+v | %s", r, debug.Stack())
				}
			}()
			next(c)
		}
	}
}
