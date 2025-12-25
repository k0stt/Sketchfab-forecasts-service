package api

import (
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewClient(t *testing.T) {
	logger := logrus.New()
	client := NewClient("https://api.sketchfab.com/v3", "test-token", logger)

	if client == nil {
		t.Fatal("Client should not be nil")
	}

	if client.baseURL != "https://api.sketchfab.com/v3" {
		t.Errorf("Expected baseURL to be 'https://api.sketchfab.com/v3', got '%s'", client.baseURL)
	}

	if client.apiToken != "test-token" {
		t.Errorf("Expected apiToken to be 'test-token', got '%s'", client.apiToken)
	}
}

func TestBuildURL(t *testing.T) {
	logger := logrus.New()
	client := NewClient("https://api.sketchfab.com/v3", "test-token", logger)

	tests := []struct {
		name     string
		endpoint string
		params   map[string]string
		expected string
	}{
		{
			name:     "Simple endpoint",
			endpoint: "/models",
			params:   nil,
			expected: "https://api.sketchfab.com/v3/models",
		},
		{
			name:     "Endpoint with params",
			endpoint: "/models",
			params:   map[string]string{"sort_by": "likes"},
			expected: "https://api.sketchfab.com/v3/models?sort_by=likes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := client.buildURL(tt.endpoint, tt.params)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestBuildSearchParams(t *testing.T) {
	params := BuildSearchParams(SearchParams{
		Query: "car",
		Sort:  "likes",
	})

	if params["q"] != "car" {
		t.Errorf("Expected query 'car', got '%s'", params["q"])
	}

	if params["sort_by"] != "likes" {
		t.Errorf("Expected sort_by 'likes', got '%s'", params["sort_by"])
	}
}
