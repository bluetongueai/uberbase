package pkg

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

func ParseInt(s string) int {
	i, _ := strconv.Atoi(s)
	return i
}

func GenerateVersion() string {
	return time.Now().UTC().Format("20060102-150405")
}

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func StripTag(tag string) string {
	split := strings.Split(tag, ":")
	if len(split) > 1 {
		return split[0]
	}
	return tag
}

func ExpandTildeLocal(relPath string) (string, error) {
	if strings.HasPrefix(relPath, "~/") {
		relPath = strings.Replace(relPath, "~", os.Getenv("HOME"), 1)
	}
	absPath, err := filepath.Abs(relPath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}
