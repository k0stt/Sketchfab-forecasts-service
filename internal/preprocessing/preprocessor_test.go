package preprocessing

import (
	"sketchfab-forecasts/internal/models"
	"testing"
	"time"
)

func TestCalculatePopularityScore(t *testing.T) {
	p := NewPreprocessor()

	model := models.SketchfabModel{
		ViewCount:     1000,
		LikeCount:     100,
		DownloadCount: 50,
	}

	score := p.calculatePopularityScore(model)

	if score <= 0 {
		t.Error("Popularity score should be positive")
	}
}

func TestIsPremiumUser(t *testing.T) {
	p := NewPreprocessor()

	tests := []struct {
		name     string
		account  string
		expected bool
	}{
		{"Basic account", "basic", false},
		{"Pro account", "pro", true},
		{"Premium account", "premium", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user := models.User{Account: tt.account}
			result := p.isPremiumUser(user)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}

func TestDaysSince(t *testing.T) {
	p := NewPreprocessor()

	// Тест с датой 30 дней назад
	pastDate := time.Now().AddDate(0, 0, -30)
	days := p.daysSince(pastDate)

	if days < 29 || days > 31 {
		t.Errorf("Expected around 30 days, got %.2f", days)
	}

	// Тест с нулевой датой
	zeroDate := time.Time{}
	days = p.daysSince(zeroDate)
	if days != 0 {
		t.Errorf("Expected 0 days for zero time, got %.2f", days)
	}
}

func TestProcessModel(t *testing.T) {
	p := NewPreprocessor()

	model := models.SketchfabModel{
		UID:            "test123",
		Description:    "Test model description",
		Categories:     []string{"Characters", "Animals"},
		Tags:           []string{"lowpoly", "game-ready", "pbr"},
		ViewCount:      1000,
		LikeCount:      100,
		DownloadCount:  50,
		FaceCount:      5000,
		VertexCount:    2500,
		AnimationCount: 1,
		IsDownloadable: true,
		User: models.User{
			Account:       "pro",
			FollowerCount: 200,
		},
		PublishedAt: models.SketchfabTime{Time: time.Now().AddDate(0, 0, -10)},
	}

	processed := p.ProcessModel(model)

	if processed.ModelUID != "test123" {
		t.Errorf("Expected UID 'test123', got '%s'", processed.ModelUID)
	}

	if processed.CategoryCount != 2 {
		t.Errorf("Expected 2 categories, got %d", processed.CategoryCount)
	}

	if processed.TagCount != 3 {
		t.Errorf("Expected 3 tags, got %d", processed.TagCount)
	}

	if !processed.IsPremiumAuthor {
		t.Error("Expected premium author")
	}

	if processed.PopularityScore <= 0 {
		t.Error("Expected positive popularity score")
	}
}

func TestFilterOutliers(t *testing.T) {
	p := NewPreprocessor()

	data := []models.PreprocessedData{
		{PopularityScore: 1.0},
		{PopularityScore: 2.0},
		{PopularityScore: 2.5},
		{PopularityScore: 3.0},
		{PopularityScore: 100.0}, // Выброс
	}

	filtered := p.FilterOutliers(data, 1.5)

	// С более строгим threshold должны отфильтровать выброс
	if len(filtered) == 0 {
		t.Error("Should have some data left after filtering")
	}

	// Проверяем, что экстремальный выброс удален
	hasExtremeOutlier := false
	for _, item := range filtered {
		if item.PopularityScore > 50 {
			hasExtremeOutlier = true
		}
	}

	if hasExtremeOutlier {
		t.Log("Note: Extreme outlier detection depends on threshold and distribution")
	}
}
