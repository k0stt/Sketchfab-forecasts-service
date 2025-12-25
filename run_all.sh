#!/bin/bash

# Скрипт для полного запуска проекта

echo "=== Sketchfab Forecasts - Полный запуск ==="
echo ""

# Шаг 1: Сбор данных
echo "1. Сбор данных с Sketchfab API..."
go run cmd/scraper/main.go -limit=500 -sort=likes
if [ $? -ne 0 ]; then
    echo "Ошибка при сборе данных!"
    exit 1
fi
echo "✓ Данные собраны"
echo ""

# Шаг 2: Предобработка
echo "2. Предобработка данных..."
go run cmd/preprocessor/main.go
if [ $? -ne 0 ]; then
    echo "Ошибка при предобработке!"
    exit 1
fi
echo "✓ Данные обработаны"
echo ""

# Шаг 3: EDA
echo "3. Разведочный анализ данных (EDA)..."
python scripts/eda.py
if [ $? -ne 0 ]; then
    echo "Ошибка при EDA!"
    exit 1
fi
echo "✓ Анализ завершен"
echo ""

# Шаг 4: Обучение модели
echo "4. Обучение модели машинного обучения..."
python scripts/train_model.py
if [ $? -ne 0 ]; then
    echo "Ошибка при обучении модели!"
    exit 1
fi
echo "✓ Модель обучена"
echo ""

# Шаг 5: Запуск сервера
echo "5. Запуск веб-сервера..."
echo "Сервер будет доступен на http://localhost:8080"
go run cmd/server/main.go
