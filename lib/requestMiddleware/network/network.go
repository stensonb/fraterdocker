package network

import (
	"fmt"
	"log"
	"math/rand"
	"os/exec"
	"strings"
	"time"
)

// Network is a docker network
type Network struct {
	name     string
	upstream string
}

func newSubnet() string {
	// TODO: this should be based on CLI argument, and a smaller (configurable?) subnet.../24 by defaut?
	// TODO: allow cidr mask to be configurable
	return fmt.Sprintf("10.42.%d.0/24", rand.Intn(255))
}

// TODO: do this with a golang docker client...cleaner/neater/quicker/etc
// shell out to execute docker from the local binary path
func createNetwork(upstream, networkName string) error {
	// TODO: this loop logic is ugly...fixme, maybe with golang client
	var err error
	var output []byte
	tries := 10
	n := 0
	done := false

	// create a network, and retry if it failed because of overlap
	for output, err = exec.Command("docker", "-H", "unix://"+upstream, "network", "create", networkName, "--subnet", newSubnet()).CombinedOutput(); !done && err != nil && n < tries; output, err = exec.Command("docker", "-H", "unix://"+upstream, "network", "create", networkName, "--subnet", newSubnet()).CombinedOutput() {
		n++
		if err != nil {
			// ignore if network already exists
			if strings.Contains(string(output), networkName+" already exists") {
				log.Printf("Docker network '%s' already exists\n", networkName)
				done = true
			} else if strings.Contains(string(output), "networks have overlapping IPv4") {
				log.Println("Overlapping subnets. Trying a different one.")
			} else {
				// try setting to a different subnet  (calculate subnet from block, hashed by hostname, maybe?)
				log.Println("Failed, not due to overlapping subnets.")
				log.Println(err.Error())
			}
		}
	}

	// if done, we have a network created, return nil
	if done {
		return nil
	}

	return err
}

// NewNetwork returns a new Network struct
func NewNetwork(networkName, upstream string) Network {
	// initialize random with a time based seed
	rand.Seed(time.Now().UnixNano())

	log.Printf("Creating docker network '%s'\n", networkName)

	err := createNetwork(upstream, networkName)
	if err != nil {
		log.Fatal(err)
	}

	return Network{name: networkName, upstream: upstream}
}

// Cleanup removes the docker network
func (n *Network) Cleanup() error {
	// TODO: remove network via the docker client
	_, err := exec.Command("docker", "-H", "unix://"+n.Upstream(), "network", "remove", n.Name()).CombinedOutput()
	if err != nil {
		// TODO: handle this better
		log.Println("Failed to remove network.")
	}

	return err
}

// Name returns the name
func (n *Network) Name() string {
	return n.name
}

// Upstream return the upstream
func (n *Network) Upstream() string {
	return n.upstream
}
