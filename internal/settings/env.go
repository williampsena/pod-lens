package settings

import (
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
	rawLabels := os.Getenv("POD_LABELS")

	if rawLabels == "" {
		return make(map[string]string)
	}
	return parseLabels(rawLabels)
}

func parseLabels(raw string) map[string]string {
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

func getEnv(env, defaultValue string) string {
	value := os.Getenv(env)

	if value == "" {
		return defaultValue
	}

	return value
}
