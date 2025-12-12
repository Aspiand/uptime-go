package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseDuration(t *testing.T) {
	testCases := []struct {
		name         string
		input        string
		defaultValue string
		expected     time.Duration
	}{
		{
			name:     "simple seconds",
			input:    "30s",
			expected: 30 * time.Second,
		},
		{
			name:     "simple minutes",
			input:    "5m",
			expected: 5 * time.Minute,
		},
		{
			name:     "simple hours",
			input:    "1h",
			expected: 1 * time.Hour,
		},
		{
			name:     "simple days",
			input:    "2d",
			expected: 2 * 24 * time.Hour,
		},
		{
			name:     "combined minutes and seconds",
			input:    "1m30s",
			expected: 90 * time.Second,
		},
		{
			name:     "combined hours, minutes, and seconds",
			input:    "1h5m10s",
			expected: 1*time.Hour + 5*time.Minute + 10*time.Second,
		},
		{
			name:         "empty input with default",
			input:        "",
			defaultValue: "10s",
			expected:     10 * time.Second,
		},
		{
			name:         "invalid input with default",
			input:        "invalid",
			defaultValue: "20m",
			expected:     20 * time.Minute,
		},
		{
			name:     "empty input without default",
			input:    "",
			expected: 0,
		},
		{
			name:     "invalid input without default",
			input:    "invalid",
			expected: 0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := parseDuration(tc.input, tc.defaultValue)
			assert.Equal(t, tc.expected, result)
		})
	}
}
