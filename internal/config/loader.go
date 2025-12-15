package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// loadFromYaml loads configuration from a YAML file into the provided target struct.
// Returns nil error if file doesn't exist (allows fallback to env vars).
func loadFromYaml(filePath string, target any) error {
	const op = "config.loadFromYaml"

	// If file doesn't exist, return nil to allow env var fallback
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("%s: failed to read file: %w", op, err)
	}

	if err := yaml.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%s: failed to unmarshal YAML: %w", op, err)
	}

	return nil
}
