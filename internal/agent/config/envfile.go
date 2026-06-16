package config

import (
	"bufio"
	"os"
	"strings"
)

// loadEnvFile sets variables from path when not already defined in the environment.
func loadEnvFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		value = strings.TrimSpace(value)
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		if os.Getenv(key) == "" {
			_ = os.Setenv(key, value)
		}
	}
	return scanner.Err()
}

// loadDefaultEnvFiles loads deploy/agent.env from the working directory.
func loadDefaultEnvFiles() {
	for _, path := range []string{"deploy/agent.env", "agent.env"} {
		_ = loadEnvFile(path)
	}
}
