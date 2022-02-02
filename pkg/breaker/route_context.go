package breaker

import (
	"breaker/pkg/protocol"
	"context"
	"net"
	"sync"
	"time"
)

// Context is a generic context in a message routing.
// It allows us to pass variables between handler and middlewares.
type Context interface {
	context.Context
	// WithContext sets the underline context.
	// It's very useful to control the workflow when send to response channel.
	WithContext(ctx context.Context) Context

	// Session returns the current Session.
	Session() Session

	// SetSession sets Session.
	SetSession(sess Session) Context

	// Request returns request message entry.
	Request() protocol.Command

	// SetRequestMessage sets request message entry directly.
	SetRequestMessage(cmd protocol.Command) Context

	// Bind decodes request message entry to v.
	Bind(v interface{}) error

	// Response returns the response message entry.
	Response() protocol.Command

	// SetResponseMessage sets response message entry directly.
	SetResponseMessage(cmd protocol.Command) Context

	// Send sends itself to current Session.
	Send() bool

	SendSync() bool
	// SendTo sends itself to Session.
	SendTo(session Session) bool

	// Get returns key value from storage.
	Get(key string) (value interface{}, exists bool)

	// Set store key value into storage.
	Set(key string, value interface{})

	// Remove deletes the key from storage.
	Remove(key string)

	Conn() net.Conn
}

func NewContext() *routeContext {
	return &routeContext{
		rawCtx: context.Background(),
	}
}

// routeContext implements the Context interface.
type routeContext struct {
	rawCtx context.Context

	storage   sync.Map
	session   Session
	reqEntry  protocol.Command
	respEntry protocol.Command
}

func (r *routeContext) Deadline() (deadline time.Time, ok bool) {
	return r.rawCtx.Deadline()
}

func (r *routeContext) Done() <-chan struct{} {
	return r.rawCtx.Done()
}

func (r *routeContext) Err() error {
	return r.rawCtx.Err()
}

func (r *routeContext) Value(key interface{}) interface{} {
	if keyAsString, ok := key.(string); ok {
		val, _ := r.Get(keyAsString)
		return val
	}
	return nil
}

func (r *routeContext) WithContext(ctx context.Context) Context {
	r.rawCtx = ctx
	return r
}

func (r *routeContext) Session() Session {
	return r.session
}

func (r *routeContext) SetSession(sess Session) Context {
	r.session = sess
	return r
}

func (r *routeContext) Request() protocol.Command {
	return r.reqEntry
}

func (r *routeContext) SetRequestMessage(cmd protocol.Command) Context {
	r.reqEntry = cmd
	return r
}

func (r *routeContext) Bind(v interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (r *routeContext) Response() protocol.Command {
	return r.respEntry
}

func (r *routeContext) SetResponseMessage(cmd protocol.Command) Context {
	r.respEntry = cmd
	return r
}

func (r *routeContext) Send() bool {
	return r.session.Send(r)
}
func (r *routeContext) SendSync() bool {
	return r.session.SendSync(r)
}

func (r *routeContext) SendTo(session Session) bool {
	return session.Send(r)
}

func (r *routeContext) Get(key string) (value interface{}, exists bool) {
	return r.storage.Load(key)
}

func (r *routeContext) Set(key string, value interface{}) {
	r.storage.Store(key, value)
}

func (r *routeContext) Remove(key string) {
	r.storage.Delete(key)
}
func (r *routeContext) Conn() net.Conn {
	return r.session.Conn()
}
func (r *routeContext) reset() {
	r.rawCtx = context.Background()
	r.session = nil
	r.reqEntry = nil
	r.respEntry = nil
	r.storage = sync.Map{}
}
