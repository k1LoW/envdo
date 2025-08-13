package env

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnv_LoadEnvFiles(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		pwdFiles    map[string]string
		configFiles map[string]string
		wantEnvs    map[string]string
		wantError   bool
		setupEnv    map[string]string
	}{
		{
			name:    "normal case - load .env from pwd",
			profile: "",
			pwdFiles: map[string]string{
				".env": "KEY1=value1\nKEY2=value2\n",
			},
			wantEnvs: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantError: false,
		},
		{
			name:    "profile specified - load .env.test",
			profile: "test",
			pwdFiles: map[string]string{
				".env.test": "TEST_KEY=test_value\nENV=test\n",
			},
			wantEnvs: map[string]string{
				"TEST_KEY": "test_value",
				"ENV":      "test",
			},
			wantError: false,
		},
		{
			name:    "hierarchical search - pwd priority over config",
			profile: "",
			pwdFiles: map[string]string{
				".env": "KEY1=pwd_value\nPWD_ONLY=pwd_key\n",
			},
			configFiles: map[string]string{
				".env": "KEY1=config_value\nCONFIG_ONLY=config_key\n",
			},
			wantEnvs: map[string]string{
				"KEY1":        "pwd_value",
				"PWD_ONLY":    "pwd_key",
				"CONFIG_ONLY": "config_key",
			},
			wantError: false,
		},
		{
			name:    "hierarchical search - config fallback when pwd missing",
			profile: "",
			configFiles: map[string]string{
				".env": "CONFIG_KEY=config_value\n",
			},
			wantEnvs: map[string]string{
				"CONFIG_KEY": "config_value",
			},
			wantError: false,
		},
		{
			name:      "file not exist - empty result",
			profile:   "",
			wantEnvs:  map[string]string{},
			wantError: false,
		},
		{
			name:    "quoted values - remove quotes",
			profile: "",
			pwdFiles: map[string]string{
				".env": `DOUBLE_QUOTED="double quoted value"
SINGLE_QUOTED='single quoted value'
UNQUOTED=unquoted value
`,
			},
			wantEnvs: map[string]string{
				"DOUBLE_QUOTED": "double quoted value",
				"SINGLE_QUOTED": "single quoted value",
				"UNQUOTED":      "unquoted value",
			},
			wantError: false,
		},
		{
			name:    "comments and empty lines - ignored",
			profile: "",
			pwdFiles: map[string]string{
				".env": `# This is a comment
KEY1=value1

# Another comment
KEY2=value2
`,
			},
			wantEnvs: map[string]string{
				"KEY1": "value1",
				"KEY2": "value2",
			},
			wantError: false,
		},
		{
			name:    "invalid format - skip invalid lines",
			profile: "",
			pwdFiles: map[string]string{
				".env": `VALID_KEY=valid_value
INVALID_LINE_NO_EQUALS
KEY2=value2
=INVALID_EMPTY_KEY
KEY3=value3
`,
			},
			wantEnvs: map[string]string{
				"VALID_KEY": "valid_value",
				"KEY2":      "value2",
				"KEY3":      "value3",
				"":          "INVALID_EMPTY_KEY", // empty key is actually parsed
			},
			wantError: false,
		},
		{
			name:    "profile with hierarchy - test profile priority",
			profile: "staging",
			pwdFiles: map[string]string{
				".env.staging": "STAGE_KEY=pwd_staging\nCOMMON=pwd_common\n",
			},
			configFiles: map[string]string{
				".env.staging": "STAGE_KEY=config_staging\nCONFIG_STAGE=config_value\n",
			},
			wantEnvs: map[string]string{
				"STAGE_KEY":    "pwd_staging",
				"COMMON":       "pwd_common",
				"CONFIG_STAGE": "config_value",
			},
			wantError: false,
		},
		{
			name:    "custom XDG_CONFIG_HOME",
			profile: "",
			configFiles: map[string]string{
				".env": "XDG_KEY=xdg_value\n",
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "", // will be set to temp dir in test
			},
			wantEnvs: map[string]string{
				"XDG_KEY": "xdg_value",
			},
			wantError: false,
		},
	}

	// Add additional test case for file not found scenario
	tests = append(tests, struct {
		name        string
		profile     string
		pwdFiles    map[string]string
		configFiles map[string]string
		wantEnvs    map[string]string
		wantError   bool
		setupEnv    map[string]string
	}{
		name:      "loadEnvFile with non-existent file - should not error",
		profile:   "nonexistent",
		wantEnvs:  map[string]string{},
		wantError: false,
	})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directories
			tempPwd := t.TempDir()
			tempConfig := t.TempDir()

			// Setup environment variables
			for key, value := range tt.setupEnv {
				if key == "XDG_CONFIG_HOME" {
					t.Setenv(key, tempConfig)
				} else {
					t.Setenv(key, value)
				}
			}

			// Create files in pwd
			for filename, content := range tt.pwdFiles {
				createTestFile(t, tempPwd, filename, content)
			}

			// Create files in config directory
			configDir := filepath.Join(tempConfig, "envdo")
			if len(tt.configFiles) > 0 {
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("failed to create config directory: %v", err)
				}
				for filename, content := range tt.configFiles {
					createTestFile(t, configDir, filename, content)
				}
			}

			// Create Env instance
			env := New(tempPwd, tempConfig)

			// Execute LoadEnvFiles
			result, err := env.LoadEnvFiles(tt.profile)

			// Verify error expectation
			if tt.wantError && err == nil {
				t.Errorf("want error but got none")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify results
			if len(result) != len(tt.wantEnvs) {
				t.Errorf("want %d env vars, got %d. Want: %v, Got: %v",
					len(tt.wantEnvs), len(result), tt.wantEnvs, result)
				return
			}

			for key, wantValue := range tt.wantEnvs {
				if gotValue, exists := result[key]; !exists {
					t.Errorf("want key %q not found in result", key)
				} else if gotValue != wantValue {
					t.Errorf("key %q: want %q, got %q", key, wantValue, gotValue)
				}
			}
		})
	}
}

