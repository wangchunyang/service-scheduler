// Package main provides REST API for starting replication.
// Endpoint: /api/1.0/replicate
// Method: GET
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

// Server implements http.Handler
type Server struct {
}

var running chan int

var kapi client.KeysAPI

// Response contains all of the data that we want to send to client
type Response struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// DiscoveryOptions is structure of service discovery
type DiscoveryOptions struct {
	EtcdURLs   []string
	ServiceURL string
	ServiceKey string
}

func setupDiscoveryService(opts *DiscoveryOptions) {
	// etcdURL := "http://172.17.42.1:2379"
	cfg := client.Config{
		Endpoints: opts.EtcdURLs,
		Transport: client.DefaultTransport,
		// set timeout per request to fail fast when the target endpoint is unavailable
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
		log.Fatal(err)
	}
	kapi = client.NewKeysAPI(c)
}
func (s *Server) start(discoveryOptions *DiscoveryOptions) {
	log.Println("startServer(): entry")

	setupDiscoveryService(discoveryOptions)

	// init the running channel with capacity 1.
	running = make(chan int, 1)
	http.Handle("/api/1.0/replicate", &Server{})
	log.Println("The server is ready!")

	//serviceURL := "http://localhost:8000/api/1.0/replicate?filter=list&repos=busybox"
	//key := "/services/repl-1"

	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for _ = range ticker.C {
			registerService(discoveryOptions.ServiceKey, discoveryOptions.ServiceURL, 60*time.Second)
			//log.Printf("register %s", serviceURL)
		}
	}()

	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}
}
func registerService(key string, serviceURL string, ttl time.Duration) {
	resp, err := kapi.Set(context.Background(), key, serviceURL, &client.SetOptions{TTL: ttl})
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("register %s: %v", key, resp)
}
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request => %s\n", r.URL.String())
	select {
	case running <- 1:
		go func() {
			log.Println("Replication is starting.")
			filter := r.URL.Query().Get("filter")
			repos := r.URL.Query().Get("repos")
			log.Printf("received http request: filter=%s, repos=%s", filter, repos)
			// time.Sleep(10 * time.Second)
			replication.Replicate(filter, repos)
			<-running
			log.Println("Replication is done.")
		}()
		s.response(w, "OK", "The replication is starting.")

	default:
		log.Println("The server is busy now. Try again later.")
		s.response(w, "BUSY", "The server is busy now. Try again later.")
	}
}
func (s *Server) response(w http.ResponseWriter, status string, message string) {
	response := Response{status, message}
	output, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Failed to marshal json byte array")
	}
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprint(w, string(output))
}
