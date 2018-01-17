package dockerProxy

import (
	"context"
	"log"
	"net"
	"net/http"

	// use my modified reverseproxy struct to support docker hijacking

	"github.com/stensonb/docker-socket-proxy/lib/config"
	"github.com/stensonb/docker-socket-proxy/lib/fromGo/httputil"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware/label"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware/network"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware/volume"
)

// DockerProxy does stuff
type DockerProxy struct {
	rp          *httputil.ReverseProxy
	middlewares []requestMiddleware.RequestMiddleware
}

// New returns a DockerProxy
func New() DockerProxy {

	// TODO: set the middleware via ENV or arg?
	ans := DockerProxy{middlewares: []requestMiddleware.RequestMiddleware{}}

	ans.middlewares = append(ans.middlewares, label.New())
	ans.middlewares = append(ans.middlewares, volume.New())
	ans.middlewares = append(ans.middlewares, network.New())

	director := func(req *http.Request) {
		// call the docker API via HTTP over unix socket (hostname must be set to something)
		req.URL.Scheme = "http"
		req.URL.Host = "docker"

		// apply middleware
		ans.applyMiddlewares(req)
	}

	_proto := "unix"
	_addr := config.DockerSocketPath

	transport := &http.Transport{
		DialContext: func(ctx context.Context, proto string, addr string) (net.Conn, error) {
			// force proto=unix and addr=upstream to ensure we talk to docker api
			// via HTTP over local unix socket
			proto = _proto
			addr = _addr

			d := net.Dialer{}
			conn, err := d.DialContext(context.Background(), proto, addr)

			return conn, err
		},
		DisableCompression: true,
	}

	ans.rp = &httputil.ReverseProxy{RawHandlerProto: _proto, RawHandlerAddress: _addr, Director: director, Transport: transport}

	log.Println("Upstream docker is", config.DockerSocketPath)

	return ans
}

// Handler returns the http.Handler for this DockerProxy
func (dp *DockerProxy) Handler() http.Handler {
	return dp.rp
}

// Cleanup cleans up all middlewares
func (dp *DockerProxy) Cleanup() error {
	for _, m := range dp.middlewares {
		m.Cleanup()
	}
	// TODO: should return errors from each middleware
	return nil
}

func (dp *DockerProxy) applyMiddlewares(req *http.Request) {
	// modify request as needed
	for _, m := range dp.middlewares {
		m.Apply(req)
	}
}
