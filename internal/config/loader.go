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

// getEnvOrDefault returns the value of an environment variable or a default value if not set.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// requireEnv returns the value of a required environment variable or an error if not set.
func requireEnv(key string) (string, error) {
	const op = "config.requireEnv"
	value, ok := os.LookupEnv(key)
	if !ok || value == "" {
		return "", fmt.Errorf("%s: required environment variable %q is not set", op, key)
	}
	return value, nil
}
