package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
)

type FileDetails struct {
	Base FileInfo `json:"base"`
	Mono FileInfo `json:"mono"`
}

type MirrorListResponse struct {
	Version   string        `json:"version"`
	Timestamp string        `json:"timestamp"`
	Mirrors   []MirrorEntry `json:"mirrors"`
}

type MirrorEntry struct {
	Name     string   `json:"name"`
	Url      string   `json:"url"`
	Checksum Checksum `json:"checksum"`
	Filesize string   `json:"filesize"`
}

type Release struct {
	Name         string `json:"name"`
	ReleaseDate  string `json:"release_date"`
	ReleaseNotes string `json:"release_notes"`
}

type FileInfo struct {
	Name      string     `json:"name"`
	Filename  string     `json:"filename"`
	Filesize  string     `json:"filesize"`
	Checksum  Checksum   `json:"checksum"`
	URL       string     `json:"url"`
	Timestamp string     `json:"timestamp"`
	Mirrors   []FileInfo `json:"mirrors"`
}

type Checksum struct {
	SHA512 string `json:"512"`
	SHA256 string `json:"256"`
}

// https://cdn.blazium.app/nightly/0.2.4/templates.json

func MirrorListHandler(w http.ResponseWriter, r *http.Request) {
	// Extract the version from the URL
	vars := mux.Vars(r)
	version := vars["version"]

	// Split the version string
	versionParts := strings.Split(version, ".")
	if len(versionParts) < 4 {
		http.Error(w, "Invalid version format", http.StatusBadRequest)
		return
	}

	// Determine base version and version type
	baseVersion := strings.Join(versionParts[0:3], ".")
	versionType := versionParts[3]

	// Construct the details.json URL
	url := fmt.Sprintf("https://cdn.blazium.app/%s/%s/templates.json", versionType, baseVersion)

	// Make a HEAD request to check if details.json exists
	resp, err := http.Head(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		// If the file doesn't exist, return an empty response
		emptyResponse := MirrorListResponse{Version: version}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(emptyResponse)
		return
	}

	// Get the details.json file
	resp, err = http.Get(url)
	if err != nil || resp.StatusCode != http.StatusOK {
		http.Error(w, "Failed to fetch details.json", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, "Failed to read details.json", http.StatusInternalServerError)
		return
	}

	// Parse details.json
	var details FileDetails
	if err := json.Unmarshal(body, &details); err != nil {
		http.Error(w, "Failed to parse details.json", http.StatusInternalServerError)
		return
	}

	// Determine whether to use Mono or Base based on the version suffix
	var fileInfo FileInfo
	if len(versionParts) > 4 && versionParts[4] == "mono" {
		fileInfo = details.Mono
	} else {
		fileInfo = details.Base
	}

	// Populate the MirrorListResponse
	mirrorList := MirrorListResponse{
		Version:   version,
		Timestamp: fileInfo.Timestamp,
	}

	// Add Base or Mono as the first MirrorEntry
	baseOrMonoEntry := MirrorEntry{
		Name:     fileInfo.Name,
		Url:      fileInfo.URL,
		Checksum: fileInfo.Checksum,
		Filesize: fileInfo.Filesize,
	}
	mirrorList.Mirrors = append(mirrorList.Mirrors, baseOrMonoEntry)

	// Add any additional mirror entries
	if len(fileInfo.Mirrors) > 0 {
		for _, mirror := range fileInfo.Mirrors {
			mirrorEntry := MirrorEntry{
				Name:     mirror.Name,
				Url:      mirror.URL,
				Checksum: mirror.Checksum,
				Filesize: mirror.Filesize,
			}
			mirrorList.Mirrors = append(mirrorList.Mirrors, mirrorEntry)
		}
	}

	// Return the response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(mirrorList)
}
