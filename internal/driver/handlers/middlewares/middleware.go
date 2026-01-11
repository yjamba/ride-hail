package middlewares

import "net/http"

type MiddlewareChain struct {
	middlewares []func(next http.HandlerFunc) http.HandlerFunc
}

func NewMiddlewareChain(middlewares ...func(next http.HandlerFunc) http.HandlerFunc) *MiddlewareChain {
	return &MiddlewareChain{
		middlewares: middlewares,
	}
}

func (mc *MiddlewareChain) WrapHandler(handler http.HandlerFunc) http.HandlerFunc {
	for i := len(mc.middlewares) - 1; i >= 0; i-- {
		handler = mc.middlewares[i](handler)
	}

	return handler
}
