# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/ghe/chun-yang-wang/service-scheduler

# Build the official-images-replication command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get golang.org/x/net/context
RUN go get github.com/coreos/etcd/client

RUN go install ghe/chun-yang-wang/service-scheduler

# Run the official-images-replication command by default when the container starts.
ENTRYPOINT /go/bin/service-scheduler -discoveryEtcdURLs=$DISCOVERY_ETCD_URLS -discoveryServiceDir=$DISCOVERY_SERVICE_DIR

EXPOSE 8000
