package volume

import (
	"log"
	"net/http"
	"strings"

	"github.com/stensonb/docker-socket-proxy/lib/config"
	"github.com/stensonb/docker-socket-proxy/lib/fromDocker"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware"
)

// BindSeparator is how docker separates volume "host" path from volume "container" path for each Bind
const BindSeparator = ":"

// Middleware stores all the Lables
type Middleware struct {
	VolumePathOverrides map[string]string
}

// New creates new Volume middleware, and substitutes hostPath
// for all source volume requests through the proxy.
//
// The intenion is that calls for "docker run -v $source:$target ..." will be rewritten
// to "docker run -v $newsource:$target ..." to ensure the volume shared to the
// newly executed docker container come from the underlying host, rather than from
// within the container making the "docker run..." call.
//
// This approach helps to make "docker-in-docker" appear real, despite it actually being
// fraterdocker (sibling, rather than nested docker containers)
func New() *Middleware {
	// TODO: make these cli/env driven?
	return &Middleware{VolumePathOverrides: map[string]string{containerWorkspacePath(): hostWorkspacePath()}}
}

func containerWorkspacePath() string {
	return config.JenkinsWorkspaceContainerPath
}

func hostWorkspacePath() string {
	return config.JenkinsWorkspaceHostPath
}

// Cleanup implemented to satisfy RequestMiddleware interface
func (v *Middleware) Cleanup() error {
	return nil
}

// Apply performs the middleware on this http.Request
func (v *Middleware) Apply(r *http.Request) {
	switch getRequestType(r) {
	case createContainer:

		config, hostConfig, networkingConfig, err := requestMiddleware.ParseRequest(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Original Binds:", hostConfig.Binds)

		//rewrite hostConfig.Binds sources
		for i, bind := range hostConfig.Binds {
			splitUp := strings.Split(bind, BindSeparator)

			// for each bind, if the first element prefix matches any of the keys
			// in v.VolumePathOverrides, substitute the value of that key from the map for that prefix
			// ensuring any data after the prefix is preserved
			for key, value := range v.VolumePathOverrides {
				if strings.HasPrefix(splitUp[0], key) {
					splitUp[0] = strings.Replace(splitUp[0], key, value, -1) //value
				}
			}

			joined := strings.Join(splitUp, BindSeparator)
			hostConfig.Binds[i] = joined
		}

		log.Println("Proxied Binds:", hostConfig.Binds)

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
	}

}

// PeriodicCleanup implemented to satisfy RequestMiddleware interface
func (v *Middleware) PeriodicCleanup() {

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
	listContainers                     // docker ps
	unmodified
)
