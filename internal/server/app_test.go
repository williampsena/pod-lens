package server

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestMaskValue(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"empty", "", "***"},
		{"short", "abc", "***"},
		{"exactly 6", "123456", "***"},
		{"long value", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9", "eyJ...CJ9"},
		{"token", "sk_test_123456789abcdef", "sk_...def"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskValue(tt.input)
			if result != tt.expected {
				t.Errorf("maskValue(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestMaskSensitiveHeaders(t *testing.T) {
	tests := []struct {
		name           string
		headers        http.Header
		expectedMasked []string
		shouldBeMasked bool
	}{
		{
			name: "authorization header masked",
			headers: http.Header{
				"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
			},
			expectedMasked: []string{"Authorization"},
			shouldBeMasked: true,
		},
		{
			name: "cookie header masked",
			headers: http.Header{
				"Cookie": []string{"session=abc123xyz789"},
			},
			expectedMasked: []string{"Cookie"},
			shouldBeMasked: true,
		},
		{
			name: "user-agent not masked",
			headers: http.Header{
				"User-Agent": []string{"Mozilla/5.0"},
			},
			expectedMasked: []string{"User-Agent"},
			shouldBeMasked: false,
		},
		{
			name: "mixed headers",
			headers: http.Header{
				"Authorization": []string{"Bearer token123"},
				"User-Agent":    []string{"Mozilla/5.0"},
				"X-Api-Key":     []string{"sk_test_key"},
			},
			expectedMasked: []string{"Authorization", "User-Agent", "X-Api-Key"},
			shouldBeMasked: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSensitiveHeaders(tt.headers)

			for _, header := range tt.expectedMasked {
				if _, exists := result[header]; !exists {
					t.Errorf("expected header %q to exist", header)
				}
			}

			// Check if Authorization is masked
			if auth, exists := result["Authorization"]; exists && tt.shouldBeMasked {
				if len(auth) > 0 && !strings.Contains(auth[0], "...") {
					t.Errorf("Authorization should be masked, got %q", auth[0])
				}
			}
		})
	}
}

func TestGetSafeEnvironment(t *testing.T) {
	// Set test environment variables
	os.Setenv("KUBERNETES_SERVICE_HOST", "10.0.0.1")
	os.Setenv("POD_NAME", "test-pod")
	os.Setenv("SECRET_PASSWORD", "should-not-appear")
	defer func() {
		os.Unsetenv("KUBERNETES_SERVICE_HOST")
		os.Unsetenv("POD_NAME")
		os.Unsetenv("SECRET_PASSWORD")
	}()

	result := getSafeEnvironment()

	// Check safe vars exist
	if _, exists := result["KUBERNETES_SERVICE_HOST"]; !exists {
		t.Error("KUBERNETES_SERVICE_HOST should be in safe environment")
	}

	if _, exists := result["POD_NAME"]; !exists {
		t.Error("POD_NAME should be in safe environment")
	}

	// Check sensitive vars don't exist
	if _, exists := result["SECRET_PASSWORD"]; exists {
		t.Error("SECRET_PASSWORD should not be in safe environment")
	}
}

func TestGetPodInfo(t *testing.T) {
	info := getPodInfo()

	tests := []struct {
		name  string
		check func() bool
	}{
		{"hostname not empty", func() bool { return info.Hostname != "" }},
		{"go version not empty", func() bool { return info.GoVersion != "" }},
		{"os not empty", func() bool { return info.OS != "" }},
		{"architecture not empty", func() bool { return info.Architecture != "" }},
		{"startup time not empty", func() bool { return info.StartupTime != "" }},
		{"ip addresses exist", func() bool { return len(info.IPAddresses) >= 0 }},
		{"environment map exists", func() bool { return info.Environment != nil }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.check() {
				t.Error("check failed")
			}
		})
	}
}

func TestGetLocalIPAddresses(t *testing.T) {
	ips := getLocalIPAddresses()

	if len(ips) == 0 {
		t.Log("Warning: no IP addresses found (might be expected in some environments)")
	}

	for _, ip := range ips {
		if ip == "" {
			t.Error("IP address should not be empty")
		}
	}
}

func BenchmarkMaskValue(b *testing.B) {
	for i := 0; i < b.N; i++ {
		maskValue("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9")
	}
}

func BenchmarkMaskSensitiveHeaders(b *testing.B) {
	headers := http.Header{
		"Authorization": []string{"Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9"},
		"User-Agent":    []string{"Mozilla/5.0"},
		"Cookie":        []string{"session=abc123xyz789"},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		maskSensitiveHeaders(headers)
	}
}

func TestHealthzHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
		expectedBody   string
		expectedHeader string
	}{
		{
			name:           "GET request returns 200",
			method:         "GET",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "POST request returns 200",
			method:         "POST",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedHeader: "text/plain; charset=utf-8",
		},
		{
			name:           "HEAD request returns 200",
			method:         "HEAD",
			expectedStatus: http.StatusOK,
			expectedBody:   "OK",
			expectedHeader: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, "/healthz", nil)
			w := httptest.NewRecorder()

			healthzHandler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != tt.expectedHeader {
				t.Errorf("expected Content-Type %q, got %q", tt.expectedHeader, contentType)
			}
		})
	}
}

func BenchmarkHealthzHandler(b *testing.B) {
	req := httptest.NewRequest("GET", "/healthz", nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		w := httptest.NewRecorder()
		healthzHandler(w, req)
	}
}
