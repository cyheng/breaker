package server

// HandlerFunc is the function type for handlers.
type HandlerFunc func(ctx Context)

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
	handlerMapper map[interface{}]HandlerFunc

	// middlewaresMapper maps message's ID to a list of middlewares.
	// These middlewares will be called before the handler in handlerMapper.
	middlewaresMapper map[interface{}][]MiddlewareFunc

	// globalMiddlewares is a list of MiddlewareFunc.
	// globalMiddlewares will be called before the ones in middlewaresMapper.
	globalMiddlewares []MiddlewareFunc

	notFoundHandler HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		handlerMapper:     make(map[interface{}]HandlerFunc),
		middlewaresMapper: make(map[interface{}][]MiddlewareFunc),
	}
}
