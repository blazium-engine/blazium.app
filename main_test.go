package main

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// TestAPIEndpoint tests the /api/mirrorlist/:version/json endpoint.
func TestAPIEndpoint(t *testing.T) {
	// Create a request to pass to the handler.
	req, err := http.NewRequest("GET", "/api/mirrorlist/v1/json", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Add the Content-Type header to simulate a JSON request.
	req.Header.Set("Content-Type", "application/json")

	// Create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()

	// Create a new router using Gorilla Mux.
	r := setupRouter()

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check if the status code is 200 OK.
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	// Check if the Content-Type is application/json in the response.
	contentType := rr.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", contentType)
	}

	// Check if the response is valid JSON.
	var response MirrorListResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
		t.Errorf("Response was not valid JSON: %v", err)
	}

}

// TestRootPath tests the root / path.
func TestRootPath(t *testing.T) {
	// Create a request to pass to the handler.
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}

	// Create a ResponseRecorder to record the response.
	rr := httptest.NewRecorder()

	// Create a new router using Gorilla Mux.
	r := setupRouter()

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check if the status code is 200 OK.
	if rr.Code != http.StatusOK {
		t.Errorf("Expected status code 200, got %d", rr.Code)
	}

	// Convert the body to a string
	body := rr.Body.String()

	// Check if the body is non-empty.
	if len(body) == 0 {
		t.Fatalf("Expected non-empty body, but got an empty body")
	}

	// Parse the HTML response using goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		t.Fatalf("Could not parse HTML: %v", err)
	}

	// Validate that the document has the correct DOCTYPE
	bodyToTest := strings.ToLower(strings.TrimSpace(body))
	if !strings.HasPrefix(bodyToTest, "<!doctype html>") {
		t.Errorf("Expected response to start with <!DOCTYPE html>, got %s", body)
	}

	// Validate the presence of important HTML tags
	if doc.Find("html").Length() == 0 {
		t.Error("Expected to find <html> tag, but none found")
	}

	if doc.Find("head").Length() == 0 {
		t.Error("Expected to find <head> tag, but none found")
	}

	if doc.Find("body").Length() == 0 {
		t.Error("Expected to find <body> tag, but none found")
	}
}

// setupRouter sets up the router for testing.
func setupRouter() *mux.Router {
	// Generate templates from configs
	if err := GenerateTemplates(); err != nil {
		log.Fatalf("Error generating templates: %v", err)
	}

	// Load runtime templates
	err := loadTemplates("./templates/runtime")
	if err != nil {
		log.Fatalf("Error loading runtime templates: %v", err)
	}

	r := mux.NewRouter()

	// Serve main.tmpl on the root path "/"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "home", nil)
	}).Methods("GET")

	// Serve all static files from the "static" directory
	staticFileDirectory := http.Dir("./static")
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/static/").Handler(staticFileHandler)

	// API endpoint for /api/mirrorlist/{version}/json
	r.HandleFunc("/api/mirrorlist/{version}/json", MirrorListHandler).Methods("GET")

	return r
}
