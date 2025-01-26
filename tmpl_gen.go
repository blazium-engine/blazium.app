package main

import (
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

type Card struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Image       string `json:"image"`
	ButtonText  string `json:"buttonText"`
}

var templates *template.Template

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := html.CommonFlags | html.HrefTargetBlank | html.LazyLoadImages
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	return markdown.Render(doc, renderer)
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
		// Used to treat strings as HTML
		"toHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		// Used to get the blog id from the indieDB link
		"getBlogId": func(s string) string {
			parts := strings.Split(s, "/")
			parts = parts[len(parts)-2:]
			return strings.Join(parts, "/")
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

// Helper function to create a clean directory for the generated templates
func prepareDirectory(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("error removing contents of directory '%s': %w", path, err)
	}

	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory '%s': %w", path, err)
	}

	return nil
}

// Helper function to parse and execute a template
func executeTemplate(templatePath string, templateName string, outputPath string, data any) error {
	// Parse the template
	tmpl, err := template.ParseFiles(templatePath)
	if err != nil {
		return fmt.Errorf("error parsing template '%s': %w", templatePath, err)
	}

	// Create the output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("error creating output file '%s': %w", outputPath, err)
	}
	defer outputFile.Close()

	// Execute the template with the provided data
	if err := tmpl.ExecuteTemplate(outputFile, templateName, data); err != nil {
		return fmt.Errorf("error executing template '%s': %w", templateName, err)
	}

	return nil
}

// Generate templates for dev tools cards
func GenerateRoadMaps() error {
	filePath := filepath.Join("data", "roadmaps.json")
	templatePath := filepath.Join("templates", "generators", "roadmaps.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "roadmaps.tmpl")

	type Embed struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		EmbedLink   string `json:"embed_link"`
		Link        string `json:"link"`
		Id          string `json:"id"`
	}

	var roadMaps struct {
		Cards  []Card
		Embeds []Embed
	}

	if err := readJSONFile(filePath, &roadMaps); err != nil {
		return fmt.Errorf("error loading roadmaps data: %w", err)
	}

	if err := executeTemplate(templatePath, "roadmaps", outputPath, roadMaps); err != nil {
		return fmt.Errorf("error generating roadmaps: %w", err)
	}
	return nil
}

// Generate templates for dev tools cards
func GenerateGames() error {
	filePath := filepath.Join("data", "games.json")
	templatePath := filepath.Join("templates", "generators", "games_cards.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "games_cards.tmpl")

	var games []Card

	if err := readJSONFile(filePath, &games); err != nil {
		return fmt.Errorf("error loading games data: %w", err)
	}

	if err := executeTemplate(templatePath, "games-cards", outputPath, games); err != nil {
		return fmt.Errorf("error generating games cards: %w", err)
	}
	return nil
}

// Generate templates for dev tools cards
func GenerateDevTools() error {
	filePath := filepath.Join("data", "dev_tools.json")
	templatePath := filepath.Join("templates", "generators", "dev_tools_cards.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "dev_tools_cards.tmpl")

	var devTools []Card

	if err := readJSONFile(filePath, &devTools); err != nil {
		return fmt.Errorf("error loading dev tools data: %w", err)
	}

	if err := executeTemplate(templatePath, "dev-tools-cards", outputPath, devTools); err != nil {
		return fmt.Errorf("error generating dev tools cards: %w", err)
	}
	return nil
}

// Generate templates for digital store buttons
func GenerateDigitalStores() error {
	filePath := filepath.Join("data", "digital_stores.json")
	templatePath := filepath.Join("templates", "generators", "digital_store_buttons.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "digital_store_buttons.tmpl")

	var digitalStores []struct {
		Name  string `json:"name"`
		Link  string `json:"link"`
		Image string `json:"image"`
	}

	if err := readJSONFile(filePath, &digitalStores); err != nil {
		return fmt.Errorf("error loading digital stores data: %w", err)
	}

	if err := executeTemplate(templatePath, "digital-store-buttons", outputPath, digitalStores); err != nil {
		return fmt.Errorf("error generating digital store buttons: %w", err)
	}
	return nil
}

// Generate links templates
func GenerateLinks() error {
	filePath := filepath.Join("data", "links.json")
	templatePath := filepath.Join("templates", "generators", "links.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "links.tmpl")

	var links map[string]string

	if err := readJSONFile(filePath, &links); err != nil {
		return fmt.Errorf("error loading links data: %w", err)
	}

	if err := executeTemplate(templatePath, "links", outputPath, links); err != nil {
		return fmt.Errorf("error generating links: %w", err)
	}
	return nil
}

// Generate release card template
func GenerateReleaseCard() error {
	filePath := filepath.Join("data", "release_card.json")
	templatePath := filepath.Join("templates", "generators", "release_card.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "release_card.tmpl")

	var card Card

	if err := readJSONFile(filePath, &card); err != nil {
		return fmt.Errorf("error loading release card data: %w", err)
	}

	if err := executeTemplate(templatePath, "release-card", outputPath, card); err != nil {
		return fmt.Errorf("error generating release card: %w", err)
	}
	return nil
}

// Generate all templates
func GenerateTemplates() error {
	dir := filepath.Join("templates", "runtime", "generated")
	if err := prepareDirectory(dir); err != nil {
		return fmt.Errorf("error preparing generated templates directory: %w", err)
	}

	if err := GenerateRoadMaps(); err != nil {
		return err
	}
	if err := GenerateGames(); err != nil {
		return err
	}
	if err := GenerateDevTools(); err != nil {
		return err
	}
	if err := GenerateDigitalStores(); err != nil {
		return err
	}
	if err := GenerateLinks(); err != nil {
		return err
	}
	if err := GenerateReleaseCard(); err != nil {
		return err
	}

	return nil
}
