package socketListener

import (
	"log"
	"net"
	"os"
	"strings"

	"github.com/stensonb/docker-socket-proxy/lib/config"
)

// SocketListener does stuff
type SocketListener struct {
	net.Listener
}

// New listens on a new socket
func New() (net.Listener, error) {
	// delete existing socket, if necessary
	if err := removeProxySocket(config.ProxySocketPath); err != nil && !strings.HasSuffix(err.Error(), "no such file or directory") {
		return SocketListener{}, err
	}

	// Looking up group ids coming up for Go 1.7
	// https://github.com/golang/go/issues/2617

	// if we failed to create a new socket, panic
	l, err := net.Listen("unix", config.ProxySocketPath)
	if err != nil {
		return l, err
	}

	if err := os.Chmod(config.ProxySocketPath, 0660); err != nil {
		return l, err
	}

	log.Println("Listening on", config.ProxySocketPath)

	return l, nil
}

// Cleanup cleans up the socketListener
func (sl *SocketListener) Cleanup() error {
	sl.Close()
	return removeProxySocket(config.ProxySocketPath)
}

func removeProxySocket(socketPath string) error {
	return os.Remove(socketPath)
}
