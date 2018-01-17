BINS=docker-socket-proxy

$(BINS): docker-socket-proxy.go
	GOOS=linux GOARCH=amd64 go build .

clean:
	rm -rf $(BINS)

all: clean $(BINS)
