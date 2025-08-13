package env

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Env represents an environment loader with configurable directories.
type Env struct {
	pwd       string
	configDir string
}

// New creates a new Env instance with specified directories.
func New(pwd, configDir string) *Env {
	return &Env{
		pwd:       pwd,
		configDir: configDir,
	}
}

// LoadEnvFiles loads .env files from multiple directories with priority.
// Priority: pwd > configDir/envdo.
func (e *Env) LoadEnvFiles(profile string) (map[string]string, error) {
	envs := make(map[string]string)

	// Determine .env filename
	filename := ".env"
	if profile != "" {
		filename = fmt.Sprintf(".env.%s", profile)
	}

	// Get directories to search
	dirs := e.getSearchDirectories()

	// Load from directories in reverse order (lower priority first)
	for i := len(dirs) - 1; i >= 0; i-- {
		envPath := filepath.Join(dirs[i], filename)
		if _, err := os.Stat(envPath); err == nil {
			if err := loadEnvFile(envPath, envs); err != nil {
				return nil, fmt.Errorf("failed to load %s: %w", envPath, err)
			}
		}
	}

	return envs, nil
}

// getSearchDirectories returns directories to search for .env files.
// Returns in priority order: [pwd, configDir/envdo].
func (e *Env) getSearchDirectories() []string {
	dirs := []string{}

	// Current directory (highest priority)
	if e.pwd != "" {
		dirs = append(dirs, e.pwd)
	}

	// Config directory/envdo
	if e.configDir != "" {
		envdoConfigDir := filepath.Join(e.configDir, "envdo")
		dirs = append(dirs, envdoConfigDir)
	}

	return dirs
}

// LoadEnvFiles loads .env files from multiple directories with priority.
// Priority: current directory > XDG_CONFIG_HOME/envdo.
// This function maintains backward compatibility by using default directories.
func LoadEnvFiles(profile string) (map[string]string, error) {
	// Get current working directory
	pwd, err := os.Getwd()
	if err != nil {
		pwd = ""
	}

	// Get config directory
	configDir := os.Getenv("XDG_CONFIG_HOME")
	if configDir == "" {
		if homeDir, err := os.UserHomeDir(); err == nil {
			configDir = filepath.Join(homeDir, ".config")
		}
	}

	// Create Env instance with default directories
	env := New(pwd, configDir)
	return env.LoadEnvFiles(profile)
}

// loadEnvFile loads environment variables from a .env file.
func loadEnvFile(filename string, envs map[string]string) error {
	file, err := os.Open(filename)
	if err != nil {
		// If file doesn't exist, silently skip without error
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Parse key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// Remove quotes if present
		if len(value) >= 2 {
			if (value[0] == '"' && value[len(value)-1] == '"') ||
				(value[0] == '\'' && value[len(value)-1] == '\'') {
				value = value[1 : len(value)-1]
			}
		}

		envs[key] = value
	}

	return scanner.Err()
}
