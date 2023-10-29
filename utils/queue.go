package utils

import (
	"log"
	"strings"
)

func GetSuffix(routingKey string) string {
	parts := strings.Split(routingKey, ".")
	log.Println("Suffix is: ", parts[len(parts)-1])
	return parts[len(parts)-1]
}
