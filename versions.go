package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
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
	Success bool          `json:"success"`
	Data    []VersionPayload `json:"data"`
}

type VersionPayload struct {
	DeployType   string `json:"deploy_type"`
	Version      string `json:"version"`
	ChangelogURL string `json:"changelog_url"`
	VersionURL   string `json:"version_url"`
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

	body, err := ioutil.ReadAll(resp.Body)
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

	body, err := ioutil.ReadAll(resp.Body)
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