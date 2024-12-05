package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gorilla/mux"
)

// MirrorListResponse represents the structure of the JSON response for the mirrorlist API.
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

type Version struct {
	Name     string    `json:"name"`
	Releases []Release `json:"releases"`
}

type Versions []Version

var templates *template.Template

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

// loadTemplates parses all templates in the basePath folder, including subfolders.
func loadTemplates(basePath string) error {
	// Define functions to be used in templates
	funcMap := template.FuncMap{
		// Creates a key:value pair from the arguments
		"dict": func(values ...any) map[string]any {
			dict := make(map[string]any)
			for i := 0; i < len(values); i += 2 {
				dict[values[i].(string)] = values[i+1]
			}
			return dict
		},
		// Creates a sequence of numbers, needed for loops in templates
		"seq": func(n int) []int {
			numbers := make([]int, n)
			for i := 0; i < n; i++ {
				numbers[i] = i + 1
			}
			return numbers
		},
		"add": func(x, y int) int { return x + y },
		"sub": func(x, y int) int { return x - y },
	}

	// Create a new template and associate the function map
	templates = template.New("").Funcs(funcMap)

	// Use ParseGlob to parse all .tmpl files in the basePath directories
	pattern := filepath.Join(basePath, "*/*.tmpl")
	_, err := templates.ParseGlob(pattern)
	if err != nil {
		return fmt.Errorf("error parsing templates: %w", err)
	}

	return nil
}

// serveTemplate renders an HTML template with the given data and writes it to the HTTP response.
func serveTemplate(w http.ResponseWriter, name string, data any) {
	err := templates.ExecuteTemplate(w, name, data)
	if err != nil {
		log.Printf("Error rendering template '%s': %v", name, err)
		http.Error(w, "Internal Server Error: Unable to render the requested page.", http.StatusInternalServerError)
	}
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

	// Serve download_options.json for the dropdowns
	r.HandleFunc("/download-options", func(w http.ResponseWriter, r *http.Request) {
		// Construct the file path for download_options.json
		filePath := filepath.Join("data", "download_options.json")

		// Read the JSON file
		data, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading download options file %s: %v", filePath, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		var downloadOptions struct {
			Options  map[string]any `json:"options"`
			Commands map[string]any `json:"commands"`
		}

		err = json.Unmarshal(data, &downloadOptions)
		if err != nil {
			log.Printf("Error parsing download options JSON file: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Set content-type to application/json and write the response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(downloadOptions)
	}).Methods("GET")

	// Handle 404
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "not_found", nil)
	})

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

	// Serve showcase.tmpl on the path "/showcase"
	// r.HandleFunc("/showcase", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "showcase", nil)
	// }).Methods("GET")

	// Serve showcase_article.tmpl on the path "/showcase/article"
	// r.HandleFunc("/showcase/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "showcase_article", nil)
	// }).Methods("GET")

	// Serve blog.tmpl on the path "/blog"
	// r.HandleFunc("/blog", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "blog", nil)
	// }).Methods("GET")

	// // Serve blog_article.tmpl on the path "/blog/article"
	// r.HandleFunc("/blog/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "blog_article", nil)
	// }).Methods("GET")

	// Serve road_maps.tmpl on the path "/road-maps"
	r.HandleFunc("/road-maps", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "road_maps", nil)
	}).Methods("GET")

	// Serve dev_tools.tmpl on the path "/dev-tools"
	r.HandleFunc("/dev-tools", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "dev_tools", nil)
	}).Methods("GET")

	// // Serve snippets.tmpl on the path "/snippets"
	// r.HandleFunc("/snippets", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "snippets", nil)
	// }).Methods("GET")

	// // Serve snippet_article.tmpl on the path "/snippets/article"
	// r.HandleFunc("/snippets/article", func(w http.ResponseWriter, r *http.Request) {
	// 	serveTemplate(w, "snippet_article", nil)
	// }).Methods("GET")

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

	// API endpoint for /api/mirrorlist/{version}.json
	// Format: 0.1.0.nightly.mono
	// Note: .mono is only on the mono-build of the Game Engine
	// URL: https://cdn.blazium.app/nightly/0.2.4/details.json
	r.HandleFunc("/api/mirrorlist/{version}.json", MirrorListHandler).Methods("GET")

	// Format: versions-nightly.json
	// Note: only for nightly,prerelease,release
	r.HandleFunc("/api/versions-{type}.json", VersionsHandler).Methods("GET")

	embedHandler := embedMiddleware(r)
	corsHandler := enableCORS(embedHandler)

	// Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
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
	// Load versions from the JSON file
	versions, err := LoadVersions()
	if err != nil {
		log.Printf("Error loading versions: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Set content-type to application/json and write the response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(versions)
}

func LoadVersions() (Versions, error) {
	// Construct the file path for versions.json
	filePath := filepath.Join("data", "versions.json")

	// Read the JSON file
	data, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading versions file %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to read versions file: %v", err)
	}

	// Parse the JSON file into the Versions struct
	var versions Versions
	err = json.Unmarshal(data, &versions)
	if err != nil {
		log.Printf("Error parsing versions JSON file: %v", err)
		return nil, fmt.Errorf("failed to parse versions JSON: %v", err)
	}

	return versions, nil
}
