package ml

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"sketchfab-forecasts/internal/models"
	"strings"

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

	// Используем расширенный скрипт predict_advanced.py
	scriptPath := "scripts/predict_advanced.py"

	// Создаем новую команду каждый раз
	cmd := exec.Command("python", scriptPath)

	// Передаем данные через stdin
	cmd.Stdin = strings.NewReader(string(inputData))

	// Получаем результат
	output, err := cmd.CombinedOutput()
	if err != nil {
		p.logger.Warnf("Advanced script failed: %v, output: %s", err, string(output))
		// Fallback на стандартный скрипт
		cmd = exec.Command("python", "scripts/predict.py", string(inputData))
		output, err = cmd.CombinedOutput()
		if err != nil {
			p.logger.Errorf("Standard script also failed: %v", err)
			return nil, fmt.Errorf("prediction failed: %w", err)
		}
	}

	// Парсим результат
	var response models.PredictionResponse
	if err := json.Unmarshal(output, &response); err != nil {
		p.logger.Errorf("Failed to parse output: %s", string(output))
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
	score += float64(req.AuthorFollowers) * 0.0005

	// ВАЖНО: Учитываем полигоны!
	// Оптимальный диапазон: 5000-50000
	if req.FaceCount >= 5000 && req.FaceCount <= 50000 {
		score += 1.0 // Бонус за оптимальное количество
	} else if req.FaceCount > 0 {
		// Чем дальше от оптимума, тем меньше бонус
		deviation := 0.0
		if req.FaceCount < 5000 {
			deviation = (5000.0 - float64(req.FaceCount)) / 5000.0
		} else {
			deviation = (float64(req.FaceCount) - 50000.0) / 100000.0
		}
		score += 1.0 - (deviation * 0.5)
	}

	// Учитываем вершины
	if req.VertexCount >= 2500 && req.VertexCount <= 25000 {
		score += 0.5
	} else if req.VertexCount > 0 {
		score += 0.2
	}

	if req.IsDownloadable {
		score += 0.5
	}

	if req.IsPremiumAuthor {
		score += 0.3
	}

	if req.AnimationCount > 0 {
		score += float64(req.AnimationCount) * 0.2
	}

	// Нормализация (шкала 0-10)
	if score > 10 {
		score = 10
	}
	if score < 0 {
		score = 0
	}

	return &models.PredictionResponse{
		PopularityScore: score,
		Category:        categorizePopularity(score),
		Confidence:      0.75,
	}
}
