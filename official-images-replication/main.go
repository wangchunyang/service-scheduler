// Package main provides REST API for starting replication.
// Endpoint: /api/1.0/replicate
// Method: GET
package main

import (
	"flag"
	"fmt"
	"strings"
)

var replication *Replication

var (
	registryURL         = flag.String("url", "docker.hos.hpecorp.net/library", "The base URL of official images on Private Registry")
	username            = flag.String("u", "", "Username to access private registry")
	password            = flag.String("p", "", "Password to access private registry")
	serverMode          = flag.Bool("server", true, "Running mode")
	filter              = flag.String("filter", "list", "list|letter")
	repos               = flag.String("repos", "", "Only use it when mode=standalone. The example is 'centos,busybox,ubuntu' or 'a-f'")
	discoveryEtcdURLs   = flag.String("discoveryEtcdURLs", "", "The URLs of etcd cluster")
	discoveryServiceURL = flag.String("discoveryServiceURL", "", "The URL of replication service running in current container")
	discoveryServiceKey = flag.String("discoveryServiceKey", "", "The URL of replication service running in current container")
)

func main() {
	fmt.Println("main(): entry")
	flag.Parse()
	fmt.Printf("Command line options: registryURL=%s, username=%s, server=%v, filter=%s, repos=%s\n", *registryURL, *username, *serverMode, *filter, *repos)

	replication = &Replication{}

	defaultRegistry := "docker.hos.hpecorp.net/library"

	if *registryURL == "" {
		registryURL = &defaultRegistry
	}

	replication.registryURL = *registryURL
	replication.username = *username
	replication.password = *password

	if *serverMode {
		server := &Server{}
		etcdURLs := strings.Split(*discoveryEtcdURLs, ",")
		server.start(&DiscoveryOptions{EtcdURLs: etcdURLs, ServiceURL: *discoveryServiceURL, ServiceKey: *discoveryServiceKey})
	} else {
		replication.Replicate(*filter, *repos)
	}
}
