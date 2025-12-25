<#
Simple Docker pipeline launcher (ASCII-only to avoid encoding issues)
#>

Write-Host "Sketchfab Forecasts - Docker Pipeline" -ForegroundColor Cyan
Write-Host ""

Write-Host "Checking Docker..." -ForegroundColor Yellow
if (-not (Get-Command docker -ErrorAction SilentlyContinue)) {
    Write-Host "Docker not found. Install Docker Desktop: https://www.docker.com/products/docker-desktop" -ForegroundColor Red
    exit 1
}

Write-Host "Docker found." -ForegroundColor Green
Write-Host ""

Write-Host "Select run mode:" -ForegroundColor Cyan
Write-Host "1. Quick start (web server only)" -ForegroundColor White
Write-Host "2. Full pipeline (scrape + preprocess + train + server)" -ForegroundColor White
Write-Host ""

$choice = Read-Host "Enter number (1 or 2)"

if ($choice -eq "1") {
    Write-Host "Starting web server..." -ForegroundColor Yellow
    docker-compose up -d --build web
    if ($LASTEXITCODE -eq 0) {
        Write-Host "Server started successfully." -ForegroundColor Green
        Write-Host "Web: http://localhost:8080" -ForegroundColor Cyan
        Write-Host "API: http://localhost:8080/api/predict" -ForegroundColor Cyan
        Write-Host "To stop: docker-compose down" -ForegroundColor Gray
    } else {
        Write-Host "Error starting server." -ForegroundColor Red
        exit 1
    }

} elseif ($choice -eq "2") {
    Write-Host "Starting full pipeline..." -ForegroundColor Yellow
    Write-Host "" -ForegroundColor White
    Write-Host "How many models to collect for training?" -ForegroundColor Cyan
    Write-Host "  Recommended: 500-1000 (better accuracy)" -ForegroundColor Gray
    Write-Host "  Quick test: 100-200" -ForegroundColor Gray
    Write-Host "" -ForegroundColor White
    $modelCount = Read-Host "Enter number (default: 500)"
    if ([string]::IsNullOrWhiteSpace($modelCount)) { $modelCount = "500" }
    Write-Host "Will collect $modelCount models" -ForegroundColor Green
    Write-Host "This may take several minutes." -ForegroundColor Gray

    Write-Host "Step 1: Scrape data..." -ForegroundColor Yellow
    docker-compose --profile tools run --rm scraper sh -c "/app/bin/scraper -limit $modelCount -sort likes"
    if ($LASTEXITCODE -ne 0) { Write-Host "Scrape failed." -ForegroundColor Red; exit 1 }
    Write-Host "Scrape complete." -ForegroundColor Green

    Write-Host "Step 2: Preprocess data..." -ForegroundColor Yellow
    docker-compose --profile tools run --rm preprocessor
    if ($LASTEXITCODE -ne 0) { Write-Host "Preprocess failed." -ForegroundColor Red; exit 1 }
    Write-Host "Preprocess complete." -ForegroundColor Green

    Write-Host "Step 3: EDA..." -ForegroundColor Yellow
    docker-compose --profile tools run --rm eda
    if ($LASTEXITCODE -ne 0) { Write-Host "EDA failed." -ForegroundColor Red; exit 1 }
    Write-Host "EDA complete. Charts saved to ./data/" -ForegroundColor Green

    Write-Host "Step 4: Train model..." -ForegroundColor Yellow
    docker-compose --profile tools run --rm trainer
    if ($LASTEXITCODE -ne 0) { Write-Host "Training failed." -ForegroundColor Red; exit 1 }
    Write-Host "Model trained and saved to ./models/" -ForegroundColor Green

    Write-Host "Step 5: Start web server..." -ForegroundColor Yellow
    docker-compose up -d --build web
    if ($LASTEXITCODE -ne 0) { Write-Host "Error starting server." -ForegroundColor Red; exit 1 }
    Write-Host "Server started." -ForegroundColor Green

    Write-Host "Pipeline finished successfully." -ForegroundColor Green
    Write-Host "Web: http://localhost:8080" -ForegroundColor Cyan
    Write-Host "API: http://localhost:8080/api/predict" -ForegroundColor Cyan
    Write-Host "EDA charts: ./data/*.png" -ForegroundColor Cyan
    Write-Host "Model: ./models/popularity_model.pkl" -ForegroundColor Cyan
    Write-Host "To stop: docker-compose down" -ForegroundColor Gray

} else {
    Write-Host "Invalid choice." -ForegroundColor Red
    exit 1
}
