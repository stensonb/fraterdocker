package fromDocker

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/versions"
)

// headers taken from docker/client/request.go
// TODO: verify license
type headers map[string][]string

type Headers struct{ *headers }

// configWrapper from docker/client/container_create.go
type ConfigWrapper struct {
	*container.Config
	HostConfig       *container.HostConfig
	NetworkingConfig *network.NetworkingConfig
}

// encodeBody() taken from docker/client/request.go
// TODO: refactor so we copy/paste less
// TODO: verify license
func EncodeBody(obj interface{}, headers headers) (io.Reader, headers, error) {
	if obj == nil {
		return nil, headers, nil
	}

	body, err := encodeData(obj)
	if err != nil {
		return nil, headers, err
	}
	if headers == nil {
		headers = make(map[string][]string)
	}
	headers["Content-Type"] = []string{"application/json"}
	return body, headers, nil
}

// encodeData() taken from docker/client/request.go
// TODO: verify license
func encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if err := json.NewEncoder(params).Encode(data); err != nil {
			return nil, err
		}
	}
	return params, nil
}

// Client{} from docker/client/client.go
// added here to add supporting functions to export private functions
// from the docker code base (functions copied line-by-line below, and noted)
// TODO: can we fix this code copying business?
type Client struct {
	// scheme sets the scheme for the client
	scheme string
	// host holds the server address to connect to
	host string
	// proto holds the client protocol i.e. unix.
	proto string
	// addr holds the client address.
	addr string
	// basePath holds the path to prepend to the requests.
	basePath string
	// client used to send and receive http requests.
	client *http.Client
	// version of the server to talk to.
	version string
	// custom http headers configured by users.
	customHTTPHeaders map[string]string
	// manualOverride is set to true when the version was set by users.
	manualOverride bool
}

func NewClient(addr string) *Client {
	return &Client{scheme: "http",
		addr:    addr,
		version: "v1.24"}
}

func (cli *Client) BuildRequest(method, path string, body io.Reader, headers headers) (*http.Request, error) {
	return cli.buildRequest(method, path, body, headers)
}

// buildRequest() from docker/client/request.go
func (cli *Client) buildRequest(method, path string, body io.Reader, headers headers) (*http.Request, error) {
	expectedPayload := (method == "POST" || method == "PUT")
	if expectedPayload && body == nil {
		body = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, err
	}
	req = cli.addHeaders(req, headers)

	if cli.proto == "unix" || cli.proto == "npipe" {
		// For local communications, it doesn't matter what the host is. We just
		// need a valid and meaningful host name. (See #189)
		req.Host = "docker"
	}

	req.URL.Host = cli.addr
	req.URL.Scheme = cli.scheme

	if expectedPayload && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "text/plain")
	}
	return req, nil
}

// addHeaders() from docker/client/request.go
func (cli *Client) addHeaders(req *http.Request, headers headers) *http.Request {
	// Add CLI Config's HTTP Headers BEFORE we set the Docker headers
	// then the user can't change OUR headers
	for k, v := range cli.customHTTPHeaders {
		if versions.LessThan(cli.version, "1.25") && k == "User-Agent" {
			continue
		}
		req.Header.Set(k, v)
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}
	return req
}
