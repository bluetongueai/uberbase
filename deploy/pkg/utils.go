package pkg

import (
	"strconv"
	"strings"
)

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func extractPort(port string) string {
	parts := strings.Split(port, ":")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}
