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

// LoadMirrors reads the mirrors from a JSON file and returns them as a slice of strings.
func LoadMirrors() ([]string, error) {
	// Construct the file path for mirrors.json
	filePath := filepath.Join("data", "mirrors.json")

	// Create a struct to unmarshal the JSON data
	var mirrorsData struct {
		Mirrors []string `json:"mirrors"`
	}

	if err := readJSONFile(filePath, &mirrorsData); err != nil {
		return nil, fmt.Errorf("error loading mirrors data: %w", err)
	}

	return mirrorsData.Mirrors, nil
}

// serveTemplate renders an HTML template with the given data and writes it to the HTTP response.
func serveTemplate(w http.ResponseWriter, name string, data any) {
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Error rendering template '%s': %v", name, err)
		http.Error(w, "Internal Server Error: Unable to render the requested page.", http.StatusInternalServerError)
	}
}

// serve the data in JSON format
func serveJSON(w http.ResponseWriter, data any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Printf("Error serving JSON : %v", err)
		http.Error(w, "Internal Server Error: Unable to serve the requested JSON.", http.StatusInternalServerError)
	}
}

// Helper function to read and parse JSON file into a slice of structs
func readJSONFile[T any](filePath string, out *T) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("error reading file '%s': %w", filePath, err)
	}
	if err := json.Unmarshal(data, out); err != nil {
		return fmt.Errorf("error decoding JSON from file '%s': %w", filePath, err)
	}
	return nil
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

		// Check if the User-Agent contains "discordbot" or "twitterbot" (case-insensitive)
		if strings.Contains(userAgent, "discordbot") || strings.Contains(userAgent, "twitterbot") {
			// Set appropriate headers for caching
			w.Header().Set("Cache-Control", "max-age=3600") // Cache the response for 1 hour

			// Serve the embed template
			serveTemplate(w, "embed", nil)
			return
		}

		// If the User-Agent is not from a bot, pass the request to the next handler
		next.ServeHTTP(w, r)
	})
}

// VersionsHandler handles the /api/versions.json endpoint
func VersionsHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildType := vars["type"]

	versions, err := fetchCerebroVersions(buildType)
	if err != nil {
		log.Printf("Error loading versions: %v", err)
		serveJSON(w, []VersionData{})
		return
	}
	serveJSON(w, versions)
}

// VersionsHandler handles the /api/versions.json endpoint
func VersionDataHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	buildType := vars["buildType"]

	versions, err := fetchCerebroVersionData(buildType)
	if err != nil {
		log.Printf("Error loading versions: %v", err)
		serveJSON(w, []VersionPayload{})
		return
	}
	serveJSON(w, versions)
}

// DownloadOptionsHandler serves the cached download options.
func DownloadOptionsHandler(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if downloadOptionsCache == nil {
		http.Error(w, "Cache not available", http.StatusServiceUnavailable)
		serveJSON(w, DownloadOptions{})
		return
	}
	serveJSON(w, downloadOptionsCache)
}

