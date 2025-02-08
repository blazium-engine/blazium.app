package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
	"github.com/quic-go/quic-go/http3"
)

type ArticleData struct {
	Image     string
	Title     string
	Published string
	Link      string
}

type MetaTags struct {
	Title       string
	Description string
	Url         string
	Image       string
}

type BlogArticle struct {
	MetaTags    MetaTags
	ArticleData ArticleData
}

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

// serve a markdown article with meta tags
func serveMarkdown(w http.ResponseWriter, filePath string, metaTags MetaTags) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		log.Printf("Error reading file '%s': %v", filePath, err)
		http.Error(w, "Failed to read "+filePath, http.StatusInternalServerError)
		return
	}
	html := string(mdToHTML(file))

	data := struct {
		MetaTags MetaTags
		Content  string
	}{
		MetaTags: metaTags,
		Content:  html,
	}

	serveTemplate(w, "md_article", data)
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow all origins, you can restrict this to a specific domain
		w.Header().Set("Access-Control-Allow-Methods", "HEAD, GET, POST, OPTIONS, PUT, DELETE")
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
		userAgent := strings.ToLower(r.Header.Get("User-Agent"))
		if strings.Contains(userAgent, "bot") {
			// Set appropriate headers for caching
			w.Header().Set("Cache-Control", "max-age=3600") // Cache the response for 1 hour
		}
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

// EditorDownloadOptionsHandler serves the cached editor download options.
func EditorDownloadOptionsHandler(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if editorDownloadOptionsCache == nil {
		http.Error(w, "Cache not available", http.StatusServiceUnavailable)
		return
	}
	serveJSON(w, editorDownloadOptionsCache)
}

// ToolsDownloadOptionsHandler serves the cached tools download options.
func ToolsDownloadOptionsHandler(w http.ResponseWriter, r *http.Request) {
	cacheMutex.RLock()
	defer cacheMutex.RUnlock()

	if toolsDownloadOptionsCache == nil {
		http.Error(w, "Cache not available", http.StatusServiceUnavailable)
		return
	}
	serveJSON(w, toolsDownloadOptionsCache)
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

	serveJSON(w, versionData)
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

	serveJSON(w, versionData)
}

