package requestMiddleware

import "net/http"

// RequestMiddleware is an interface for all request middleware
type RequestMiddleware interface {
	Apply(*http.Request)
	Cleanup() error
}
