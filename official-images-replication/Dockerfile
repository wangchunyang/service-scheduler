# Start from a Debian image with the latest version of Go installed
# and a workspace (GOPATH) configured at /go.
FROM golang

# Copy the local package files to the container's workspace.
ADD . /go/src/ghe/chun-yang-wang/official-images-replication

# Build the official-images-replication command inside the container.
# (You may fetch or manage dependencies here,
# either manually or with a tool like "godep".)
RUN go get github.com/fsouza/go-dockerclient
RUN go get golang.org/x/net/context
RUN go get github.com/coreos/etcd/client

RUN go install ghe/chun-yang-wang/official-images-replication

# Run the official-images-replication command by default when the container starts.
ENTRYPOINT /go/bin/official-images-replication -url=$REGISTRY_URL -u=$USERNAME -p=$PASSWORD -server=$SERVER_MODE -repos=$REPOS -discoveryEtcdURLs=$DISCOVERY_ETCD_URLS -discoveryServiceKey=$DISCOVERY_SERVICE_KEY -discoveryServiceURL=$DISCOVERY_SERVICE_URL

EXPOSE 8000
