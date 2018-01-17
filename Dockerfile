FROM ubuntu 
ADD docker-socket-proxy /usr/local/bin
ENTRYPOINT ["/usr/local/bin/docker-socket-proxy"]
