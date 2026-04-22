package config

import (
	"errors"
	"fmt"
	"os"
	"strings"
)

func MustGet(key string) string {
	value := os.Getenv(key)
	if value == "" {
		panic(fmt.Sprintf("config.MustGet: %s is required variable", key))
	}
	return value
}

func Get(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func LoadDotEnv(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("config.LoadDotEnv: %w", err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, "=")
		if !found {
			continue
		}
		key = strings.TrimSpace(key)
		if os.Getenv(key) == "" {
			if err := os.Setenv(key, value); err != nil {
				return fmt.Errorf("config.LoadDotEnv: %w", err)
			}
		}
	}
	return nil
}
