package preprocessing

import (
	"math"
	"sketchfab-forecasts/internal/models"
	"strings"
	"time"
)

// Preprocessor обработчик данных
type Preprocessor struct{}

// NewPreprocessor создает новый препроцессор
func NewPreprocessor() *Preprocessor {
	return &Preprocessor{}
}

// ProcessModel преобразует сырую модель в обработанные данные
func (p *Preprocessor) ProcessModel(model models.SketchfabModel) models.PreprocessedData {
	return models.PreprocessedData{
		ModelUID:           model.UID,
		CategoryCount:      len(model.Categories),
		TagCount:           len(model.Tags),
		DescriptionLength:  len(strings.TrimSpace(model.Description)),
		FaceCount:          model.FaceCount,
		VertexCount:        model.VertexCount,
		AnimationCount:     model.AnimationCount,
		IsDownloadable:     model.IsDownloadable,
		IsPremiumAuthor:    p.isPremiumUser(model.User),
		AuthorFollowers:    model.User.FollowerCount,
		DaysSincePublished: p.daysSince(model.PublishedAt.ToTime()),
		PopularityScore:    p.calculatePopularityScore(model),
	}
}

// ProcessModels обрабатывает массив моделей
func (p *Preprocessor) ProcessModels(rawModels []models.SketchfabModel) []models.PreprocessedData {
	processed := make([]models.PreprocessedData, 0, len(rawModels))

	for _, model := range rawModels {
		processed = append(processed, p.ProcessModel(model))
	}

	return processed
}

// calculatePopularityScore вычисляет общий показатель популярности
// Формула: взвешенная сумма лайков, просмотров и скачиваний
func (p *Preprocessor) calculatePopularityScore(model models.SketchfabModel) float64 {
	// Веса для различных метрик
	const (
		viewWeight     = 0.3
		likeWeight     = 0.4
		downloadWeight = 0.3
	)

	// Нормализация логарифмом для снижения влияния выбросов
	views := math.Log1p(float64(model.ViewCount))
	likes := math.Log1p(float64(model.LikeCount))
	downloads := math.Log1p(float64(model.DownloadCount))

	score := (views * viewWeight) + (likes * likeWeight) + (downloads * downloadWeight)

	return score
}

// isPremiumUser проверяет, является ли пользователь премиум
func (p *Preprocessor) isPremiumUser(user models.User) bool {
	return user.Account == "pro" || user.Account == "premium"
}

// daysSince вычисляет количество дней с даты публикации
func (p *Preprocessor) daysSince(publishedAt time.Time) float64 {
	if publishedAt.IsZero() {
		return 0
	}

	duration := time.Since(publishedAt)
	return duration.Hours() / 24
}

// NormalizeData нормализует численные признаки
func (p *Preprocessor) NormalizeData(data []models.PreprocessedData) []models.PreprocessedData {
	if len(data) == 0 {
		return data
	}

	// Находим min и max для каждого признака
	stats := p.calculateStats(data)

	// Нормализуем данные
	normalized := make([]models.PreprocessedData, len(data))
	for i, item := range data {
		normalized[i] = item
		normalized[i].FaceCount = p.normalize(float64(item.FaceCount), stats.minFaceCount, stats.maxFaceCount)
		normalized[i].VertexCount = p.normalize(float64(item.VertexCount), stats.minVertexCount, stats.maxVertexCount)
		normalized[i].AuthorFollowers = p.normalize(float64(item.AuthorFollowers), stats.minFollowers, stats.maxFollowers)
	}

	return normalized
}

type dataStats struct {
	minFaceCount, maxFaceCount     float64
	minVertexCount, maxVertexCount float64
	minFollowers, maxFollowers     float64
}

func (p *Preprocessor) calculateStats(data []models.PreprocessedData) dataStats {
	stats := dataStats{
		minFaceCount:   math.MaxFloat64,
		maxFaceCount:   0,
		minVertexCount: math.MaxFloat64,
		maxVertexCount: 0,
		minFollowers:   math.MaxFloat64,
		maxFollowers:   0,
	}

	for _, item := range data {
		face := float64(item.FaceCount)
		vertex := float64(item.VertexCount)
		followers := float64(item.AuthorFollowers)

		if face < stats.minFaceCount {
			stats.minFaceCount = face
		}
		if face > stats.maxFaceCount {
			stats.maxFaceCount = face
		}

		if vertex < stats.minVertexCount {
			stats.minVertexCount = vertex
		}
		if vertex > stats.maxVertexCount {
			stats.maxVertexCount = vertex
		}

		if followers < stats.minFollowers {
			stats.minFollowers = followers
		}
		if followers > stats.maxFollowers {
			stats.maxFollowers = followers
		}
	}

	return stats
}

func (p *Preprocessor) normalize(value, min, max float64) int {
	if max == min {
		return 0
	}
	normalized := (value - min) / (max - min)
	return int(normalized * 100) // Масштабируем 0-100
}

// FilterOutliers удаляет выбросы из данных
func (p *Preprocessor) FilterOutliers(data []models.PreprocessedData, threshold float64) []models.PreprocessedData {
	if len(data) == 0 {
		return data
	}

	// Вычисляем среднее и стандартное отклонение для PopularityScore
	mean, stdDev := p.calculateMeanStdDev(data)

	filtered := make([]models.PreprocessedData, 0, len(data))
	for _, item := range data {
		// Проверяем, находится ли значение в пределах threshold * stdDev от среднего
		if math.Abs(item.PopularityScore-mean) <= threshold*stdDev {
			filtered = append(filtered, item)
		}
	}

	return filtered
}

func (p *Preprocessor) calculateMeanStdDev(data []models.PreprocessedData) (mean, stdDev float64) {
	if len(data) == 0 {
		return 0, 0
	}

	// Среднее
	sum := 0.0
	for _, item := range data {
		sum += item.PopularityScore
	}
	mean = sum / float64(len(data))

	// Стандартное отклонение
	variance := 0.0
	for _, item := range data {
		diff := item.PopularityScore - mean
		variance += diff * diff
	}
	variance /= float64(len(data))
	stdDev = math.Sqrt(variance)

	return mean, stdDev
}
