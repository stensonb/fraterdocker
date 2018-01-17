package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/stensonb/docker-socket-proxy/lib/proxyHandler"
	"github.com/stensonb/docker-socket-proxy/lib/socketListener"
)

var sl socketListener.SocketListener
var ph proxyHandler.ProxyHandler

func main() {

	handleSigterm()

	var err error
	ph, err = proxyHandler.New()
	if err != nil {
		log.Fatal(err)
	}

	sl, err := socketListener.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := http.Serve(sl, ph.Handler()); err != nil {
		log.Fatal(err)
	}
}

func shutdownNicely() {
	if err := sl.Cleanup(); err != nil {
		log.Println(err)
	}

	if err := ph.Cleanup(); err != nil {
		log.Println(err)
	}
}

// handleSigterm does stuff
func handleSigterm() {
	// make sure we cleanup on exit
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, os.Interrupt, os.Kill, syscall.SIGTERM)
	go func(c chan os.Signal) {
		sig := <-c
		log.Printf("Caught signal %s: shutting down.\n", sig)
		shutdownNicely()
		os.Exit(0)
	}(sigc)
}
