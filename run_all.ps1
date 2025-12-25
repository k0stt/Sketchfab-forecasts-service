# Скрипт для полного запуска проекта (Windows PowerShell)

Write-Host "=== Sketchfab Forecasts - Полный запуск ===" -ForegroundColor Cyan
Write-Host ""

# Шаг 1: Сбор данных
Write-Host "1. Сбор данных с Sketchfab API..." -ForegroundColor Yellow
go run cmd/scraper/main.go -limit=500 -sort=likes
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка при сборе данных!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Данные собраны" -ForegroundColor Green
Write-Host ""

# Шаг 2: Предобработка
Write-Host "2. Предобработка данных..." -ForegroundColor Yellow
go run cmd/preprocessor/main.go
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка при предобработке!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Данные обработаны" -ForegroundColor Green
Write-Host ""

# Шаг 3: EDA
Write-Host "3. Разведочный анализ данных (EDA)..." -ForegroundColor Yellow
python scripts/eda.py
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка при EDA!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Анализ завершен" -ForegroundColor Green
Write-Host ""

# Шаг 4: Обучение модели
Write-Host "4. Обучение модели машинного обучения..." -ForegroundColor Yellow
python scripts/train_model.py
if ($LASTEXITCODE -ne 0) {
    Write-Host "Ошибка при обучении модели!" -ForegroundColor Red
    exit 1
}
Write-Host "✓ Модель обучена" -ForegroundColor Green
Write-Host ""

# Шаг 5: Запуск сервера
Write-Host "5. Запуск веб-сервера..." -ForegroundColor Yellow
Write-Host "Сервер будет доступен на http://localhost:8080" -ForegroundColor Green
go run cmd/server/main.go
