# Руководство по использованию

## Быстрый старт

### 1. Установка зависимостей

#### Go зависимости
```bash
go mod download
```

#### Python зависимости
```bash
pip install -r requirements.txt
```

### 2. Настройка

API токен уже настроен в файле `.env`:
```
SKETCHFAB_API_TOKEN=b893f37576f94e5aab5ab42e0166d0aa
```

### 3. Запуск проекта

#### Вариант 1: Полный автоматический запуск (PowerShell)
```powershell
.\run_all.ps1
```

#### Вариант 2: Пошаговый запуск

**Шаг 1: Сбор данных**
```bash
go run cmd/scraper/main.go -limit=500 -sort=likes
```

Параметры:
- `-limit` - количество моделей для сбора (по умолчанию 500)
- `-sort` - сортировка: `likes`, `views`, `recent`
- `-output` - путь к выходному файлу (по умолчанию `data/raw_models.json`)

**Шаг 2: Предобработка данных**
```bash
go run cmd/preprocessor/main.go
```

Параметры:
- `-input` - файл с сырыми данными
- `-output` - файл для обработанных данных
- `-filter` - фильтровать выбросы (по умолчанию true)
- `-threshold` - порог для фильтрации (по умолчанию 3.0)

**Шаг 3: Разведочный анализ (EDA)**
```bash
python scripts/eda.py
```

Результаты:
- Графики сохраняются в `data/`
- Статистический отчет выводится в консоль

**Шаг 4: Обучение модели**
```bash
python scripts/train_model.py
```

Результаты:
- Модель сохраняется в `models/popularity_model.pkl`
- Scaler сохраняется в `models/scaler.pkl`
- Графики в `data/`

**Шаг 5: Запуск веб-сервера**
```bash
go run cmd/server/main.go
```

Сервер запустится на `http://localhost:8080`

## API Endpoints

### 1. Health Check
```http
GET /health
```

Ответ:
```json
{
  "status": "healthy",
  "service": "sketchfab-forecasts"
}
```

### 2. Прогнозирование популярности
```http
POST /api/predict
Content-Type: application/json

{
  "category_count": 2,
  "tag_count": 8,
  "description_length": 250,
  "face_count": 10000,
  "vertex_count": 5000,
  "animation_count": 0,
  "is_downloadable": true,
  "is_premium_author": false,
  "author_followers": 150
}
```

Ответ:
```json
{
  "popularity_score": 3.45,
  "category": "medium",
  "confidence": 0.8
}
```

Категории популярности:
- `low` - низкая (< 2.0)
- `medium` - средняя (2.0 - 4.0)
- `high` - высокая (> 4.0)

### 3. Статистика
```http
GET /api/stats
```

Ответ:
```json
{
  "total_models": 500,
  "average_views": 1250,
  "average_likes": 89,
  "average_downloads": 45,
  "top_categories": [...],
  "top_tags": [...]
}
```

## Примеры использования

### cURL

**Прогнозирование:**
```bash
curl -X POST http://localhost:8080/api/predict \
  -H "Content-Type: application/json" \
  -d '{
    "category_count": 3,
    "tag_count": 10,
    "description_length": 300,
    "face_count": 15000,
    "vertex_count": 8000,
    "animation_count": 1,
    "is_downloadable": true,
    "is_premium_author": true,
    "author_followers": 500
  }'
```

**Статистика:**
```bash
curl http://localhost:8080/api/stats
```

### Python

```python
import requests

# Прогнозирование
data = {
    "category_count": 2,
    "tag_count": 8,
    "description_length": 250,
    "face_count": 10000,
    "vertex_count": 5000,
    "animation_count": 0,
    "is_downloadable": True,
    "is_premium_author": False,
    "author_followers": 150
}

response = requests.post('http://localhost:8080/api/predict', json=data)
result = response.json()

print(f"Популярность: {result['popularity_score']:.2f}")
print(f"Категория: {result['category']}")
print(f"Уверенность: {result['confidence']:.0%}")
```

### JavaScript

```javascript
// Прогнозирование
const data = {
  category_count: 2,
  tag_count: 8,
  description_length: 250,
  face_count: 10000,
  vertex_count: 5000,
  animation_count: 0,
  is_downloadable: true,
  is_premium_author: false,
  author_followers: 150
};

fetch('http://localhost:8080/api/predict', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json'
  },
  body: JSON.stringify(data)
})
  .then(response => response.json())
  .then(result => {
    console.log('Популярность:', result.popularity_score);
    console.log('Категория:', result.category);
    console.log('Уверенность:', result.confidence);
  });
```

## Структура данных

### Входные признаки

| Признак | Тип | Описание | Диапазон |
|---------|-----|----------|----------|
| `category_count` | int | Количество категорий | 0-10 |
| `tag_count` | int | Количество тегов | 0-50 |
| `description_length` | int | Длина описания | 0-5000 |
| `face_count` | int | Количество полигонов | 0+ |
| `vertex_count` | int | Количество вершин | 0+ |
| `animation_count` | int | Количество анимаций | 0+ |
| `is_downloadable` | bool | Доступна для скачивания | true/false |
| `is_premium_author` | bool | Премиум аккаунт автора | true/false |
| `author_followers` | int | Подписчики автора | 0+ |

### Формула популярности

```
popularity_score = 0.3 * log(views + 1) + 
                   0.4 * log(likes + 1) + 
                   0.3 * log(downloads + 1)
```

## Troubleshooting

### Ошибка: "SKETCHFAB_API_TOKEN не установлен"
Убедитесь, что файл `.env` существует и содержит токен.

### Ошибка: "API returned status 429"
Превышен лимит запросов к API. Подождите минуту и повторите.

### Ошибка: "Python script error"
Убедитесь, что установлены все Python зависимости:
```bash
pip install -r requirements.txt
```

### Порт 8080 уже занят
Измените порт в `.env`:
```
PORT=3000
```

## Дополнительная информация

- [Документация Sketchfab API](https://docs.sketchfab.com/data-api/)
- [Scikit-learn Documentation](https://scikit-learn.org/)
- [Chi Router](https://github.com/go-chi/chi)
