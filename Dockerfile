# Multi-stage build для оптимизации размера образа

# Stage 1: Build Go приложения
FROM golang:1.21-alpine AS go-builder

WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем весь код
COPY . .

# Собираем все Go бинарники
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/scraper ./cmd/scraper
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/preprocessor ./cmd/preprocessor
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/server ./cmd/server

# Stage 2: Production образ
FROM python:3.11-slim

WORKDIR /app

# Устанавливаем системные зависимости
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Копируем Python зависимости и устанавливаем их
COPY requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

# Копируем Go бинарники из builder
COPY --from=go-builder /app/bin/scraper /app/bin/scraper
COPY --from=go-builder /app/bin/preprocessor /app/bin/preprocessor
COPY --from=go-builder /app/bin/server /app/bin/server

# Копируем Python скрипты
COPY scripts/ ./scripts/

# Копируем необходимые файлы
COPY .env .env

# Создаем директории для данных и моделей
RUN mkdir -p /app/data /app/models

# Открываем порт для веб-сервера
EXPOSE 8080

# По умолчанию запускаем веб-сервер
CMD ["/app/bin/server"]
