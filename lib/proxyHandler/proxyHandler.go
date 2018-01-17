package proxyHandler

import (
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/stensonb/docker-socket-proxy/lib/dockerProxy"
)

// ProxyHandler does stuff
type ProxyHandler struct {
	dp dockerProxy.DockerProxy
}

// New builds a docker proxy handler
func New() (ProxyHandler, error) {
	dp := dockerProxy.New()
	return ProxyHandler{dp: dp}, nil
}

// Cleanup does stuff
func (ph *ProxyHandler) Cleanup() error {
	return ph.dp.Cleanup()
}

// Handler returns the http.Handler
func (ph *ProxyHandler) Handler() http.Handler {
	return handlers.CombinedLoggingHandler(os.Stdout, ph.dp.Handler())
}
