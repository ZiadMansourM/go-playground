package middleware

import "net/http"

type Middleware func(http.Handler) http.HandlerFunc

func ChainMiddlewares(handler http.Handler, middlewares ...Middleware) http.Handler {
	for _, mw := range middlewares {
		handler = mw(handler)
	}
	return handler
}
