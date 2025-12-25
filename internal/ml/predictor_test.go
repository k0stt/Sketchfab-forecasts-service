package ml

import (
	"sketchfab-forecasts/internal/models"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestNewPredictor(t *testing.T) {
	logger := logrus.New()
	predictor := NewPredictor(logger)

	if predictor == nil {
		t.Fatal("Predictor should not be nil")
	}
}

func TestMockPredict(t *testing.T) {
	logger := logrus.New()
	predictor := NewPredictor(logger)

	tests := []struct {
		name     string
		request  models.PredictionRequest
		minScore float64
		maxScore float64
	}{
		{
			name: "Basic model",
			request: models.PredictionRequest{
				CategoryCount:     1,
				TagCount:          5,
				DescriptionLength: 100,
				FaceCount:         1000,
				VertexCount:       500,
				AnimationCount:    0,
				IsDownloadable:    false,
				IsPremiumAuthor:   false,
				AuthorFollowers:   10,
			},
			minScore: 0,
			maxScore: 3,
		},
		{
			name: "Premium model",
			request: models.PredictionRequest{
				CategoryCount:     3,
				TagCount:          10,
				DescriptionLength: 500,
				FaceCount:         10000,
				VertexCount:       5000,
				AnimationCount:    2,
				IsDownloadable:    true,
				IsPremiumAuthor:   true,
				AuthorFollowers:   1000,
			},
			minScore: 2,
			maxScore: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := predictor.MockPredict(tt.request)

			if result == nil {
				t.Fatal("Result should not be nil")
			}

			if result.PopularityScore < tt.minScore || result.PopularityScore > tt.maxScore {
				t.Errorf("Score %.2f not in expected range [%.2f, %.2f]",
					result.PopularityScore, tt.minScore, tt.maxScore)
			}

			if result.Category == "" {
				t.Error("Category should not be empty")
			}

			if result.Confidence <= 0 || result.Confidence > 1 {
				t.Errorf("Confidence %.2f should be between 0 and 1", result.Confidence)
			}
		})
	}
}

func TestCategorizePopularity(t *testing.T) {
	tests := []struct {
		score    float64
		expected string
	}{
		{1.0, "low"},
		{1.9, "low"},
		{2.0, "medium"},
		{3.5, "medium"},
		{4.0, "high"},
		{10.0, "high"},
	}

	for _, tt := range tests {
		result := categorizePopularity(tt.score)
		if result != tt.expected {
			t.Errorf("Score %.1f: expected '%s', got '%s'", tt.score, tt.expected, result)
		}
	}
}
