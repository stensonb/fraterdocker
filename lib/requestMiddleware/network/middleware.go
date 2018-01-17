package network

import (
	"log"
	"net/http"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/stensonb/docker-socket-proxy/lib/config"
	"github.com/stensonb/docker-socket-proxy/lib/fromDocker"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware"
)

// Middleware stores stuff
type Middleware struct {
	Network
}

// New creates new Network middleware, creating a new docker network
// based on the hostname, using subnet details
func New() *Middleware {
	n := NewNetwork(name(), config.DockerSocketPath)

	return &Middleware{Network: n}
}

// Cleanup is called to remove the docker network when this
// process is shutdown cleanly
func (m Middleware) Cleanup() error {
	return m.Network.Cleanup()
}

// Apply performs the middleware on this http.Request
func (m Middleware) Apply(r *http.Request) {

	switch getRequestType(r) {
	case createContainer:
		config, hostConfig, networkingConfig, err := requestMiddleware.ParseRequest(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Original NetworkMode:", hostConfig.NetworkMode)

		// set the container's network name - this overwrites any desired network
		// name set by the client --
		// TODO: feature to support (somehow) both default and non-default networks
		hostConfig.NetworkMode = (container.NetworkMode)(m.Network.Name())

		log.Println("Proxied NetworkMode:", hostConfig.NetworkMode)

		// create new body
		newBody := fromDocker.ConfigWrapper{
			Config:           config,
			HostConfig:       hostConfig,
			NetworkingConfig: networkingConfig,
		}

		// encode it just like docker would
		newEncodedBody, newHeaders, err := fromDocker.EncodeBody(newBody, nil)
		if err != nil {
			log.Fatal(err)
		}

		c := fromDocker.NewClient("docker")
		newrequest, err := c.BuildRequest(r.Method, r.URL.String(), newEncodedBody, newHeaders)
		if err != nil {
			log.Fatal(err)
		}

		// set value of original pointer to value of newrequest
		*r = *newrequest
	default:
		// don't modify request
	}

}

// PeriodicCleanup implemented to satisfy RequestMiddleware interface
func (m *Middleware) PeriodicCleanup() {

}

func getRequestType(req *http.Request) requestType {
	if req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "containers/create") {
		return createContainer
	}

	return unmodified
}

type requestType uint

const (
	createContainer requestType = iota // docker create
	unmodified
)

// Name returns the network name
func name() string {
	return config.Hostname
}
