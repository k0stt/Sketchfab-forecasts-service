package main

import (
	"encoding/json"
	"flag"
	"os"
	"sketchfab-forecasts/internal/models"
	"sketchfab-forecasts/internal/preprocessing"

	"github.com/sirupsen/logrus"
)

var (
	inputFile      = flag.String("input", "data/raw_models.json", "Путь к файлу с сырыми данными")
	outputFile     = flag.String("output", "data/preprocessed_data.json", "Путь к выходному файлу")
	filterOutliers = flag.Bool("filter", true, "Фильтровать выбросы")
	threshold      = flag.Float64("threshold", 3.0, "Порог для фильтрации выбросов (в стандартных отклонениях)")
)

func main() {
	flag.Parse()

	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	logger.Info("Начинаем предобработку данных...")

	// Загрузка сырых данных
	logger.Infof("Загрузка данных из %s", *inputFile)
	rawModels, err := loadRawModels(*inputFile)
	if err != nil {
		logger.Fatalf("Ошибка при загрузке данных: %v", err)
	}

	logger.Infof("Загружено %d моделей", len(rawModels))

	// Создание препроцессора
	preprocessor := preprocessing.NewPreprocessor()

	// Обработка данных
	logger.Info("Обработка данных...")
	processedData := preprocessor.ProcessModels(rawModels)

	logger.Infof("Обработано %d записей", len(processedData))

	// Фильтрация выбросов
	if *filterOutliers {
		logger.Infof("Фильтрация выбросов (threshold=%.1f)...", *threshold)
		beforeCount := len(processedData)
		processedData = preprocessor.FilterOutliers(processedData, *threshold)
		logger.Infof("Удалено %d выбросов, осталось %d записей", beforeCount-len(processedData), len(processedData))
	}

	// Сохранение обработанных данных
	logger.Infof("Сохранение обработанных данных в %s", *outputFile)
	if err := saveProcessedData(processedData, *outputFile); err != nil {
		logger.Fatalf("Ошибка при сохранении данных: %v", err)
	}

	logger.Info("Предобработка завершена успешно!")
	printDataStats(processedData, logger)
}

func loadRawModels(filename string) ([]models.SketchfabModel, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var models []models.SketchfabModel
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&models); err != nil {
		return nil, err
	}

	return models, nil
}

func saveProcessedData(data []models.PreprocessedData, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(data)
}

func printDataStats(data []models.PreprocessedData, logger *logrus.Logger) {
	if len(data) == 0 {
		return
	}

	totalScore := 0.0
	minScore := data[0].PopularityScore
	maxScore := data[0].PopularityScore

	for _, item := range data {
		totalScore += item.PopularityScore
		if item.PopularityScore < minScore {
			minScore = item.PopularityScore
		}
		if item.PopularityScore > maxScore {
			maxScore = item.PopularityScore
		}
	}

	avgScore := totalScore / float64(len(data))

	logger.Info("=== Статистика обработанных данных ===")
	logger.Infof("Средний показатель популярности: %.4f", avgScore)
	logger.Infof("Минимальный: %.4f", minScore)
	logger.Infof("Максимальный: %.4f", maxScore)
}
