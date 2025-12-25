package main

import (
	"encoding/json"
	"flag"
	"os"
	"sketchfab-forecasts/internal/api"
	"sketchfab-forecasts/internal/models"

	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

var (
	limit      = flag.Int("limit", 500, "Количество моделей для сбора")
	outputFile = flag.String("output", "data/raw_models.json", "Путь к выходному файлу")
	sort       = flag.String("sort", "likes", "Сортировка: likes, views, recent")
)

func main() {
	flag.Parse()

	// Настройка логирования
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
	logger.SetLevel(logrus.InfoLevel)

	// Загрузка переменных окружения
	if err := godotenv.Load(); err != nil {
		logger.Warn("Файл .env не найден, используем системные переменные")
	}

	apiToken := os.Getenv("SKETCHFAB_API_TOKEN")
	if apiToken == "" {
		logger.Fatal("SKETCHFAB_API_TOKEN не установлен")
	}

	apiURL := os.Getenv("SKETCHFAB_API_URL")
	if apiURL == "" {
		apiURL = "https://api.sketchfab.com/v3"
	}

	logger.Info("Начинаем сбор данных с Sketchfab API...")
	logger.Infof("Лимит моделей: %d", *limit)
	logger.Infof("Сортировка: %s", *sort)

	// Создание API клиента
	client := api.NewClient(apiURL, apiToken, logger)

	// Параметры поиска
	searchParams := api.SearchParams{
		Sort:         *sort,
		Downloadable: false,
	}
	params := api.BuildSearchParams(searchParams)

	// Сбор моделей
	modelsData, err := client.FetchModels(params, *limit)
	if err != nil {
		logger.Fatalf("Ошибка при сборе моделей: %v", err)
	}

	logger.Infof("Собрано %d моделей", len(modelsData))

	// Сохранение в файл
	if err := saveModels(modelsData, *outputFile, logger); err != nil {
		logger.Fatalf("Ошибка при сохранении данных: %v", err)
	}

	logger.Infof("Данные успешно сохранены в %s", *outputFile)
	printStats(modelsData, logger)
}

func saveModels(modelsData []models.SketchfabModel, filename string, logger *logrus.Logger) error {
	// Создаем директорию если не существует
	if err := os.MkdirAll("data", 0755); err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")

	return encoder.Encode(modelsData)
}

func printStats(modelsData []models.SketchfabModel, logger *logrus.Logger) {
	if len(modelsData) == 0 {
		return
	}

	totalViews := 0
	totalLikes := 0
	totalDownloads := 0
	categoriesMap := make(map[string]int)
	tagsMap := make(map[string]int)

	for _, model := range modelsData {
		totalViews += model.ViewCount
		totalLikes += model.LikeCount
		totalDownloads += model.DownloadCount

		for _, cat := range model.Categories {
			categoriesMap[cat]++
		}

		for _, tag := range model.Tags {
			tagsMap[tag]++
		}
	}

	logger.Info("=== Статистика собранных данных ===")
	logger.Infof("Всего моделей: %d", len(modelsData))
	logger.Infof("Средние просмотры: %d", totalViews/len(modelsData))
	logger.Infof("Средние лайки: %d", totalLikes/len(modelsData))
	logger.Infof("Средние скачивания: %d", totalDownloads/len(modelsData))
	logger.Infof("Уникальных категорий: %d", len(categoriesMap))
	logger.Infof("Уникальных тегов: %d", len(tagsMap))
}
