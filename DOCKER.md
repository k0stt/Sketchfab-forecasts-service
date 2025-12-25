# 🐳 Docker - Руководство по использованию

## Быстрый старт

### 1. Запуск веб-сервера (одна команда!)

```bash
docker-compose up -d web
```

Откройте браузер: **http://localhost:8080**

### 2. Остановка

```bash
docker-compose down
```

## Полный цикл работы с данными

### Шаг 1: Сбор данных с Sketchfab API

```bash
docker-compose --profile tools run --rm scraper
```

Параметры можно изменить в `docker-compose.yml` в секции `scraper.command`

### Шаг 2: Предобработка данных

```bash
docker-compose --profile tools run --rm preprocessor
```

### Шаг 3: Разведочный анализ (EDA)

```bash
docker-compose --profile tools run --rm eda
```

Графики будут сохранены в папке `data/`

### Шаг 4: Обучение ML модели

```bash
docker-compose --profile tools run --rm trainer
```

Модель сохраняется в папке `models/`

### Шаг 5: Запуск веб-сервера

```bash
docker-compose up -d web
```

## Автоматический запуск всего пайплайна

Создайте скрипт для последовательного выполнения всех этапов:

**PowerShell:**
```powershell
# Полный пайплайн
docker-compose --profile tools run --rm scraper
docker-compose --profile tools run --rm preprocessor
docker-compose --profile tools run --rm eda
docker-compose --profile tools run --rm trainer
docker-compose up -d web

Write-Host "✅ Проект запущен на http://localhost:8080"
```

**Bash:**
```bash
#!/bin/bash
# Полный пайплайн
docker-compose --profile tools run --rm scraper
docker-compose --profile tools run --rm preprocessor
docker-compose --profile tools run --rm eda
docker-compose --profile tools run --rm trainer
docker-compose up -d web

echo "✅ Проект запущен на http://localhost:8080"
```

## Команды Docker Compose

### Основные команды

```bash
# Запустить веб-сервер
docker-compose up -d web

# Посмотреть логи
docker-compose logs -f web

# Остановить все сервисы
docker-compose down

# Пересобрать образы
docker-compose build

# Пересобрать и запустить
docker-compose up -d --build web

# Удалить все (включая volumes)
docker-compose down -v
```

### Работа с инструментами

```bash
# Запустить scraper
docker-compose --profile tools run --rm scraper

# Запустить preprocessor
docker-compose --profile tools run --rm preprocessor

# Запустить EDA
docker-compose --profile tools run --rm eda

# Запустить обучение модели
docker-compose --profile tools run --rm trainer
```

### Отладка

```bash
# Зайти в контейнер
docker-compose run --rm web sh

# Посмотреть запущенные контейнеры
docker-compose ps

# Посмотреть использование ресурсов
docker stats
```

## Структура сервисов

| Сервис | Описание | Порт | Команда запуска |
|--------|----------|------|-----------------|
| **web** | Веб-сервер с API | 8080 | `docker-compose up -d web` |
| **scraper** | Сбор данных | - | `docker-compose --profile tools run --rm scraper` |
| **preprocessor** | Предобработка | - | `docker-compose --profile tools run --rm preprocessor` |
| **eda** | Анализ данных | - | `docker-compose --profile tools run --rm eda` |
| **trainer** | Обучение ML | - | `docker-compose --profile tools run --rm trainer` |

## Volumes (Данные)

Данные хранятся в локальных папках и монтируются в контейнеры:

- `./data` → `/app/data` - собранные данные и графики
- `./models` → `/app/models` - обученные ML модели

Это позволяет:
- ✅ Сохранять данные между перезапусками
- ✅ Редактировать файлы на хосте
- ✅ Быстро бэкапить данные

## Переменные окружения

Настраиваются в файле `.env`:

```env
SKETCHFAB_API_TOKEN=b893f37576f94e5aab5ab42e0166d0aa
SKETCHFAB_API_URL=https://api.sketchfab.com/v3
PORT=8080
```

## Customization

### Изменить количество собираемых моделей

В `docker-compose.yml` измените:

```yaml
scraper:
  command: /app/bin/scraper -limit=1000 -sort=views
```

### Изменить порт веб-сервера

В `docker-compose.yml`:

```yaml
web:
  ports:
    - "3000:8080"  # хост:контейнер
```

### Добавить свой сервис

```yaml
my-service:
  build: .
  volumes:
    - ./data:/app/data
  command: python my_script.py
  profiles:
    - tools
```

## Преимущества Docker подхода

✅ **Одна команда для запуска** - `docker-compose up -d web`  
✅ **Нет конфликтов зависимостей** - изолированное окружение  
✅ **Работает везде** - Windows, Linux, macOS  
✅ **Легко масштабировать** - можно запустить несколько инстансов  
✅ **Простой деплой** - готово к продакшену  
✅ **Воспроизводимость** - одинаковое окружение для всех  

## Production деплой

### Docker Swarm

```bash
docker stack deploy -c docker-compose.yml sketchfab-forecasts
```

### Kubernetes

Можно сгенерировать манифесты из docker-compose:

```bash
kompose convert
kubectl apply -f .
```

## Troubleshooting

**Проблема:** Порт 8080 занят
```bash
# Измените порт в docker-compose.yml
ports:
  - "3000:8080"
```

**Проблема:** Не хватает памяти
```bash
# Увеличьте лимиты в docker-compose.yml
deploy:
  resources:
    limits:
      memory: 2G
```

**Проблема:** Ошибка доступа к API
```bash
# Проверьте .env файл
cat .env
```

**Проблема:** Образ не пересобирается
```bash
# Принудительная пересборка
docker-compose build --no-cache
```

## Мониторинг

### Логи

```bash
# Все логи
docker-compose logs -f

# Только веб-сервер
docker-compose logs -f web

# Последние 100 строк
docker-compose logs --tail=100 web
```

### Метрики

```bash
# Использование ресурсов
docker stats sketchfab-forecasts-web

# Информация о контейнере
docker inspect sketchfab-forecasts-web
```

## Бэкап данных

```bash
# Создать архив данных
tar -czf backup-$(date +%Y%m%d).tar.gz data/ models/

# Восстановить
tar -xzf backup-20251217.tar.gz
```

---

**🐳 Docker делает проект невероятно простым в использовании!**

Теперь весь проект запускается одной командой:
```bash
docker-compose up -d web
```

И сразу доступен на http://localhost:8080
