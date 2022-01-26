package server

import (
	"breaker/pkg/protocol"
	"context"
	"sync"
)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	// WithContext sets the underline context.
	// It's very useful to control the workflow when send to response channel.
	WithContext(ctx context.Context) Context

	// Session returns the current session.
	Session() Session

	// SetSession sets session.
	SetSession(sess Session) Context

	// Request returns request message entry.
	Request() protocol.Command

	// SetRequest encodes data with session's codec and sets request message entry.
	SetRequest(id, data interface{}) error

	// MustSetRequest encodes data with session's codec and sets request message entry.
	// panics on error.
	MustSetRequest(id, data interface{}) Context

	// SetRequestMessage sets request message entry directly.
	SetRequestMessage(cmd protocol.Command) Context

	// Bind decodes request message entry to v.
	Bind(v interface{}) error

	// Response returns the response message entry.
	Response() protocol.Command

	// SetResponse encodes data with session's codec and sets response message entry.
	SetResponse(id, data interface{}) error

	// MustSetResponse encodes data with session's codec and sets response message entry.
	// panics on error.
	MustSetResponse(id, data interface{}) Context

	// SetResponseMessage sets response message entry directly.
	SetResponseMessage(cmd protocol.Command) Context

	// Send sends itself to current session.
	Send() bool

	// SendTo sends itself to session.
	SendTo(session Session) bool

	// Get returns key value from storage.
	Get(key string) (value interface{}, exists bool)

	// Set store key value into storage.
	Set(key string, value interface{})

	// Remove deletes the key from storage.
	Remove(key string)

	// Copy returns a copy of Context.
	Copy() Context
}

func NewContext() *routeContext {
	return &routeContext{
		rawCtx: context.Background(),
	}
}

// routeContext implements the Context interface.
type routeContext struct {
	rawCtx    context.Context
	mu        sync.RWMutex
	storage   map[string]interface{}
	session   Session
	reqEntry  protocol.Command
	respEntry protocol.Command
}

func (r *routeContext) WithContext(ctx context.Context) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Session() Session {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SetSession(sess Session) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Request() protocol.Command {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SetRequest(id, data interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) MustSetRequest(id, data interface{}) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SetRequestMessage(cmd protocol.Command) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Bind(v interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Response() protocol.Command {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SetResponse(id, data interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) MustSetResponse(id, data interface{}) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SetResponseMessage(cmd protocol.Command) Context {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Send() bool {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) SendTo(session Session) bool {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Get(key string) (value interface{}, exists bool) {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Set(key string, value interface{}) {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Remove(key string) {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Copy() Context {
	//TODO implement me
	panic("implement me")
}
