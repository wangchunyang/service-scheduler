package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	reposURL = "https://hub.docker.com/v2/repositories/library/?page_size=200&page=1"
	tagsURL  = "https://hub.docker.com/v2/repositories/library/%s/tags/?page_size=200&page=1"
)

// DockerHub is a proxy for http://hub.docker.com
type DockerHub struct {
}

// RepositoriesResponse defines data structure for JSON response of "Listing Repositories"
type RepositoriesResponse struct {
	Next     string       `json:"next"`
	Previous string       `json:"previous"`
	Results  []Repository `json:"results"`
}

// Repository defines data structure for each repository in RepositoriesResponse.
type Repository struct {
	User            string `json:"user"`
	Name            string `json:"name"`
	Namespace       string `json:"namespace"`
	Status          int64  `json:"status"`
	Description     string `json:"description"`
	FullDescription string `json:"full_description"`
	IsPrivate       bool   `json:"is_private"`
	IsAutomated     bool   `json:"is_automated"`
	CanEdit         bool   `json:"can_edit"`
	StarCount       int64  `json:"star_count"`
	PullCount       int64  `json:"pull_count"`
	LastUpdated     string `json:"last_updated"`
}

// TagsResponse defines data structure for JSON response of "Listing Image Tags"
type TagsResponse struct {
	Count    int64  `json:"count"`
	Next     string `json:"next"`
	Previous string `json:"previous"`
	Results  []Tag  `json:"results"`
}

// Tag defines data structure for each tag in TagsResponse
type Tag struct {
	Name        string `json:"name"`
	FullSize    int64  `json:"full_size"`
	ID          int64  `json:"id"`
	Repository  int64  `json:"repository"`
	Creator     int64  `json:"creator"`
	LastUpdater int64  `json:"last_updater"`
	ImageID     string `json:"image_id"`
	V2          bool   `json:"v2"`
}

// GetRepos function is listing repositories on http://hub.docker.com
func (hub *DockerHub) GetRepos() ([]Repository, error) {
	//reqURL :=https://hub.docker.com/v2/repositories/library/?page_size=200&page=1
	res, err := http.Get(reposURL)
	if err != nil {
		log.Printf("Failed to call REST api %s: %s", reposURL, err)
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Failed to get ropos: %s", err)
		return nil, err
	}
	//fmt.Printf("%s\n", robots)

	var reposRes RepositoriesResponse

	err = json.Unmarshal(data, &reposRes)
	return reposRes.Results, nil
}

// GetTags function is listing image tags on http://hub.docker.com
func (hub *DockerHub) GetTags(repo string) ([]Tag, error) {
	//reqURL := "https://hub.docker.com/v2/repositories/library/" + repo + "/tags/?page_size=200&page=1"
	reqURL := fmt.Sprintf(tagsURL, repo)
	res, err := http.Get(reqURL)
	if err != nil {
		log.Printf("Failed to call REST api: %s, %s", reqURL, err)
		return nil, err
	}

	data, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		log.Printf("Failed to get image tags: %s", err)
		return nil, err
	}
	var tagsRes TagsResponse
	err = json.Unmarshal(data, &tagsRes)
	return tagsRes.Results, nil
}
