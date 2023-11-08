package utils

import (
	"testing"
)

func TestGetSuffix(t *testing.T) {
	tests := []struct {
		name       string
		routingKey string
		want       string
	}{
		{"Simple case", "test.imageCreate", "imageCreate"},
		{"Multiple dots", "test.multiple.dots.imageCreated", "imageCreated"},
		{"No dots", "nodots", "nodots"},
		{"Empty string", "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if suffix := GetSuffix(tt.routingKey); suffix != tt.want {
				t.Errorf("GetSuffix() = %v, want %v", suffix, tt.want)
			}
		})
	}
}
