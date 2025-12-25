# Makefile для Sketchfab Forecasts проекта

.PHONY: help install scrape preprocess eda train server run-all clean test docker-build docker-up docker-down docker-logs docker-pipeline

help:
	@echo "Доступные команды:"
	@echo ""
	@echo "🐳 Docker (рекомендуется):"
	@echo "  make docker-up        - Запустить веб-сервер в Docker"
	@echo "  make docker-pipeline  - Полный пайплайн в Docker"
	@echo "  make docker-down      - Остановить все контейнеры"
	@echo "  make docker-logs      - Посмотреть логи"
	@echo "  make docker-build     - Пересобрать образы"
	@echo ""
	@echo "💻 Локальный запуск:"
	@echo "  make install      - Установить зависимости Go и Python"
	@echo "  make scrape       - Собрать данные с Sketchfab API"
	@echo "  make preprocess   - Предобработать данные"
	@echo "  make eda          - Провести разведочный анализ"
	@echo "  make train        - Обучить ML модель"
	@echo "  make server       - Запустить веб-сервер"
	@echo "  make run-all      - Выполнить все шаги последовательно"
	@echo "  make clean        - Очистить сгенерированные файлы"
	@echo "  make test         - Запустить тесты"

# Docker команды
docker-build:
	@echo "🐳 Сборка Docker образов..."
	docker-compose build

docker-up:
	@echo "🐳 Запуск веб-сервера в Docker..."
	docker-compose up -d web
	@echo "✅ Сервер запущен на http://localhost:8080"

docker-down:
	@echo "🛑 Остановка контейнеров..."
	docker-compose down

docker-logs:
	@echo "📋 Логи сервера:"
	docker-compose logs -f web

docker-pipeline:
	@echo "🐳 Запуск полного пайплайна в Docker..."
	docker-compose --profile tools run --rm scraper
	docker-compose --profile tools run --rm preprocessor
	docker-compose --profile tools run --rm eda
	docker-compose --profile tools run --rm trainer
	docker-compose up -d web
	@echo "✅ Пайплайн завершен! Сервер на http://localhost:8080"

# Локальные команды
install:
	@echo "Установка зависимостей..."
	go mod download
	pip install -r requirements.txt

scrape:
	@echo "Сбор данных..."
	go run cmd/scraper/main.go -limit=500 -sort=likes

preprocess:
	@echo "Предобработка данных..."
	go run cmd/preprocessor/main.go

eda:
	@echo "Разведочный анализ..."
	python scripts/eda.py

train:
	@echo "Обучение модели..."
	python scripts/train_model.py

server:
	@echo "Запуск сервера..."
	go run cmd/server/main.go

run-all: scrape preprocess eda train server

clean:
	@echo "Очистка..."
	rm -f data/*.json
	rm -f data/*.png
	rm -f models/*.pkl
	rm -f models/*.joblib

test:
	@echo "Запуск тестов..."
	go test ./...