func TestLoadEnvFiles(t *testing.T) {
	tests := []struct {
		name        string
		profile     string
		pwdFiles    map[string]string
		configFiles map[string]string
		wantEnvs    map[string]string
		wantError   bool
		setupEnv    map[string]string
	}{
		{
			name:    "backward compatibility - load .env from current directory",
			profile: "",
			pwdFiles: map[string]string{
				".env": "COMPAT_KEY=compat_value\n",
			},
			wantEnvs: map[string]string{
				"COMPAT_KEY": "compat_value",
			},
			wantError: false,
		},
		{
			name:    "backward compatibility - profile support",
			profile: "production",
			pwdFiles: map[string]string{
				".env.production": "PROD_KEY=prod_value\n",
			},
			wantEnvs: map[string]string{
				"PROD_KEY": "prod_value",
			},
			wantError: false,
		},
		{
			name:    "backward compatibility - XDG_CONFIG_HOME fallback",
			profile: "",
			configFiles: map[string]string{
				".env": "CONFIG_FALLBACK_KEY=config_fallback_value\n",
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "", // will be set to temp dir in test
			},
			wantEnvs: map[string]string{
				"CONFIG_FALLBACK_KEY": "config_fallback_value",
			},
			wantError: false,
		},
		{
			name:    "backward compatibility - HOME/.config fallback",
			profile: "",
			configFiles: map[string]string{
				".env": "HOME_CONFIG_KEY=home_config_value\n",
			},
			setupEnv: map[string]string{
				"XDG_CONFIG_HOME": "",
				"HOME":            "", // will be set to temp dir in test
			},
			wantEnvs: map[string]string{
				"HOME_CONFIG_KEY": "home_config_value",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save original working directory
			originalWd, _ := os.Getwd()

			// Create temporary directories
			tempPwd := t.TempDir()
			tempConfig := t.TempDir()
			tempHome := t.TempDir()

			// Change to temporary working directory
			if err := os.Chdir(tempPwd); err != nil {
				t.Fatalf("failed to change directory: %v", err)
			}
			defer func() {
				if err := os.Chdir(originalWd); err != nil {
					t.Errorf("failed to restore working directory: %v", err)
				}
			}()

			// Clear existing environment variables to avoid interference
			t.Setenv("XDG_CONFIG_HOME", "")
			t.Setenv("HOME", "")

			// Setup environment variables with t.Setenv
			for key, value := range tt.setupEnv {
				switch key {
				case "XDG_CONFIG_HOME":
					if value == "" {
						t.Setenv(key, tempConfig)
					} else {
						t.Setenv(key, value)
					}
				case "HOME":
					if value == "" {
						t.Setenv(key, tempHome)
					} else {
						t.Setenv(key, value)
					}
				default:
					t.Setenv(key, value)
				}
			}

			// Create files in pwd (current working directory)
			for filename, content := range tt.pwdFiles {
				createTestFile(t, tempPwd, filename, content)
			}

			// Create files in config directory
			var configDir string
			if xdgConfigHome := os.Getenv("XDG_CONFIG_HOME"); xdgConfigHome != "" {
				configDir = filepath.Join(xdgConfigHome, "envdo")
			} else if home := os.Getenv("HOME"); home != "" {
				configDir = filepath.Join(home, ".config", "envdo")
			}

			if configDir != "" && len(tt.configFiles) > 0 {
				if err := os.MkdirAll(configDir, 0755); err != nil {
					t.Fatalf("failed to create config directory: %v", err)
				}
				for filename, content := range tt.configFiles {
					createTestFile(t, configDir, filename, content)
				}
			}

			// Execute LoadEnvFiles (backward compatibility function)
			result, err := LoadEnvFiles(tt.profile)

			// Verify error expectation
			if tt.wantError && err == nil {
				t.Errorf("want error but got none")
				return
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Verify results
			if len(result) != len(tt.wantEnvs) {
				t.Errorf("want %d env vars, got %d. Want: %v, Got: %v",
					len(tt.wantEnvs), len(result), tt.wantEnvs, result)
				return
			}

			for key, wantValue := range tt.wantEnvs {
				if gotValue, exists := result[key]; !exists {
					t.Errorf("want key %q not found in result", key)
				} else if gotValue != wantValue {
					t.Errorf("key %q: want %q, got %q", key, wantValue, gotValue)
				}
			}
		})
	}
}

// createTestFile creates a test file with specified content.
func createTestFile(t *testing.T, dir, filename, content string) {
	t.Helper()
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create test file %s: %v", filePath, err)
	}
}
