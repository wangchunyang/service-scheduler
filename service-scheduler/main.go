package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/context"

	"github.com/coreos/etcd/client"
)

var (
	discoveryEtcdURLs   = flag.String("discoveryEtcdURLs", "", "The URLs of etcd cluster")
	discoveryServiceDir = flag.String("discoveryServiceDir", "", "The directory of replication service running in current container")
)

var (
	// ErrDirNotFound indicates the directory is not found in Service Discovery
	ErrDirNotFound = errors.New("Directory Not Found")

	// ErrServiceNotFound indicates the service is not found in Service Discovery
	ErrServiceNotFound = errors.New("Service Not Found")

	// ErrServiceNotExecuted indicates the service is not executed successfully
	ErrServiceNotExecuted = errors.New("Service Not Executed")
)

// ServiceDescription describes the meta-data of a service registered in Service-Discovery
type ServiceDescription struct {
	Key string `json:"key"`
	URL string `json:"url"`
}

// ServiceExecutionResult indicates the response from executing the service
type ServiceExecutionResult struct {
	Status  string
	Message string
}

var kapi client.KeysAPI

func setup(etcdURLs []string) {
	// etcdURL := "http://172.17.42.1:2379"
	cfg := client.Config{
		Endpoints: etcdURLs,
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

func main() {
	flag.Parse()
	log.Printf("Scheduler is starting... discoveryEtcdURLs=%s, discoveryServiceDir=%s", *discoveryEtcdURLs, *discoveryServiceDir)
	etcdURLs := strings.Split(*discoveryEtcdURLs, ",")
	setup(etcdURLs)
	go func() { schedule() }()
	ticker := time.NewTicker(60 * time.Second)
	go func() {
		for _ = range ticker.C {
			schedule()
			//log.Printf("register %s", serviceURL)
		}
	}()

	http.HandleFunc("/api/1.0/scheduler", handler)
	err := http.ListenAndServe(":8000", nil)
	if err != nil {
		log.Fatal(err)
	}

}
func schedule() {
	services, err := discoverServices(*discoveryServiceDir)
	if err != nil {
		log.Printf("scheduler %s: %v", *discoveryServiceDir, err)
		return
	}
	for _, service := range services {
		route(service)
	}
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Scheduler is Running")
}

func discoverServices(dir string) ([]string, error) {
	resp, err := kapi.Get(context.Background(), dir, &client.GetOptions{Recursive: true})
	if err != nil {
		log.Printf("discoverServices %s: %v", dir, err)
		return nil, ErrServiceNotFound
	}
	nodes := resp.Node.Nodes
	log.Printf("discoverServices %s: %v", dir, nodes)
	keys := make([]string, len(nodes))
	for i, node := range nodes {
		keys[i] = node.Key
	}

	return keys, nil
}

func route(serviceKey string) {
	serviceURL, err := findService(serviceKey)
	if err != nil {
		log.Printf("scheduler %s: %v", serviceKey, err)
		return
	}

	result, err := executeService(serviceURL)
	if err != nil {
		log.Printf("service-discovery %s: %v", serviceKey, err)
		return
	}
	log.Printf("service-discovery %s: %v", serviceKey, result)
}

func findService(serviceKey string) (string, error) {
	resp, err := kapi.Get(context.Background(), serviceKey, nil)
	if err != nil {
		log.Printf("finder %s: %v", serviceKey, err)
		return "", ErrServiceNotFound
	}

	log.Printf("finder %s: %v", serviceKey, resp)

	return resp.Node.Value, nil
}

func executeService(serviceURL string) (*ServiceExecutionResult, error) {
	res, err := http.Get(serviceURL)
	if err != nil {
		log.Printf("executor %s: failed to run http.Get() - %v", serviceURL, err)
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if res.StatusCode != 200 {
		log.Printf("executor %s: failed to execute service - %s", serviceURL, data)
		return nil, ErrServiceNotExecuted
	}

	var ser ServiceExecutionResult
	err = json.Unmarshal(data, &ser)
	if err != nil {
		log.Printf("executor %s: failed to unmarshal ServiceExecutionResult - %v", serviceURL, err)
		return nil, err
	}
	return &ser, nil
}
