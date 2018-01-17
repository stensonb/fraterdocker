package config

import (
	"os"

	"github.com/namsral/flag"
)

// DefaultProxySocketPath is the path to the proxy socket this process will create/manage
const DefaultProxySocketPath = "/var/run/docker.sock"

// DefaultDockerSocketPath is the path to Docker Engine (which this process will proxy to)
const DefaultDockerSocketPath = "/var/run/docker-orig.sock"

// ProxySocketPath is the socket we listen on
var ProxySocketPath string

// DockerSocketPath is the upstream docker socket
var DockerSocketPath string

// Hostname is a helper which always returns a string
var Hostname string

// JenkinsWorkspaceHostPath is the path to the workspace on the docker host
var JenkinsWorkspaceHostPath string

// JenkinsWorkspaceContainerPath is the path to the workspace inside the jenkins build agent
var JenkinsWorkspaceContainerPath string

// New returns a config struct
//func New() error { //} (Config, error) {
func init() {
	flag.StringVar(&ProxySocketPath, "proxySocket", DefaultProxySocketPath, "Path to proxy socket")
	flag.StringVar(&DockerSocketPath, "dockerSocket", DefaultDockerSocketPath, "Path to backend docker socket")
	flag.StringVar(&JenkinsWorkspaceHostPath, "jenkinsWorkspaceHostPath", "/tmp", "Path to shared workspace on docker host")
	flag.StringVar(&JenkinsWorkspaceContainerPath, "jenkinsWorkspaceContainerPath", "/tmp", "Path to shared workspace inside jenkins build agent")

	flag.Parse()

	var err error
	Hostname, err = os.Hostname()
	if err != nil {
		panic(err)
	}
}
