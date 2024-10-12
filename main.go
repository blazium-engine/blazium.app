package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// MirrorListResponse represents the structure of the JSON response for the mirrorlist API.
type MirrorListResponse struct {
	Version string   `json:"version"`
	Mirrors []string `json:"mirrors"`
}

// LoadMirrors reads the mirrors from a JSON file and returns them as a slice of strings.
func LoadMirrors() ([]string, error) {
	// Construct the file path for mirrors.json
	filePath := filepath.Join("data", "mirrors.json")

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading mirrors file %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to read mirrors file: %v", err)
	}

	// Create a struct to unmarshal the JSON data
	var mirrorsData struct {
		Mirrors []string `json:"mirrors"`
	}

	// Parse the JSON file
	err = json.Unmarshal(data, &mirrorsData)
	if err != nil {
		log.Printf("Error parsing mirrors JSON file: %v", err)
		return nil, fmt.Errorf("failed to parse mirrors JSON: %v", err)
	}

	return mirrorsData.Mirrors, nil
}

func main() {
	// Create a new router using Gorilla Mux
	r := mux.NewRouter()

	// Serve index.html on the root path "/"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "./static/index.html")
	})

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Set the content type to application/json
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		// Define a health check response structure
		response := map[string]string{"status": "healthy"}

		// Encode the response as JSON and send it
		json.NewEncoder(w).Encode(response)
	})

	// Serve all static files from the "static" directory
	staticFileDirectory := http.Dir("./static")
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/static/").Handler(staticFileHandler)

	// API endpoint for /api/mirrorlist/:version/json
	r.HandleFunc("/api/mirrorlist/{version}/json", MirrorListHandler).Methods("GET")

	embedHandler := embedMiddleware(r)
	corsHandler := enableCORS(embedHandler)

	// Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}

// MirrorListHandler handles the /api/mirrorlist/:version/json endpoint
func MirrorListHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	version := vars["version"]

	// Load mirrors from the JSON file
	mirrors, err := LoadMirrors()
	if err != nil {
		log.Printf("Error loading mirrors: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create the response
	response := MirrorListResponse{
		Version: version,
		Mirrors: mirrors,
	}

	// Convert the response to JSON
	jsonResponse, err := json.Marshal(response)
	if err != nil {
		log.Printf("Error generating JSON response: %v", err)
		http.Error(w, "Error generating JSON", http.StatusInternalServerError)
		return
	}

	// Set content-type to application/json and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_, err = w.Write(jsonResponse)
	if err != nil {
		log.Printf("Error writing JSON response: %v", err)
	}
}

func enableCORS(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Set CORS headers
        w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins, you can restrict this to a specific domain
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        // Handle preflight OPTIONS requests
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }

        // Call the next handler
        next.ServeHTTP(w, r)
    })
}

func embedMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Get the User-Agent header and convert it to lowercase for case-insensitive comparison
        userAgent := strings.ToLower(r.Header.Get("User-Agent"))

        // Check if the User-Agent contains "discordbot" (case-insensitive)
        if strings.Contains(userAgent, "discordbot") {
            // Set appropriate headers for HTML content and caching
            w.Header().Set("Content-Type", "text/html; charset=utf-8")
            w.Header().Set("Cache-Control", "max-age=3600") // Cache the response for 1 hour

            // Write the Open Graph meta tags for Discord embeds
            w.Write([]byte(`
                <!DOCTYPE html>
                <html lang="en">
                <head>
                    <meta charset="UTF-8">
                    <meta name="viewport" content="width=device-width, initial-scale=1.0">
                    <meta property="og:title" content="Blazium Engine">
                    <meta property="og:description" content="Blazium Engine forked from Godot.">
                    <meta property="og:image" content="https://blazium.app/static/assets/logo.png">
                    <meta property="og:url" content="https://blazium.app">
                    <meta property="og:type" content="website">
                    <meta name="twitter:card" content="summary_large_image">
                    <meta property="og:site_name" content="Blazium Engine">
                    <title>Blazium Engine</title>
                </head>
                <body>
                    <h1>Welcome to Blazium Engine</h1>
                </body>
                </html>
            `))
            return
        }

        // If the User-Agent is not from Discord, pass the request to the next handler
        next.ServeHTTP(w, r)
    })
}