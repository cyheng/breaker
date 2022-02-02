package breaker

import (
	"breaker/pkg/protocol"
	"reflect"
)

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx Context)

var nilHandler HandlerFunc = func(ctx Context) {}

// MiddlewareFunc is the function type for middlewares.
// A common pattern is like:
//
// 	var mf MiddlewareFunc = func(next HandlerFunc) HandlerFunc {
// 		return func(ctx Context) {
// 			next(ctx)
// 		}
// 	}
type MiddlewareFunc func(next HandlerFunc) HandlerFunc
type Router struct {
	// handlerMapper maps message's ID to handler.
	// Handler will be called around middlewares.
	handlerMapper map[reflect.Type]HandlerFunc

	// middlewaresMapper maps message's ID to a list of middlewares.
	// These middlewares will be called before the handler in handlerMapper.
	middlewaresMapper map[reflect.Type][]MiddlewareFunc

	// globalMiddlewares is a list of MiddlewareFunc.
	// globalMiddlewares will be called before the ones in middlewaresMapper.
	globalMiddlewares []MiddlewareFunc

	notFoundHandler HandlerFunc
}

func (r *Router) handleRequest(ctx Context) {
	cmd := ctx.Request()
	if cmd == nil {
		return
	}
	var handler HandlerFunc
	methodID := reflect.TypeOf(cmd)
	if v, has := r.handlerMapper[methodID]; has {
		handler = v
	}

	var mws = r.globalMiddlewares
	if v, has := r.middlewaresMapper[methodID]; has {
		mws = append(mws, v...) // append to global ones
	}

	// create the handlers stack
	wrapped := r.wrapHandlers(handler, mws)

	// and call the handlers stack
	wrapped(ctx)
}

// 	var wrapped HandlerFunc = m1(m2(m3(handle)))
func (r *Router) wrapHandlers(handler HandlerFunc, middles []MiddlewareFunc) (wrapped HandlerFunc) {
	if handler == nil {
		handler = r.notFoundHandler
	}
	if handler == nil {
		handler = nilHandler
	}
	wrapped = handler
	for i := len(middles) - 1; i >= 0; i-- {
		m := middles[i]
		wrapped = m(wrapped)
	}
	return wrapped
}

func (r *Router) setNotFoundHandler(handler HandlerFunc) {
	r.notFoundHandler = handler
}

func (r *Router) registerMiddleware(middlewares ...MiddlewareFunc) {
	for _, mm := range middlewares {
		if mm != nil {
			r.globalMiddlewares = append(r.globalMiddlewares, mm)
		}
	}
}

func (r *Router) register(cmd protocol.Command, handler HandlerFunc, middlewares ...MiddlewareFunc) {
	id := reflect.TypeOf(cmd)
	if handler != nil {
		r.handlerMapper[id] = handler
	}
	ms := make([]MiddlewareFunc, 0, len(middlewares))
	for _, mm := range middlewares {
		if mm != nil {
			ms = append(ms, mm)
		}
	}
	if len(ms) != 0 {
		r.middlewaresMapper[id] = ms
	}
}

func NewRouter() *Router {
	return &Router{
		handlerMapper:     make(map[reflect.Type]HandlerFunc),
		middlewaresMapper: make(map[reflect.Type][]MiddlewareFunc),
	}
}
