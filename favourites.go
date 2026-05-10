package main

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// getFavouritesPath returns the path to the favourites JSON file.
// Uses XDG_CONFIG_HOME if set, otherwise ~/.config/tailscale-portal/favourites.json
func getFavouritesPath() string {
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			home = "."
		}
		configDir = filepath.Join(home, ".config")
	}
	return filepath.Join(configDir, "tailscale-portal", "favourites.json")
}

// loadFavourites loads favourites from the JSON file.
// If the file doesn't exist, returns an empty slice with no error.
// If the file is corrupted, returns an empty slice with an error.
// Sorts favourites alphabetically by Name before returning.
func loadFavourites() ([]Favourite, error) {
	path := getFavouritesPath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return []Favourite{}, nil
		}
		return []Favourite{}, fmt.Errorf("failed to read favourites file: %w", err)
	}

	var file FavouritesFile
	if err := json.Unmarshal(data, &file); err != nil {
		return []Favourite{}, fmt.Errorf("favourites file corrupted, starting fresh: %w", err)
	}

	// Sort alphabetically by name
	sortFavourites(file.Favourites)
	return file.Favourites, nil
}

// saveFavourites saves favourites to the JSON file atomically.
func saveFavourites(favourites []Favourite) error {
	path := getFavouritesPath()
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	file := FavouritesFile{
		Version:    "1.0",
		Favourites: favourites,
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal favourites: %w", err)
	}

	// Write atomically: write to temp file, then rename
	tempPath := path + ".tmp"
	if err := os.WriteFile(tempPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write favourites file: %w", err)
	}
	if err := os.Rename(tempPath, path); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to save favourites file: %w", err)
	}

	return nil
}

// sortFavourites sorts favourites alphabetically by Name (case-insensitive).
func sortFavourites(favourites []Favourite) {
	sort.Slice(favourites, func(i, j int) bool {
		return strings.ToLower(favourites[i].Name) < strings.ToLower(favourites[j].Name)
	})
}

// generateID creates a simple random ID string.
func generateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
