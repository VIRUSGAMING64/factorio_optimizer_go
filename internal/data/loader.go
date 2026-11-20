package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadRootData loads the Factorio data from a JSON file (usually recipes.json).
// It accepts either a filesystem path or a name relative to the embedded data directory.
func LoadRootData(path string) (*RootData, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Try relative to common locations
		candidates := []string{
			filepath.Join("data", path),
			filepath.Join("..", "data", path),
			filepath.Join("internal", "..", "data", path),
		}
		var lastErr error = err
		for _, c := range candidates {
			data, err = os.ReadFile(c)
			if err == nil {
				break
			}
			lastErr = err
		}
		if data == nil {
			return nil, fmt.Errorf("load root data %q: %w", path, lastErr)
		}
	}

	var root RootData
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("parse root data: %w", err)
	}
	return &root, nil
}

// MustLoadRootData is like LoadRootData but panics on error (useful in tests / early startup).
func MustLoadRootData(path string) *RootData {
	rd, err := LoadRootData(path)
	if err != nil {
		panic(err)
	}
	return rd
}
