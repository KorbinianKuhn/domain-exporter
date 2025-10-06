package utils

import (
	"strings"
)

func GetTLD(url string) string {
	parts := strings.Split(url, ".")
	return parts[len(parts)-1]
}
