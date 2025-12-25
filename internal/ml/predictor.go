package ml

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sketchfab-forecasts/internal/models"

	"github.com/sirupsen/logrus"
)

// Predictor интерфейс для прогнозирования
type Predictor struct {
	logger *logrus.Logger
}

// NewPredictor создает новый предиктор
func NewPredictor(logger *logrus.Logger) *Predictor {
	return &Predictor{
		logger: logger,
	}
}

// Predict делает прогноз популярности модели
func (p *Predictor) Predict(req models.PredictionRequest) (*models.PredictionResponse, error) {
	// Подготовка данных для Python скрипта
	inputData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Вызов Python скрипта для предсказания
	cmd := exec.Command("python", "scripts/predict.py")
	cmd.Stdin = nil

	// Передаем данные через аргументы командной строки
	cmd.Args = append(cmd.Args, string(inputData))

	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Errorf("Python script error: %s", string(output))
		return nil, fmt.Errorf("prediction failed: %w", err)
	}

	// Парсим результат
	var response models.PredictionResponse
	if err := json.Unmarshal(output, &response); err != nil {
		return nil, fmt.Errorf("failed to parse prediction: %w", err)
	}

	return &response, nil
}

// categorizePopularity категоризирует показатель популярности
func categorizePopularity(score float64) string {
	if score < 2.0 {
		return "low"
	} else if score < 4.0 {
		return "medium"
	}
	return "high"
}

// MockPredict делает прогноз без вызова Python (для демонстрации)
func (p *Predictor) MockPredict(req models.PredictionRequest) *models.PredictionResponse {
	// Простая эвристика для демонстрации
	score := 0.0

	// Влияние различных факторов
	score += float64(req.TagCount) * 0.15
	score += float64(req.CategoryCount) * 0.2
	score += float64(req.DescriptionLength) * 0.001
	score += float64(req.AuthorFollowers) * 0.0001

	if req.IsDownloadable {
		score += 0.5
	}

	if req.IsPremiumAuthor {
		score += 0.3
	}

	if req.AnimationCount > 0 {
		score += 0.2
	}

	// Нормализация
	if score > 10 {
		score = 10
	}

	return &models.PredictionResponse{
		PopularityScore: score,
		Category:        categorizePopularity(score),
		Confidence:      0.75,
	}
}
