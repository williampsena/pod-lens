package settings

import (
	"os"
	"testing"
)

func TestGetPort(t *testing.T) {
	tests := []struct {
		name     string
		portEnv  string
		expected int
	}{
		{
			name:     "default port 80",
			portEnv:  "",
			expected: 80,
		},
		{
			name:     "custom port 8080",
			portEnv:  "8080",
			expected: 8080,
		},
		{
			name:     "custom port 3000",
			portEnv:  "3000",
			expected: 3000,
		},
		{
			name:     "invalid port returns default",
			portEnv:  "invalid",
			expected: 80,
		},
		{
			name:     "port 0",
			portEnv:  "0",
			expected: 0,
		},
		{
			name:     "high port number",
			portEnv:  "65535",
			expected: 65535,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("PORT", tt.portEnv)
			defer os.Unsetenv("PORT")

			result := GetPort()
			if result != tt.expected {
				t.Errorf("GetPort() = %d, want %d", result, tt.expected)
			}
		})
	}
}

func TestGetTheme(t *testing.T) {
	tests := []struct {
		name     string
		themeEnv string
		expected string
	}{
		{
			name:     "theme set",
			themeEnv: "dark",
			expected: "dark",
		},
		{
			name:     "theme empty",
			themeEnv: "",
			expected: "light",
		},
		{
			name:     "light theme",
			themeEnv: "light",
			expected: "light",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("THEME", tt.themeEnv)
			defer os.Unsetenv("THEME")

			result := GetTheme()
			if result != tt.expected {
				t.Errorf("GetTheme() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestGetPodLabels(t *testing.T) {
	tests := []struct {
		name      string
		labelsEnv string
		expected  PodLabels
	}{
		{
			name:      "empty labels",
			labelsEnv: "",
			expected:  make(map[string]string),
		},
		{
			name:      "single label",
			labelsEnv: "app=myapp",
			expected:  map[string]string{"app": "myapp"},
		},
		{
			name:      "multiple labels",
			labelsEnv: "app=myapp,env=prod,version=1.0",
			expected: map[string]string{
				"app":     "myapp",
				"env":     "prod",
				"version": "1.0",
			},
		},
		{
			name:      "labels with spaces",
			labelsEnv: "description=my app,team=backend",
			expected: map[string]string{
				"description": "my app",
				"team":        "backend",
			},
		},
		{
			name:      "invalid format ignored",
			labelsEnv: "app=myapp,invalid,env=prod",
			expected: map[string]string{
				"app": "myapp",
				"env": "prod",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("POD_LABELS", tt.labelsEnv)
			defer os.Unsetenv("POD_LABELS")
			defer os.Unsetenv("POD_LABELS_FILE")

			result := GetPodLabels()
			if len(result) != len(tt.expected) {
				t.Errorf("GetPodLabels() returned %d labels, want %d", len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				value, exists := result[key]
				if !exists {
					t.Errorf("GetPodLabels() missing key %q", key)
				}
				if value != expectedValue {
					t.Errorf("GetPodLabels()[%q] = %q, want %q", key, value, expectedValue)
				}
			}
		})
	}
}

func TestParseLabelsFromEnv(t *testing.T) {
	tests := []struct {
		name     string
		raw      string
		expected map[string]string
	}{
		{
			name:     "empty string",
			raw:      "",
			expected: make(map[string]string),
		},
		{
			name:     "single pair",
			raw:      "key=value",
			expected: map[string]string{"key": "value"},
		},
		{
			name:     "multiple pairs",
			raw:      "app=myapp,env=staging",
			expected: map[string]string{"app": "myapp", "env": "staging"},
		},
		{
			name:     "value with equals sign",
			raw:      "connection=user=admin,password=secret",
			expected: map[string]string{"connection": "user=admin", "password": "secret"},
		},
		{
			name:     "invalid pairs without equals",
			raw:      "app=myapp,invalid",
			expected: map[string]string{"app": "myapp"},
		},
		{
			name:     "pairs with special characters",
			raw:      "url=http://example.com,path=/api/v1",
			expected: map[string]string{"url": "http://example.com", "path": "/api/v1"},
		},
		{
			name:     "single invalid pair",
			raw:      "invalid",
			expected: make(map[string]string),
		},
		{
			name:     "trailing comma",
			raw:      "app=myapp,",
			expected: map[string]string{"app": "myapp"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseLabelsFromEnv(tt.raw)

			if len(result) != len(tt.expected) {
				t.Errorf("parseLabelsFromEnv(%q) returned %d pairs, want %d", tt.raw, len(result), len(tt.expected))
			}

			for key, expectedValue := range tt.expected {
				value, exists := result[key]
				if !exists {
					t.Errorf("parseLabelsFromEnv(%q) missing key %q", tt.raw, key)
				}
				if value != expectedValue {
					t.Errorf("parseLabelsFromEnv(%q)[%q] = %q, want %q", tt.raw, key, value, expectedValue)
				}
			}
		})
	}
}

func TestParseLabelsFromFile(t *testing.T) {
	// Create a temporary test file
	tmpFile, err := os.CreateTemp("", "labels_*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	// Write test data in downwardAPI format: key="value"
	testData := `app.kubernetes.io/name="my-application"
app.kubernetes.io/instance="my-application-86f8d48a9-1.0.0"
ApplicationName="my-application"
Environment="stable"
helm.sh/chart="platform-webapp-1.0.1"
`
	if _, err := tmpFile.WriteString(testData); err != nil {
		t.Fatalf("Failed to write to temp file: %v", err)
	}
	tmpFile.Close()

	tests := []struct {
		name     string
		filePath string
		expected map[string]string
		wantErr  bool
	}{
		{
			name:     "valid file format",
			filePath: tmpFile.Name(),
			expected: map[string]string{
				"app.kubernetes.io/name":     "my-application",
				"app.kubernetes.io/instance": "my-application-86f8d48a9-1.0.0",
				"ApplicationName":            "my-application",
				"Environment":                "stable",
				"helm.sh/chart":              "platform-webapp-1.0.1",
			},
			wantErr: false,
		},
		{
			name:     "non-existent file",
			filePath: "/nonexistent/file.txt",
			expected: nil,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseLabelsFromFile(tt.filePath)

			if (err != nil) != tt.wantErr {
				t.Errorf("parseLabelsFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if len(result) != len(tt.expected) {
					t.Errorf("parseLabelsFromFile() returned %d pairs, want %d", len(result), len(tt.expected))
				}

				for key, expectedValue := range tt.expected {
					value, exists := result[key]
					if !exists {
						t.Errorf("parseLabelsFromFile() missing key %q", key)
					}
					if value != expectedValue {
						t.Errorf("parseLabelsFromFile()[%q] = %q, want %q", key, value, expectedValue)
					}
				}
			}
		})
	}
}

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name         string
		envVar       string
		envValue     string
		defaultValue string
		expected     string
	}{
		{
			name:         "env var exists",
			envVar:       "TEST_VAR",
			envValue:     "test_value",
			defaultValue: "default",
			expected:     "test_value",
		},
		{
			name:         "env var not set",
			envVar:       "NONEXISTENT_VAR",
			envValue:     "",
			defaultValue: "default_value",
			expected:     "default_value",
		},
		{
			name:         "env var empty string uses default",
			envVar:       "TEST_EMPTY",
			envValue:     "",
			defaultValue: "default",
			expected:     "default",
		},
		{
			name:         "env var with spaces",
			envVar:       "TEST_SPACES",
			envValue:     "  spaced value  ",
			defaultValue: "default",
			expected:     "  spaced value  ",
		},
		{
			name:         "empty default",
			envVar:       "TEST_EMPTY_DEFAULT",
			envValue:     "",
			defaultValue: "",
			expected:     "",
		},
		{
			name:         "env var with special chars",
			envVar:       "TEST_SPECIAL",
			envValue:     "!@#$%^&*()",
			defaultValue: "default",
			expected:     "!@#$%^&*()",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envValue != "" {
				os.Setenv(tt.envVar, tt.envValue)
				defer os.Unsetenv(tt.envVar)
			} else {
				os.Unsetenv(tt.envVar)
			}

			result := getEnv(tt.envVar, tt.defaultValue)
			if result != tt.expected {
				t.Errorf("getEnv(%q, %q) = %q, want %q", tt.envVar, tt.defaultValue, result, tt.expected)
			}
		})
	}
}
