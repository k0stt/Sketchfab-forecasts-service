#!/bin/bash

echo "🐳 Sketchfab Forecasts - Docker Pipeline"
echo ""

# Проверка Docker
echo "Проверка Docker..."
if ! command -v docker &> /dev/null; then
    echo "❌ Docker не установлен!"
    echo "Установите Docker: https://docs.docker.com/get-docker/"
    exit 1
fi

echo "✅ Docker найден"
echo ""

# Выбор режима
echo "Выберите режим запуска:"
echo "1. Быстрый старт (только веб-сервер)"
echo "2. Полный пайплайн (сбор данных + обучение + сервер)"
echo ""

read -p "Введите номер (1 или 2): " choice

if [ "$choice" == "1" ]; then
    echo ""
    echo "🚀 Запуск веб-сервера..."
    docker-compose up -d --build web
    
    if [ $? -eq 0 ]; then
        echo ""
        echo "✅ Сервер успешно запущен!"
        echo ""
        echo "🌐 Веб-интерфейс: http://localhost:8080"
        echo "📡 API endpoint: http://localhost:8080/api/predict"
        echo "❤️  Health check: http://localhost:8080/health"
        echo ""
        echo "Для остановки: docker-compose down"
    else
        echo "❌ Ошибка при запуске сервера"
    fi
    
elif [ "$choice" == "2" ]; then
    echo ""
    echo "🔄 Запуск полного пайплайна..."
    echo "Это займет 10-15 минут"
    echo ""
    
    # Шаг 1: Сбор данных
    echo "1️⃣  Сбор данных с Sketchfab API..."
    docker-compose --profile tools run --rm scraper
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при сборе данных"
        exit 1
    fi
    echo "✅ Данные собраны"
    echo ""
    
    # Шаг 2: Предобработка
    echo "2️⃣  Предобработка данных..."
    docker-compose --profile tools run --rm preprocessor
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при предобработке"
        exit 1
    fi
    echo "✅ Данные обработаны"
    echo ""
    
    # Шаг 3: EDA
    echo "3️⃣  Разведочный анализ данных (EDA)..."
    docker-compose --profile tools run --rm eda
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при анализе"
        exit 1
    fi
    echo "✅ Анализ завершен, графики в папке data/"
    echo ""
    
    # Шаг 4: Обучение модели
    echo "4️⃣  Обучение ML модели..."
    docker-compose --profile tools run --rm trainer
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при обучении"
        exit 1
    fi
    echo "✅ Модель обучена и сохранена в models/"
    echo ""
    
    # Шаг 5: Запуск сервера
    echo "5️⃣  Запуск веб-сервера..."
    docker-compose up -d --build web
    if [ $? -ne 0 ]; then
        echo "❌ Ошибка при запуске сервера"
        exit 1
    fi
    echo "✅ Сервер запущен"
    echo ""
    
    echo "🎉 Пайплайн успешно завершен!"
    echo ""
    echo "🌐 Веб-интерфейс: http://localhost:8080"
    echo "📡 API endpoint: http://localhost:8080/api/predict"
    echo "📊 Графики EDA: ./data/*.png"
    echo "🤖 ML модель: ./models/popularity_model.pkl"
    echo ""
    echo "Для остановки: docker-compose down"
    
else
    echo "❌ Неверный выбор"
    exit 1
fi
