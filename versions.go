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
	"slices"
	"sync"
	"time"
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

type EditorDownloadOptions struct {
	Versions map[string][]string `json:"versions"`
	Options  map[string][]string `json:"options"`
	Commands map[string]string   `json:"commands"`
}

type ToolsDownloadOptions struct {
	Versions map[string][]string `json:"versions"`
	Names    map[string]string   `json:"names"`
	Os       []string            `json:"os"`
}

type EditorFilesDownloads map[string]map[string]map[string]int
type EditorFilesAnalytics struct {
	Timestamp            string
	EditorFilesDownloads EditorFilesDownloads
}

var (
	cacheMutex                 sync.RWMutex
	editorDownloadOptionsCache *EditorDownloadOptions
	toolsDownloadOptionsCache  *ToolsDownloadOptions
	editorFilesAnalyticsCache  *EditorFilesAnalytics

	editorFilesDownloads EditorFilesDownloads
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

	// Update editor download options cache
	var fileEditorOptions struct {
		Options  map[string][]string `json:"options"`
		Commands map[string]string   `json:"commands"`
	}

	filePath := filepath.Join("data", "editor_download_options.json")

	if err := readJSONFile(filePath, &fileEditorOptions); err != nil {
		log.Printf("Error reading %s: %v", filePath, err)
		return
	}

	versionsJson, err := getEditorVersions()
	if err != nil {
		log.Printf("Error fetching versions: %v", err)
		return
	}

	editorDownloadOptionsCache = &EditorDownloadOptions{
		Versions: versionsJson,
		Options:  fileEditorOptions.Options,
		Commands: fileEditorOptions.Commands,
	}

	// Update tools download options cache
	var fileToolsOptions struct {
		Names map[string]string `json:"names"`
		Os    []string          `json:"os"`
	}

	filePath = filepath.Join("data", "tools_download_options.json")

	if err := readJSONFile(filePath, &fileToolsOptions); err != nil {
		log.Printf("Error reading %s: %v", filePath, err)
		return
	}

	tools := make([]string, 0, len(fileToolsOptions.Names))
	for _, value := range fileToolsOptions.Names {
		tools = append(tools, value)
	}

	versionsJson, err = getToolsVersions(tools)
	if err != nil {
		log.Printf("Error fetching versions: %v", err)
		return
	}

	toolsDownloadOptionsCache = &ToolsDownloadOptions{
		Versions: versionsJson,
		Names:    fileToolsOptions.Names,
		Os:       fileToolsOptions.Os,
	}

	// Update editor file analytics
	editorFilesAnalyticsCache = &EditorFilesAnalytics{
		Timestamp:            time.Now().UTC().Format(time.DateTime),
		EditorFilesDownloads: editorFilesDownloads,
	}
	// Clean the memory then allocate
	editorFilesDownloads = nil
	editorFilesDownloads = make(EditorFilesDownloads)
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

// getEditorVersions fetches the version data for all build types
// and returns them in more manageable format.
func getEditorVersions() (map[string][]string, error) {
	versions := make(map[string][]string)
	buildTypes := []string{"nightly", "pre-release", "release"}

	var versionsData []VersionPayload
	var err error
	for _, buildType := range buildTypes {
		if os.Args[1] == "--local" {
			versionsData, err = localEditorVersions(buildType)
		} else {
			versionsData, err = fetchCerebroVersionData(buildType)
		}
		if err != nil {
			log.Printf("Error loading editor versions: %v", err)
			return map[string][]string{}, nil
		}
		for _, version := range versionsData {
			versions[buildType] = append(versions[buildType], version.Version)
		}
	}
	for i, versionList := range versions {
		slices.Reverse(versionList)
		versions[i] = versionList
	}
	return versions, nil
}

// getToolsVersions fetches the version data for all build types
// and returns them in more manageable format.
func getToolsVersions(tools []string) (map[string][]string, error) {
	versions := make(map[string][]string)

	var versionsData []ToolData
	var err error
	for _, tool := range tools {
		if os.Args[1] == "--local" {
			versionsData, err = localToolsVersions(tool, "windows")
		} else {
			versionsData, err = fetchCerebroTools(tool, "windows")
		}
		if err != nil {
			log.Printf("Error loading tool versions: %v", err)
			return map[string][]string{}, nil
		}
		for _, version := range versionsData {
			versions[tool] = append(versions[tool], version.Version)
		}
	}
	for i, versionList := range versions {
		slices.Reverse(versionList)
		versions[i] = versionList
	}
	return versions, nil
}

// Used for local editor versions fetch
func localEditorVersions(buildType string) ([]VersionPayload, error) {
	url := fmt.Sprintf("https://blazium.app/api/versions/data/%s", buildType)
	resp, err := http.Get(url)
	if err != nil {
		return []VersionPayload{}, fmt.Errorf("failed to fetch versions for %s: %w", buildType, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []VersionPayload{}, fmt.Errorf("received non-OK HTTP status for %s: %d", buildType, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []VersionPayload{}, fmt.Errorf("failed to read response body for %s: %w", buildType, err)
	}

	var versionsData []VersionPayload
	if err := json.Unmarshal(body, &versionsData); err != nil {
		return []VersionPayload{}, fmt.Errorf("failed to parse versions JSON for %s: %w", buildType, err)
	}
	return versionsData, nil
}

// Used for local tools versions fetch
func localToolsVersions(tool string, os string) ([]ToolData, error) {
	url := fmt.Sprintf("https://blazium.app/api/tools/%s/%s", tool, os)
	resp, err := http.Get(url)
	if err != nil {
		return []ToolData{}, fmt.Errorf("failed to fetch versions for %s: %w", tool, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []ToolData{}, fmt.Errorf("received non-OK HTTP status for %s: %d", tool, resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []ToolData{}, fmt.Errorf("failed to read response body for %s: %w", tool, err)
	}

	var versionsData []ToolData
	if err := json.Unmarshal(body, &versionsData); err != nil {
		return []ToolData{}, fmt.Errorf("failed to parse versions JSON for %s: %w", tool, err)
	}
	return versionsData, nil
}
