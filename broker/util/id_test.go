package util

import (
	"math/rand"
	"testing"
)

func TestGenId(t *testing.T) {
	src = rand.New(rand.NewSource(12345)) // Use a fixed seed for our source
	tests := []struct {
		name   string
		length int
		want   string
	}{
		{"GenId10", 10, "YJneTwAEKA"},
		{"GenId20", 20, "fxHzmLWvyoMxayksLWgS"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GenId(tt.length); got != tt.want {
				t.Errorf("GenId() = %v, want %v", got, tt.want)
			}
		})
	}
}
