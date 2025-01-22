package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/gorilla/mux"
)

type VersionData struct {
	Name     string    `json:"name"`
	Releases []Release `json:"releases"`
}

type VersionResponse struct {
	Data    []VersionData `json:"data"`
	Success bool          `json:"success"`
}

type ResponsePayload struct {
	Success bool             `json:"success"`
	Data    []VersionPayload `json:"data"`
}

type ToolData struct {
	Name    string `json:"name"`
	Type    string `json:"type"`
	Version string `json:"version"`
	OS      string `json:"os"`
	File    string `json:"file"`
	Sig     string `json:"sig"`
}

type VersionPayload struct {
	DeployType   string `json:"deploy_type"`
	Version      string `json:"version"`
	ChangelogURL string `json:"changelog_url"`
	VersionURL   string `json:"version_url"`
}

type DownloadOptions struct {
	Versions map[string][]string `json:"versions"`
	Options  map[string][]string `json:"options"`
	Commands map[string]string   `json:"commands"`
}

var (
	downloadOptionsCache *DownloadOptions
	cacheMutex           sync.RWMutex
)

func fetchCerebroTools(toolType string, osType string) ([]ToolData, error) {
	cerebroURL := os.Getenv("CEREBRO_URL")
	blaziumAuth := os.Getenv("BLAZIUM_AUTH")

	if cerebroURL == "" || blaziumAuth == "" {
		return nil, errors.New("CEREBRO_URL or BLAZIUM_AUTH environment variable is not set")
	}

	url := fmt.Sprintf("%s/api/v1/tools/%s/%s", cerebroURL, toolType, osType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("BLAZIUM_AUTH", blaziumAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse []ToolData
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return apiResponse, nil
}

func handleFetchCerebroTools(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	toolType := vars["toolType"]
	osType := vars["osType"]

	if toolType == "" || osType == "" {
		http.Error(w, "Tool type, OS type are required", http.StatusBadRequest)
		return
	}

	versionData, err := fetchCerebroTools(toolType, osType)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch version data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionData)
}

func fetchCerebroVersions(buildType string) ([]VersionData, error) {
	cerebroURL := os.Getenv("CEREBRO_URL")
	blaziumAuth := os.Getenv("BLAZIUM_AUTH")

	if cerebroURL == "" || blaziumAuth == "" {
		return nil, errors.New("CEREBRO_URL or BLAZIUM_AUTH environment variable is not set")
	}

	url := fmt.Sprintf("%s/api/v1/versions/%s", cerebroURL, buildType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("BLAZIUM_AUTH", blaziumAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse VersionResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if !apiResponse.Success || apiResponse.Data == nil {
		return []VersionData{}, nil
	}

	return apiResponse.Data, nil
}

func handleFetchCerebroToolData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	toolType := vars["toolType"]
	osType := vars["osType"]
	toolVersion := vars["toolVersion"]

	if toolType == "" || osType == "" || toolVersion == "" {
		http.Error(w, "Tool type, OS type, and tool version are required", http.StatusBadRequest)
		return
	}

	versionData, err := fetchCerebroToolData(toolType, osType, toolVersion)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to fetch version data: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(versionData)
}

func fetchCerebroToolData(toolType string, osType string, toolVersion string) (*ToolData, error) {
	cerebroURL := os.Getenv("CEREBRO_URL")
	blaziumAuth := os.Getenv("BLAZIUM_AUTH")

	if cerebroURL == "" || blaziumAuth == "" {
		return nil, errors.New("CEREBRO_URL or BLAZIUM_AUTH environment variable is not set")
	}

	url := fmt.Sprintf("%s/api/v1/tools/%s/%s/%s", cerebroURL, toolType, osType, toolVersion)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("BLAZIUM_AUTH", blaziumAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse ToolData
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	return &apiResponse, nil
}

func fetchCerebroVersionData(buildType string) ([]VersionPayload, error) {
	cerebroURL := os.Getenv("CEREBRO_URL")
	blaziumAuth := os.Getenv("BLAZIUM_AUTH")

	if cerebroURL == "" || blaziumAuth == "" {
		return nil, errors.New("CEREBRO_URL or BLAZIUM_AUTH environment variable is not set")
	}

	url := fmt.Sprintf("%s/api/v1/data/versions/%s", cerebroURL, buildType)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("BLAZIUM_AUTH", blaziumAuth)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var apiResponse ResponsePayload
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if !apiResponse.Success || apiResponse.Data == nil {
		return []VersionPayload{}, nil
	}

	return apiResponse.Data, nil
}

// updateCache reads the options the JSON file
// and adds the available versions.
func updateCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()

	var fileOptions struct {
		Options  map[string][]string `json:"options"`
		Commands map[string]string   `json:"commands"`
	}

	filePath := filepath.Join("data", "download_options.json")

	if err := readJSONFile(filePath, &fileOptions); err != nil {
		log.Printf("Error reading download_options.json: %v", err)
		return
	}

	versionsJson, err := getVersions()
	if err != nil {
		log.Printf("Error fetching versions: %v", err)
		return
	}

	downloadOptionsCache = &DownloadOptions{
		Versions: versionsJson,
		Options:  fileOptions.Options,
		Commands: fileOptions.Commands,
	}
}

// startCacheUpdater starts a ticker to update the cache every 30 minutes
func startCacheUpdater() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	// Update the cache initially
	updateCache()

	for range ticker.C {
		updateCache()
	}
}

// getVersions fetches the version data for all build types
// and returns them in more managable format.
func getVersions() (map[string][]string, error) {
	versions := make(map[string][]string)
	buildTypes := []string{"nightly", "pre-release", "release"}

	for _, buildType := range buildTypes {
		versionsData, err := fetchCerebroVersionData(buildType)
		if err != nil {
			log.Printf("Error loading versions: %v", err)
			return map[string][]string{}, nil
		}
		for _, version := range versionsData {
			versions[buildType] = append(versions[buildType], version.Version)
		}
	}
	return versions, nil
}
