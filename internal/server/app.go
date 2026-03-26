package server

import (
	"bufio"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/williampsena/pod-lens/internal/settings"
)

type PodInfo struct {
	Hostname     string
	IPAddresses  []string
	GoVersion    string
	OS           string
	Architecture string
	StartupTime  string
	Environment  map[string]string
}

type PageData struct {
	Labels  map[string]string
	Headers map[string][]string
	Theme   string
	Pod     PodInfo
}

var cachedTemplate *template.Template

// Sensitive headers that should be masked
var sensitiveHeaders = map[string]bool{
	"Authorization":       true,
	"Cookie":              true,
	"Set-Cookie":          true,
	"X-Api-Key":           true,
	"X-Auth-Token":        true,
	"X-Access-Token":      true,
	"X-Refresh-Token":     true,
	"X-Csrf-Token":        true,
	"Proxy-Authorization": true,
	"Www-Authenticate":    true,
	"Cf-Authorization":    true,
}

// Safe environment variables to show (non-sensitive)
var safeEnvPrefixes = []string{
	"KUBERNETES_",
	"POD_",
	"NODE_",
	"SERVICE_",
	"NAMESPACE",
	"APP_",
	"THEME",
	"PORT",
	"LANG",
	"TERM",
	"PATH",
	"HOME",
	"USER",
	"GOPATH",
	"GOROOT",
}

var podStartTime = time.Now()

// maskValue masks sensitive values, showing only first and last 3 characters
func maskValue(value string) string {
	if len(value) <= 6 {
		return "***"
	}
	return value[:3] + "..." + value[len(value)-3:]
}

// maskSensitiveHeaders processes headers and masks sensitive values
func maskSensitiveHeaders(headers http.Header) map[string][]string {
	result := make(map[string][]string)

	for key, values := range headers {
		if sensitiveHeaders[key] || sensitiveHeaders[strings.ToLower(key)] {
			// Mask all values for sensitive headers
			maskedValues := make([]string, len(values))
			for i, value := range values {
				maskedValues[i] = maskValue(value)
			}
			result[key] = maskedValues
		} else {
			result[key] = values
		}
	}

	return result
}

// getLocalIPAddresses returns all local IP addresses
func getLocalIPAddresses() []string {
	var ips []string
	interfaces, err := net.Interfaces()
	if err != nil {
		return ips
	}

	for _, iface := range interfaces {
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				ips = append(ips, ipnet.IP.String())
			}
		}
	}

	return ips
}

// getSafeEnvironment returns environment variables that are safe to display
func getSafeEnvironment() map[string]string {
	safeEnv := make(map[string]string)
	envVars := os.Environ()

	for _, envVar := range envVars {
		key := strings.Split(envVar, "=")[0]

		// Check if it's in safe prefixes
		isSafe := false
		for _, prefix := range safeEnvPrefixes {
			if strings.HasPrefix(key, prefix) {
				isSafe = true
				break
			}
		}

		if isSafe && len(envVar) > 0 {
			parts := strings.SplitN(envVar, "=", 2)
			if len(parts) == 2 {
				safeEnv[parts[0]] = parts[1]
			}
		}
	}

	return safeEnv
}

// getPodInfo collects pod information similar to traefik/whoami
func getPodInfo() PodInfo {
	hostname, _ := os.Hostname()
	ips := getLocalIPAddresses()

	return PodInfo{
		Hostname:     hostname,
		IPAddresses:  ips,
		GoVersion:    runtime.Version(),
		OS:           runtime.GOOS,
		Architecture: runtime.GOARCH,
		StartupTime:  podStartTime.Format("2006-01-02T15:04:05Z07:00"),
		Environment:  getSafeEnvironment(),
	}
}

func handler(w http.ResponseWriter, r *http.Request) {
	theme := settings.GetTheme()

	data := PageData{
		Labels:  settings.GetPodLabels(),
		Pod:     getPodInfo(),
		Headers: maskSensitiveHeaders(r.Header),
		Theme:   theme,
	}

	if cachedTemplate == nil {
		cachedTemplate = template.Must(template.ParseFiles("pages/index.html"))
	}
	cachedTemplate.Execute(w, data)
}

func staticHandler(w http.ResponseWriter, r *http.Request) {
	fileName := strings.TrimPrefix(r.URL.Path, "/static/")
	filePath := filepath.Join("static/", filepath.Clean(fileName))

	fileInfo, err := os.Stat(filePath)
	if err != nil || fileInfo.IsDir() {
		fmt.Printf("❌ Static file not found: %s\n", filePath)
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	fmt.Printf("📄 Serving static file: %s\n", filePath)
	http.ServeFile(w, r, filePath)
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func RunAndServer() error {
	port := fmt.Sprintf(":%v", settings.GetPort())

	mux := http.NewServeMux()
	mux.HandleFunc("/", handler)
	mux.HandleFunc("/static/", staticHandler)
	mux.HandleFunc("/healthz", healthzHandler)

	server := &http.Server{
		Addr:    port,
		Handler: mux,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	keyChan := make(chan string, 1)

	go func() {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println("💡 Press 'c' + Enter to shutdown, or Ctrl+C")
		for {
			input, err := reader.ReadString('\n')
			if err != nil {
				// In Kubernetes containers, stdin may be closed or unavailable
				// Exit gracefully instead of tight looping
				return
			}
			input = strings.TrimSpace(strings.ToLower(input))
			if input == "c" {
				keyChan <- "c"
				return
			}
		}
	}()

	go func() {
		fmt.Printf("🚀 Starting pod-lens server on http://localhost%s\n", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("❌ Server error: %v", err)
		}
	}()

	// Wait for either signal or 'c' key press
	select {
	case sig := <-sigChan:
		fmt.Printf("\n📢 Received signal: %s\n", sig)
	case <-keyChan:
		fmt.Println("\n🔤 Received 'c' command")
	}

	fmt.Println("🛑 Shutting down server...")

	if err := server.Close(); err != nil {
		return fmt.Errorf("❌ Error closing server: %v", err)
	}

	fmt.Println("✅ Server shutdown complete")

	return nil
}
