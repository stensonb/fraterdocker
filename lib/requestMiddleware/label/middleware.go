package label

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/stensonb/docker-socket-proxy/lib/fromDocker"
	"github.com/stensonb/docker-socket-proxy/lib/requestMiddleware"
)

// LabelToAppendToContainers is applied to all containers started via this proxy
const LabelToAppendToContainers = "jenkins-build-agent-container"

// Version is applied to all containers started via this proxy
const Version = "1.0.0"

// Middleware stores all the Lables
type Middleware struct {
	Labels []Label
}

// New creates new Label middleware, and applies "Version",
// "LabelToAppendToContainers", and the "Hostname" to all containers started
// via this proxy
func New() Middleware {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	return Middleware{Labels: []Label{
		{LabelToAppendToContainers, Version},
		{"hostname", hn},
	}}
}

// Cleanup implemented to satisfy RequestMiddleware interface
func (lm Middleware) Cleanup() error {
	return nil
}

func (lm *Middleware) addLabelsToConfig(config *container.Config) {
	for _, l := range lm.Labels {
		l.AddToConfig(config)
	}
}

// Apply performs the middleware on this http.Request
func (lm Middleware) Apply(r *http.Request) {
	switch getRequestType(r) {
	case createContainer:

		config, hostConfig, networkingConfig, err := requestMiddleware.ParseRequest(r)
		if err != nil {
			log.Fatal(err)
		}

		log.Println("Original Labels:", config.Labels)

		lm.addLabelsToConfig(config)

		log.Println("Proxied Labels:", config.Labels)

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
	case listContainers:
		lm.setLabelFilter(r)
	default:
	}
}

// PeriodicCleanup implemented to satisfy RequestMiddleware interface
func (lm *Middleware) PeriodicCleanup() {

}

func getRequestType(req *http.Request) requestType {
	//log.Println(req.URL.Path)

	if req.Method == http.MethodPost && strings.HasSuffix(req.URL.Path, "containers/create") {
		//log.Println("found a createContainer request")
		return createContainer
	}

	if req.Method == http.MethodGet && strings.HasSuffix(req.URL.Path, "containers/json") {
		//log.Println("found a listContainers request")
		return listContainers
	}

	return unmodified
}

type requestType uint

const (
	createContainer requestType = iota // docker create
	listContainers                     // docker ps
	unmodified
)

func addLabelsToFilters(requestFilters filters.Args, labels []Label) {
	for _, l := range labels {
		// if missing, add the label we'll use to "namespace" containers
		if !requestFilters.Include(l.String()) {
			requestFilters.Add("label", l.String())
		}
	}
}

func (lm *Middleware) setLabelFilter(r *http.Request) {
	// get filters passed in to proxy
	requestFilters, err := filters.FromParam(r.Form.Get("filters"))
	if err != nil {
		log.Fatal(err)
	}

	addLabelsToFilters(requestFilters, lm.Labels)

	// convert filters to a param string
	param, err := filters.ToParam(requestFilters)
	if err != nil {
		log.Fatal(err)
	}

	// set the url encoded value of param as the rawquery
	r.URL.RawQuery = "filters=" + url.QueryEscape(param)
}
