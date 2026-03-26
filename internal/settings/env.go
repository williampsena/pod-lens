package settings

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type PodLabels = map[string]string

func GetPort() int {
	rawPort := getEnv("PORT", "80")
	port, err := strconv.Atoi(rawPort)
	if err != nil {
		return 80
	}

	return port
}

func GetTheme() string {
	return getEnv("THEME", "light")
}

func GetPodLabels() PodLabels {
	// Try to read from file first (downwardAPI)
	labelsFile := os.Getenv("POD_LABELS_FILE")
	if labelsFile != "" {
		labels, err := parseLabelsFromFile(labelsFile)
		if err == nil {
			return labels
		}
	}

	// Fallback to environment variable
	rawLabels := os.Getenv("POD_LABELS")
	if rawLabels == "" {
		return make(map[string]string)
	}
	return parseLabelsFromEnv(rawLabels)
}

func parseLabelsFromEnv(raw string) map[string]string {
	labels := make(map[string]string)
	pairs := strings.Split(raw, ",")

	for _, pair := range pairs {
		kv := strings.SplitN(pair, "=", 2)
		if len(kv) == 2 {
			labels[kv[0]] = kv[1]
		}
	}
	return labels
}

func parseLabelsFromFile(filePath string) (map[string]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	labels := make(map[string]string)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // Skip empty lines
		}

		// Parse format: key="value"
		kv := strings.SplitN(line, "=", 2)
		if len(kv) == 2 {
			key := strings.TrimSpace(kv[0])
			value := strings.TrimSpace(kv[1])
			// Remove surrounding quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			labels[key] = value
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return labels, nil
}

func getEnv(env, defaultValue string) string {
	value := os.Getenv(env)

	if value == "" {
		return defaultValue
	}

	return value
}
