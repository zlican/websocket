package websocket

import (
	"errors"
	"sync"
)

type (
	Request struct {
		RequestId string
		Event     string
		Client    *Client
		Send      []byte
		Data      []byte
	}
	DisposeFunc    func(*Request) *Response
	MiddlewareFunc func(DisposeFunc) DisposeFunc

	Router struct {
		handlers sync.Map
	}

	Route struct {
		final           DisposeFunc
		middlewares     []MiddlewareFunc
		postMiddlewares []MiddlewareFunc
	}
)

var (
	RouteManager *Router
)

type DistributeHandler interface {
	Distribute(client *Client, message []byte) (err error)
}

func NewRouter() *Router {
	RouteManager = &Router{}
	return RouteManager
}

// Register event to memory
func (r *Router) Register(event string, handler DisposeFunc) *Route {
	route := &Route{
		final: handler,
	}
	r.handlers.Store(event, route)
	return route
}

// GetRoute Get register's route
func (r *Router) GetRoute(event string) (route *Route, err error) {
	value, ok := r.handlers.Load(event)
	if !ok {
		return nil, errors.New("current event not supported")
	}
	route, okk := value.(*Route)
	if !okk {
		return nil, errors.New("handler type error")
	}
	return
}

// Use Route Create a new middleware chain
func (r *Route) Use(middleware ...MiddlewareFunc) *Route {
	r.middlewares = append(r.middlewares, middleware...)
	return r
}

// UsePost Route Create a new middleware chain
func (r *Route) UsePost(middleware ...MiddlewareFunc) *Route {
	r.postMiddlewares = append(r.postMiddlewares, middleware...)
	return r
}

// Execute the middleware chain
func (r *Route) Execute(request *Request) *Response {
	handler := r.final
	for i := len(r.middlewares) - 1; i >= 0; i-- {
		handler = r.middlewares[i](handler)
	}
	response := handler(request)
	for _, middleware := range r.postMiddlewares {
		response = middleware(func(req *Request) *Response {
			return response
		})(request)
	}
	return response
}