// BlogHandler handles the /blog endpoint
func BlogHandler(w http.ResponseWriter, r *http.Request) {
	// if not htmx request or blog button pressed, serve base page
	if r.Header.Get("hx-request") != "true" || r.Header.Get("hx-trigger-name") == "blog-btn" {
		articleType := "articles"
		keyWord := ""
		page := ""
		if r.URL.Query().Has("t") {
			articleType = r.URL.Query().Get("t")
		}
		if r.URL.Query().Has("s") {
			keyWord = r.URL.Query().Get("s")
		}
		if r.URL.Query().Has("p") {
			page = r.URL.Query().Get("p")
		}

		data := struct {
			ArticleType string
			KeyWord     string
			Page        string
		}{
			ArticleType: articleType,
			KeyWord:     keyWord,
			Page:        page,
		}

		serveTemplate(w, "blog", data)
		return
	}

	// if htmx request, serve blog articles
	url := "https://www.indiedb.com/engines/blazium-engine"

	if r.URL.Query().Has("t") {
		articleType := r.URL.Query().Get("t")
		url += "/" + articleType
	} else {
		url += "/articles"
	}
	if r.URL.Query().Has("p") {
		if page := r.URL.Query().Get("p"); page != "" {
			url += "/page/" + page
		}
	}
	if r.URL.Query().Has("s") {
		keyWord := r.URL.Query().Get("s")
		url += "?filter=t&kw=" + keyWord
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}

	client := &http.Client{
		Transport: &http3.Transport{},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to make request: %v", err)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	var articles []ArticleData

	doc.Find("div.table div.row.rowcontent").Each(func(i int, s *goquery.Selection) {
		image, _ := s.Find("img").Attr("src")
		image = strings.ReplaceAll(image, "/cache", "")
		image = strings.ReplaceAll(image, "/crop_120x90", "")

		title := s.Find("h4").Text()

		published := s.Find("span.date time").Text()

		link, _ := s.Find("a.image").Attr("href")

		article := ArticleData{
			Image:     image,
			Title:     title,
			Published: published,
			Link:      link,
		}
		articles = append(articles, article)
	})

	pagination := doc.Find("div.pagination div.pages")
	currentPage, _ := strconv.Atoi(pagination.Find("span.current").Text())
	pagesAmount, _ := strconv.Atoi(pagination.Children().Last().Text())

	data := struct {
		Articles   []ArticleData
		Pagination map[string]int
	}{
		Articles:   articles,
		Pagination: map[string]int{"CurrentPage": currentPage, "PagesAmount": pagesAmount},
	}

	serveTemplate(w, "blogs-articles", data)

}

// BlogArticleHandler handles the /blog/article endpoint
func BlogArticleHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	articleType := vars["type"]
	id := vars["id"]

	url := fmt.Sprintf("https://www.indiedb.com/groups/indiedb/%s/%s", articleType, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("failed to create request: %v", err)
	}

	client := &http.Client{
		Transport: &http3.Transport{},
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("failed to make request: %v", err)
	}

	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	article := doc.Find("article #readarticle")

	image, exists := doc.Find("article meta[itemprop=image]").Attr("content")
	if !exists {
		log.Fatal("image not found in article")
	}

	indieDBLink, exists := doc.Find("article meta[itemprop=mainEntityOfPage]").Attr("itemid")
	if !exists {
		log.Fatal("indiedb link not found in article")
	}

	title := article.Find("div.title span.heading").Text()
	published := article.Find("p.introduction").Text()
	description := article.Find("p.introductiontext").Text()

	// Process each iframe
	article.Find("iframe").Each(func(i int, iframe *goquery.Selection) {
		// Create the wrapping div
		iframe.WrapHtml("<div class='iframe-placeholder'></div>")
		// Add the section sibling
		iframe.AfterHtml(`<section><p>
				We need your consent to show this embed, by clicking <strong>"Accept"</strong>, you agree to the use of cookies.
				This will activate <strong>all</strong> embeds.
				For more information, please review our <a href="/privacy-policy">Privacy Policy</a>.</p>
				<button type="button" class="secondary-btn" onclick="acceptCookies()">Accept</button>
			</section>
		`)
	})

	articleContent := article.Find("#articlecontent")

	// Remove preview image in article content, only useful in IndeDB
	selector := `p:has(img:only-child):first-of-type,
		h1:has(img:only-child):first-of-type,
		h2:has(img:only-child):first-of-type,
		h3:has(img:only-child):first-of-type,
		h3:has(img:only-child):first-of-type`
	articleContent.Find(selector).Remove()

	content, err := articleContent.Html()
	if err != nil {
		log.Fatal(err)
	}

	data := BlogArticle{
		MetaTags: MetaTags{
			Image:       image,
			Title:       "Blazium Engine - " + title,
			Description: description,
			Url:         fmt.Sprintf("/blog/article/%s/%s", articleType, id),
		},
		ArticleData: ArticleData{
			Title:     title,
			Published: published,
			Image:     content, // recycling Image for the content string
			Link:      indieDBLink,
		},
	}

	serveTemplate(w, "blog_article", data)
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
	r.HandleFunc("/roadmaps", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "roadmaps", nil)
	}).Methods("GET")

	// Serve privacy_policy.md on the path "/privacy-policy"
	r.HandleFunc("/privacy-policy", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("data", "articles", "privacy_policy.md")
		metaTags := MetaTags{
			Title:       "Blazium Engine - Privacy policy",
			Description: "Blazium website's privacy policy",
			Url:         "/privacy-policy",
		}
		serveMarkdown(w, filePath, metaTags)
	}).Methods("GET")

	// Serve terms_of_service.md on the path "/terms-of-service"
	r.HandleFunc("/terms-of-service", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("data", "articles", "terms_of_service.md")
		metaTags := MetaTags{
			Title:       "Blazium Engine - Terms of service",
			Description: "Blazium website's terms of service",
			Url:         "/terms-of-service",
		}
		serveMarkdown(w, filePath, metaTags)
	}).Methods("GET")

	// Serve licenses.tmpl on the path "/licenses"
	r.HandleFunc("/licenses", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("data", "articles", "licenses.md")
		metaTags := MetaTags{
			Title:       "Blazium Engine - Licenses",
			Description: "Blazium Engine and website licenses",
			Url:         "/licenses",
		}
		serveMarkdown(w, filePath, metaTags)
	}).Methods("GET")

	// Serve what_is_blazium.tmpl on the path "/what-is-blazium"
	r.HandleFunc("/what-is-blazium", func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join("data", "articles", "what_is_blazium.md")
		metaTags := MetaTags{
			Title:       "Blazium Engine - What is Blazium?",
			Description: "A game engine for 2D and 3D, Free and Open-Source, easy to use, there is more but not enough space here",
			Url:         "/what-is-blazium",
		}
		serveMarkdown(w, filePath, metaTags)
	}).Methods("GET")

	// Serve brand_kit.tmpl on the path "/brand-kit"
	r.HandleFunc("/brand-kit", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "brand_kit", nil)
	}).Methods("GET")

	// Serve dev_tools.tmpl on the path "/dev-tools"
	r.HandleFunc("/dev-tools", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "dev_tools", nil)
	}).Methods("GET")

	// Serve dev_tools.tmpl on the path "/dev-tools/download"
	r.HandleFunc("/dev-tools/download", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "dev_tools_download", nil)
	}).Methods("GET")

	// Serve games.tmpl on the path "/games"
	r.HandleFunc("/games", func(w http.ResponseWriter, r *http.Request) {
		serveTemplate(w, "games", nil)
	}).Methods("GET")

	// Serve a game page on the path "/games/{gameName}"
	r.HandleFunc("/games/{gameName}", func(w http.ResponseWriter, r *http.Request) {
		// Get the game name from the URL
		vars := mux.Vars(r)
		gameName := vars["gameName"]

		filePath := filepath.Join("data", "articles", "games", gameName+".md")
		file, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("Error reading file '%s': %v", filePath, err)
			http.Error(w, "Failed to read "+filePath, http.StatusInternalServerError)
			return
		}
		content := string(mdToHTML(file))

		// Get metadata
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(content))
		if err != nil {
			log.Fatal(err)
		}
		img, _ := doc.Find("meta[name='cover-image']").Attr("content")
		description, _ := doc.Find("meta[name='short-description']").Attr("content")
		title, _ := doc.Find("meta[name='game-name']").Attr("content")

		data := BlogArticle{
			MetaTags: MetaTags{
				Image:       img,
				Title:       "Blazium Engine - " + title,
				Description: description,
				Url:         "/games/" + gameName,
			},
			ArticleData: ArticleData{
				Title: title,
				Image: content, // Recycling Image for the content string
			},
		}

		serveTemplate(w, "blog_article", data)
	}).Methods("GET")

	// Serve blog.tmpl on the path "/blog"
	r.HandleFunc("/blog", BlogHandler).Methods("GET")

	// Serve blog_article.tmpl on the path "/blog/article"
	r.HandleFunc("/blog/article/{type}/{id}", BlogArticleHandler).Methods("GET")

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

	r.HandleFunc("/api/tools/{toolType}/{osType}", handleFetchCerebroTools).Methods("GET")
	r.HandleFunc("/api/tools/{toolType}/{osType}/{toolVersion}", handleFetchCerebroToolData).Methods("GET")

	// Serve download options for the editor download dropdowns
	r.HandleFunc("/api/download-options/editor", EditorDownloadOptionsHandler).Methods("GET")
	// Serve download options for the tools download dropdowns
	r.HandleFunc("/api/download-options/tools", ToolsDownloadOptionsHandler).Methods("GET")

	embedHandler := embedMiddleware(r)
	corsHandler := enableCORS(embedHandler)

	// Start the background cache updater
	go startCacheUpdater()

	// Start the server
	fmt.Println("Starting server on :8080")
	log.Fatal(http.ListenAndServe(":8080", corsHandler))
}
