# 🧪 Тест API с новыми возможностями

Write-Host "=== Тестирование Sketchfab Forecasts API ===" -ForegroundColor Cyan
Write-Host ""

# Тест 1: Модель с оптимальными полигонами
Write-Host "📊 Тест 1: Оптимальные полигоны (15,000)" -ForegroundColor Green
$body1 = @{
    tags = @("lowpoly", "pbr", "game", "character")
    description = "High quality game character with PBR textures. Optimized for real-time rendering with detailed normal maps."
    category_count = 2
    tag_count = 4
    description_length = 120
    face_count = 15000
    vertex_count = 8000
    animation_count = 2
    is_downloadable = $true
    is_premium_author = $true
    author_followers = 500
} | ConvertTo-Json

$result1 = Invoke-RestMethod -Uri http://localhost:8080/api/predict -Method Post -Body $body1 -ContentType "application/json"
Write-Host "  Популярность: $($result1.popularity_score.ToString('F2')) ($($result1.popularity_category))" -ForegroundColor Yellow
if ($result1.quality_rating) {
    Write-Host "  Рейтинг качества: $($result1.quality_rating.score.ToString('F1'))/100 ($($result1.quality_rating.grade))" -ForegroundColor Cyan
}
Write-Host ""

# Тест 2: Слишком мало полигонов
Write-Host "📊 Тест 2: Мало полигонов (500)" -ForegroundColor Green
$body2 = @{
    tags = @("simple", "lowpoly")
    description = "Very simple model"
    category_count = 1
    tag_count = 2
    description_length = 20
    face_count = 500
    vertex_count = 250
    animation_count = 0
    is_downloadable = $false
    is_premium_author = $false
    author_followers = 10
} | ConvertTo-Json

$result2 = Invoke-RestMethod -Uri http://localhost:8080/api/predict -Method Post -Body $body2 -ContentType "application/json"
Write-Host "  Популярность: $($result2.popularity_score.ToString('F2')) ($($result2.popularity_category))" -ForegroundColor Yellow
if ($result2.quality_rating) {
    Write-Host "  Рейтинг качества: $($result2.quality_rating.score.ToString('F1'))/100 ($($result2.quality_rating.grade))" -ForegroundColor Cyan
}
Write-Host ""

# Тест 3: Слишком много полигонов
Write-Host "📊 Тест 3: Много полигонов (150,000)" -ForegroundColor Green
$body3 = @{
    tags = @("highpoly", "detailed", "showcase")
    description = "Ultra detailed high poly model with millions of details"
    category_count = 1
    tag_count = 3
    description_length = 55
    face_count = 150000
    vertex_count = 75000
    animation_count = 0
    is_downloadable = $true
    is_premium_author = $true
    author_followers = 1000
} | ConvertTo-Json

$result3 = Invoke-RestMethod -Uri http://localhost:8080/api/predict -Method Post -Body $body3 -ContentType "application/json"
Write-Host "  Популярность: $($result3.popularity_score.ToString('F2')) ($($result3.popularity_category))" -ForegroundColor Yellow
if ($result3.quality_rating) {
    Write-Host "  Рейтинг качества: $($result3.quality_rating.score.ToString('F1'))/100 ($($result3.quality_rating.grade))" -ForegroundColor Cyan
}
Write-Host ""

# Тест 4: Идеальная модель
Write-Host "📊 Тест 4: Идеальная модель" -ForegroundColor Green
$body4 = @{
    tags = @("lowpoly", "pbr", "game", "character", "rigged", "animated", "unity", "unreal")
    description = "Professional game-ready character model. Features: PBR textures (4K), fully rigged skeleton, 10 animations included, LOD levels, optimized UV maps. Perfect for mobile and desktop games. Includes diffuse, normal, roughness, metallic, and AO maps."
    category_count = 3
    tag_count = 8
    description_length = 230
    face_count = 25000
    vertex_count = 12500
    animation_count = 10
    is_downloadable = $true
    is_premium_author = $true
    author_followers = 2000
    category = "game_desktop"
    account_type = "premium"
    has_textures = $true
    has_pbr = $true
    is_rigged = $true
    is_animated = $true
} | ConvertTo-Json

$result4 = Invoke-RestMethod -Uri http://localhost:8080/api/predict -Method Post -Body $body4 -ContentType "application/json"
Write-Host "  Популярность: $($result4.popularity_score.ToString('F2')) ($($result4.popularity_category))" -ForegroundColor Yellow
if ($result4.quality_rating) {
    Write-Host "  Рейтинг качества: $($result4.quality_rating.score.ToString('F1'))/100 ($($result4.quality_rating.grade))" -ForegroundColor Cyan
    
    if ($result4.quality_rating.recommendations -and $result4.quality_rating.recommendations.Count -gt 0) {
        Write-Host "  Рекомендации:" -ForegroundColor Magenta
        foreach ($rec in $result4.quality_rating.recommendations) {
            Write-Host "    • $rec" -ForegroundColor Gray
        }
    } else {
        Write-Host "  ✅ Отлично! Модель соответствует высоким стандартам." -ForegroundColor Green
    }
}
Write-Host ""

# Сравнение результатов
Write-Host "=== Сравнение ===" -ForegroundColor Cyan
Write-Host "Тест 1 (оптимум):  $($result1.popularity_score.ToString('F2'))" -ForegroundColor Green
Write-Host "Тест 2 (мало):     $($result2.popularity_score.ToString('F2'))" -ForegroundColor Yellow
Write-Host "Тест 3 (много):    $($result3.popularity_score.ToString('F2'))" -ForegroundColor Yellow
Write-Host "Тест 4 (идеал):    $($result4.popularity_score.ToString('F2'))" -ForegroundColor Green
Write-Host ""

Write-Host "✅ Все тесты завершены!" -ForegroundColor Green
Write-Host "📖 Откройте http://localhost:8080 для веб-интерфейса" -ForegroundColor Cyan
