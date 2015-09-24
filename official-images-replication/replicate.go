package main

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/fsouza/go-dockerclient"
)

// Replication is an abstraction for images replication.
type Replication struct {
	registryURL string
	username    string
	password    string
}

const (
	endpoint = "unix:///var/run/docker.sock"
)

var dockerClient *docker.Client
var dockerHub DockerHub

func init() {
	//os.Setenv("HTTP_PROXY", "http://web-proxy.atl.hp.com:8080")
	client, err := docker.NewClient(endpoint)
	dockerClient = client
	if err != nil {
		// If error happend when create docer.Client, the program should stop and exit
		log.Fatalf("Failed to initiate docker client: %v", err)
	}
}

// Replicate function replicates the official images from hub.docker.com to private registry
// requiredRepos: such as "centos,busybox,ubuntu"
func (r *Replication) Replicate(filter string, repoNames string) error {
	log.Printf("===official-images-replication=== STARTING: filter=%s, repoNames=%s", filter, repoNames)
	start := time.Now()
	repos, err := dockerHub.GetRepos()
	if err != nil {
		return err
	}
	countOfRepos := len(repos)
	var countOfTags int
	for _, repo := range repos {
		log.Printf("repository: %s\n", repo.Name)
		if !include(filter, repoNames, repo.Name) {
			continue
		}
		tags, err := dockerHub.GetTags(repo.Name)
		if err != nil {
			continue
		}
		countOfTags += len(tags)
		for _, tag := range tags {
			err := r.pullImage(repo.Name, tag.Name)
			if err != nil {
				continue
			}
			err = r.tagImage(repo.Name, tag.Name)
			if err != nil {
				continue
			}
			err = r.pushImage(repo.Name, tag.Name)
			if err != nil {
				continue
			}
		}
	}
	duration := time.Now().Sub(start)
	log.Printf("===official-images-replication=== DONE. \n\tTotal Repositories: %d\n\tTotal Tags: %d\n\tTotal Duration: %s",
		countOfRepos, countOfTags, fmtDuration(duration))
	return nil
}
func include(filter string, repoNames string, repoName string) bool {
	if len(repoNames) == 0 {
		return true
	}

	if filter == "list" {
		// centos,busybox,ubuntu => ,centos,busybox,ubuntu,
		if strings.Contains(","+repoNames+",", ","+repoName+",") {
			return true
		}
		return false
	}

	if filter == "letter" {
		matched, err := regexp.MatchString("^["+repoNames+"].*", repoName)
		if err == nil {
			return matched
		}
		return true
	}
	return true
}

// Example of parameters:
// 		repo => busybox
// 		tag => ubuntu-14.04
func (r *Replication) pullImage(repo string, tag string) error {
	t := time.Now().Format("2006/01/02 15:04:05")
	fmt.Printf("\t%s pull %s:%s", t, repo, tag)
	start := time.Now()
	err := dockerClient.PullImage(docker.PullImageOptions{Repository: repo + ":" + tag}, docker.AuthConfiguration{})
	duration := time.Now().Sub(start)
	if err != nil {
		// If error happend here, the program should stop and exit
		fmt.Printf(" (%s), FAIL: %v\n", fmtDuration(duration), err)
		return err
	}
	fmt.Printf(" (%s), ", fmtDuration(duration))
	return nil
}

// Example of parameters:
// 		repo => busybox
// 		tag => ubuntu-14.04
func (r *Replication) tagImage(repo string, tag string) error {
	//example => localhost:5000/busybox
	newRepo := r.registryURL + "/" + repo
	err := dockerClient.TagImage(repo+":"+tag, docker.TagImageOptions{
		Repo:  newRepo,
		Tag:   tag,
		Force: true,
	})
	if err != nil {
		fmt.Printf("tag %s:%s, FAIL: %v\n", newRepo, tag, err)
		return err
	}
	return nil
}

func (r *Replication) pushImage(repo string, tag string) error {
	start := time.Now()
	//example => localhost:5000/busybox
	newRepo := r.registryURL + "/" + repo
	fmt.Printf("push %s:%s ", newRepo, tag)
	auth := docker.AuthConfiguration{}
	auth.Username = r.username
	auth.Password = r.password
	err := dockerClient.PushImage(docker.PushImageOptions{Name: newRepo, Tag: tag}, auth)
	duration := time.Now().Sub(start)
	if err != nil {
		fmt.Printf("(%s), FAIL: %v\n", fmtDuration(duration), err)
		return err
	}
	fmt.Printf("(%s)\n", fmtDuration(duration))
	return nil
}

// fmtDuration returns a string representing d in the form "87.00s".
func fmtDuration(d time.Duration) string {
	return fmt.Sprintf("%.2fs", d.Seconds())
}