func main() {
	// Generate templates from configs
	if err := GenerateTemplates(); err != nil {
		log.Fatalf("Error generating templates: %v", err)
	}

	// Load runtime templates
	err := loadTemplates("./templates/runtime")
	if err != nil {
		log.Fatalf("Error loading runtime templates: %v", err)
	}

	// Create a new router using Gorilla Mux
	r := mux.NewRouter()

	// Handle 404
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "not_found", nil)
	})

	// Serve robots.txt on the root path "/robots.txt"
	r.HandleFunc("/robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, filepath.Join("static", "robots.txt"))
	}).Methods("GET")

	// Serve main.tmpl on the root path "/"
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "home", nil)
	}).Methods("GET")

	// Redirect "/download" to "/download/prebuilt-binaries"
	r.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Location", "/download/prebuilt-binaries")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}).Methods("GET")

	// Serve the download page handling the different tabs
	r.HandleFunc("/download/{tab}", func(w http.ResponseWriter, r *http.Request) {
		// Get the template name from the URL
		vars := mux.Vars(r)
		page := vars["tab"]
		serveTemplate(w, page, nil)
	}).Methods("GET")

	// Serve road_maps.tmpl on the path "/road-maps"
	r.HandleFunc("/road-maps", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "road_maps", nil)
	}).Methods("GET")

	// Serve privacy_policy.tmpl on the path "/privacy-policy"
	r.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("data", "markdown", "privacy_policy.md")
		file, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file '%s': %v", filePath, err)
			http.Error(w, "Failed to read "+filePath, http.StatusInternalServerError)
			return
		}
		html := string(mdToHTML(file))
		serveTemplate(w, "privacy_policy", html)
	}).Methods("GET")

	// Serve terms_of_service.tmpl on the path "/terms-of-service"
	// r.HandleFunc("/terms-of-service", func(w http.ResponseWriter, r *http.Request) {
	// 	filePath := filepath.Join("data", "markdown", "terms_of_service.md")
	// 	file, err := os.ReadFile(filePath)
	// 	if err != nil {
	// 		log.Printf("Error reading file '%s': %v", filePath, err)
	// 		http.Error(w, "Failed to read "+filePath, http.StatusInternalServerError)
	// 		return
	// 	}
	// 	html := string(mdToHTML(file))
	// 	serveTemplate(w, "terms_of_service", html)
	// }).Methods("GET")

	// Serve blog.tmpl on the path "/blog"
	// r.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "blog", nil)
	// }).Methods("GET")

	// // Serve blog_article.tmpl on the path "/blog/article"
	// r.HandleFunc("/blog/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "blog_article", nil)
	// }).Methods("GET")

	// // Serve snippets.tmpl on the path "/snippets"
	// r.HandleFunc("/snippets", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "snippets", nil)
	// }).Methods("GET")

	// // Serve snippet_article.tmpl on the path "/snippets/article"
	// r.HandleFunc("/snippets/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "snippet_article", nil)
	// }).Methods("GET")

	// Serve showcase.tmpl on the path "/showcase"
	// r.HandleFunc("/showcase", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "showcase", nil)
	// }).Methods("GET")

	// Serve showcase_article.tmpl on the path "/showcase/article"
	// r.HandleFunc("/showcase/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "showcase_article", nil)
	// }).Methods("GET")

	// Serve dev_tools.tmpl on the path "/dev-tools"
	// r.HandleFunc("/dev-tools", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "dev_tools", nil)
	// }).Methods("GET")

	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		// Define a health check response structure
		response := map[string]string{"status": "healthy"}
		serveJSON(w, response)
	})

	// Serve all static files from the "static" directory
	staticFileDirectory := http.Dir("./static")
	staticFileHandler := http.StripPrefix("/static/", http.FileServer(staticFileDirectory))
	r.PathPrefix("/static/").Handler(staticFileHandler)

	// API endpoint for /api/mirrorlist/{version}.json
	// Format: 0.1.0.nightly.mono
	// Note: .mono is only on the mono-build of the Game Engine
	// URL: https://cdn.blazium.app/nightly/0.2.4/details.json
	r.HandleFunc("/api/mirrorlist/{version}.json", MirrorListHandler).Methods("GET")

	// Format: versions-nightly.json
	// Note: only for nightly, prerelease, release
	r.HandleFunc("/api/versions-{type}.json", VersionsHandler).Methods("GET")

	r.HandleFunc("/api/versions/data/{buildType}", VersionDataHandler).Methods("GET")

	// Serve download options for the dropdowns
	r.HandleFunc("/api/download-options", DownloadOptionsHandler).Methods("GET")

	embedHandler := embedMiddleware(r)
	corsHandler := enableCORS(embedHandler)

	// Start the background cache updater
	go startCacheUpdater()

	// Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
