# Fraterdocker

This is binary provides a wrapper socket for docker.socket which provides a virtual cleanroom (but not secure (SEE BELOW)) for running jenkins build agents in docker.

Fraterdocker

## Features

- jenkins jobs may use standard `docker` CLI and API calls against the provided socket
- calls to `docker ps` will return only those containers built via this wrapper socket

## Known Limitations

## How

The wrapper socket intercepts all API calls, and:
- injects container labels on all `docker run` API calls
- adds label filters for the container labels on all `docker ps` API calls

## Why?

When jenkins agents use docker, they call `docker ps -aq | xargs docker rm -f` to ensure a "cleanroom" for docker builds, functional tests, etc.  This wrapper prevents the jenkins agent from killing itself (since it's a docker container) and any other non-jenkins container from being destroyed.

HUGE DISCLAIMER: This is not meant as a "secure" solution.  This only labels containers and filters for them.

## issues

2. cleanup on shutdown (hook os.signal and call cleanup on all middlewares, remove socket, etc)
3. refactor/organize code
4. update readme (specific invariants (paths, sockets, middlewares...))
5. tests (middleware unit tests, etc)
6. build pipeline

## Roapmap

1. define security model
2. implement security model
