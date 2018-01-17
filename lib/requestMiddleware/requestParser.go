package requestMiddleware

import (
	"log"
	"net/http"

	"github.com/docker/docker/api/server/httputils"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/runconfig"
)

// ParseRequest is a common function for use by middleware
func ParseRequest(r *http.Request) (*container.Config, *container.HostConfig, *network.NetworkingConfig, error) {

	// parse using docker code
	if err := httputils.ParseForm(r); err != nil {
		log.Fatal(err)
	}

	// verify json using docker code
	if err := httputils.CheckForJSON(r); err != nil {
		log.Fatal(err)
	}

	// decode body using docker api code
	decoder := &runconfig.ContainerDecoder{}

	return decoder.DecodeConfig(r.Body)
}
