package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
)

// Helper function to ensure directories exist
func ensureDirExists(path string) error {
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return fmt.Errorf("error creating directory '%s': %w", path, err)
	}
	return nil
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

type Card struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Link        string `json:"link"`
	Image       string `json:"image"`
}

// Generate templates for dev tools cards
func GenerateRoadMaps() error {
	filePath := filepath.Join("data", "road_maps.json")
	templatePath := filepath.Join("templates", "generators", "road_maps_cards.tmpl")
	outputPath := filepath.Join("templates", "runtime", "generated", "road_maps_cards.tmpl")

	var roadMaps []Card

	if err := readJSONFile(filePath, &roadMaps); err != nil {
		return fmt.Errorf("error loading road maps data: %w", err)
	}

	if err := executeTemplate(templatePath, "road-maps-cards", outputPath, roadMaps); err != nil {
		return fmt.Errorf("error generating road maps cards: %w", err)
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

// Generate all templates
func GenerateTemplates() error {
	dir := filepath.Join("templates", "runtime", "generated")
	if err := ensureDirExists(dir); err != nil {
		return fmt.Errorf("error ensuring generated templates directory: %w", err)
	}

	if err := GenerateRoadMaps(); err != nil {
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

	return nil
}
