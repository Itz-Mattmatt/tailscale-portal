package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetFavouritesPath(t *testing.T) {
	path := getFavouritesPath()
	if path == "" {
		t.Error("Expected non-empty path")
	}
	if !filepath.IsAbs(path) {
		t.Error("Expected absolute path")
	}
}

func TestLoadFavouritesEmpty(t *testing.T) {
	// Use a temp directory to avoid touching real config
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}()

	favourites, err := loadFavourites()
	if err != nil {
		t.Errorf("Expected no error for missing file, got %v", err)
	}
	if len(favourites) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(favourites))
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}()

	favourites := []Favourite{
		{ID: "1", Name: "Alpha", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
		{ID: "2", Name: "Beta", Port: 8080, Target: "http://localhost:8080", Path: "/api", Protocol: "http", Mode: "funnel"},
	}

	if err := saveFavourites(favourites); err != nil {
		t.Fatalf("Failed to save favourites: %v", err)
	}

	loaded, err := loadFavourites()
	if err != nil {
		t.Fatalf("Failed to load favourites: %v", err)
	}

	if len(loaded) != 2 {
		t.Fatalf("Expected 2 favourites, got %d", len(loaded))
	}

	if loaded[0].Name != "Alpha" {
		t.Errorf("Expected first favourite to be Alpha, got %s", loaded[0].Name)
	}
	if loaded[1].Name != "Beta" {
		t.Errorf("Expected second favourite to be Beta, got %s", loaded[1].Name)
	}

	if loaded[0].Port != 443 {
		t.Errorf("Expected port 443, got %d", loaded[0].Port)
	}
	if loaded[0].Protocol != "https" {
		t.Errorf("Expected protocol https, got %s", loaded[0].Protocol)
	}
}

func TestLoadFavouritesSortsAlphabetically(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}()

	favourites := []Favourite{
		{ID: "1", Name: "Zebra", Port: 443, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
		{ID: "2", Name: "alpha", Port: 8080, Target: "http://localhost:8080", Path: "/api", Protocol: "http", Mode: "funnel"},
		{ID: "3", Name: "Apple", Port: 3000, Target: "http://localhost:3000", Path: "/", Protocol: "https", Mode: "serve"},
	}

	if err := saveFavourites(favourites); err != nil {
		t.Fatalf("Failed to save favourites: %v", err)
	}

	loaded, err := loadFavourites()
	if err != nil {
		t.Fatalf("Failed to load favourites: %v", err)
	}

	// Should be sorted case-insensitively: alpha, Apple, Zebra
	expected := []string{"alpha", "Apple", "Zebra"}
	for i, exp := range expected {
		if loaded[i].Name != exp {
			t.Errorf("Expected favourite %d to be %s, got %s", i, exp, loaded[i].Name)
		}
	}
}

func TestLoadFavouritesCorrupted(t *testing.T) {
	oldHome := os.Getenv("HOME")
	oldXDG := os.Getenv("XDG_CONFIG_HOME")
	tmpDir := t.TempDir()
	os.Setenv("HOME", tmpDir)
	os.Unsetenv("XDG_CONFIG_HOME")
	defer func() {
		os.Setenv("HOME", oldHome)
		os.Setenv("XDG_CONFIG_HOME", oldXDG)
	}()

	path := getFavouritesPath()
	dir := filepath.Dir(path)
	os.MkdirAll(dir, 0755)
	os.WriteFile(path, []byte("not valid json"), 0644)

	favourites, err := loadFavourites()
	if err == nil {
		t.Error("Expected error for corrupted file")
	}
	if len(favourites) != 0 {
		t.Errorf("Expected empty slice for corrupted file, got %d items", len(favourites))
	}
}

func TestGenerateID(t *testing.T) {
	id1 := generateID()
	id2 := generateID()

	if id1 == "" {
		t.Error("Expected non-empty ID")
	}
	if id1 == id2 {
		t.Error("Expected different IDs")
	}
	if len(id1) != 16 {
		t.Errorf("Expected 16 char hex ID, got %d chars: %s", len(id1), id1)
	}
}

func TestSortFavourites(t *testing.T) {
	favourites := []Favourite{
		{Name: "Charlie"},
		{Name: "alpha"},
		{Name: "Bravo"},
		{Name: "delta"},
	}

	sortFavourites(favourites)

	expected := []string{"alpha", "Bravo", "Charlie", "delta"}
	for i, exp := range expected {
		if favourites[i].Name != exp {
			t.Errorf("Expected favourite %d to be %s, got %s", i, exp, favourites[i].Name)
		}
	}
}
