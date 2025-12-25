package main

import (
	"encoding/json"
	"net/http"
	"os"
	"sketchfab-forecasts/internal/ml"
	"sketchfab-forecasts/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
	"github.com/sirupsen/logrus"
)

type Server struct {
	router    *chi.Mux
	logger    *logrus.Logger
	predictor *ml.Predictor
}

func NewServer(logger *logrus.Logger) *Server {
	s := &Server{
		router:    chi.NewRouter(),
		logger:    logger,
		predictor: ml.NewPredictor(logger),
	}

	s.setupMiddleware()
	s.setupRoutes()

	return s
}

func (s *Server) setupMiddleware() {
	s.router.Use(middleware.Logger)
	s.router.Use(middleware.Recoverer)
	s.router.Use(middleware.RequestID)
	s.router.Use(corsMiddleware)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) setupRoutes() {
	// Health check
	s.router.Get("/health", s.handleHealth)

	// API routes
	s.router.Route("/api", func(r chi.Router) {
		r.Post("/predict", s.handlePredict)
		r.Get("/stats", s.handleStats)
		r.Get("/model-info", s.handleModelInfo)
		r.Get("/eda-charts", s.handleEdaCharts)
	})

	// Serve static files (EDA charts and other data)
	fs := http.FileServer(http.Dir("data"))
	s.router.Handle("/data/*", http.StripPrefix("/data", fs))

	// Serve frontend
	s.router.Get("/", s.handleIndex)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	respondJSON(w, http.StatusOK, map[string]string{
		"status":  "healthy",
		"service": "sketchfab-forecasts",
	})
}

func (s *Server) handlePredict(w http.ResponseWriter, r *http.Request) {
	var req models.PredictionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.logger.Errorf("Failed to decode request: %v", err)
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	s.logger.Infof("Prediction request: %+v", req)

	// Используем mock предсказание (без Python, для быстрого запуска)
	// Для использования реальной ML модели замените на: prediction, err := s.predictor.Predict(req)
	prediction := s.predictor.MockPredict(req)

	s.logger.Infof("Prediction result: %+v", prediction)

	respondJSON(w, http.StatusOK, prediction)
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	// В реальном приложении здесь бы была загрузка статистики из БД
	stats := models.Stats{
		TotalModels:      500,
		AverageViews:     1250,
		AverageLikes:     89,
		AverageDownloads: 45,
		TopCategories: []models.CategoryStat{
			{Name: "Characters", Count: 120, AvgPopularity: 4.5},
			{Name: "Architecture", Count: 95, AvgPopularity: 3.8},
			{Name: "Animals", Count: 78, AvgPopularity: 4.2},
		},
		TopTags: []models.TagStat{
			{Name: "lowpoly", Count: 245},
			{Name: "pbr", Count: 189},
			{Name: "game-ready", Count: 156},
		},
	}

	respondJSON(w, http.StatusOK, stats)
}

func (s *Server) handleIndex(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html lang="ru">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Sketchfab Forecasts - ML Прогнозирование популярности</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        body {
            font-family: 'Segoe UI', Tahoma, Geneva, Verdana, sans-serif;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            min-height: 100vh;
            padding: 20px;
        }
        .container {
            max-width: 1200px;
            margin: 0 auto;
        }
        .card {
            background: white;
            border-radius: 15px;
            padding: 30px;
            margin-bottom: 20px;
            box-shadow: 0 10px 30px rgba(0,0,0,0.2);
        }
        h1 {
            color: #667eea;
            margin-bottom: 10px;
            text-align: center;
        }
        h2 {
            color: #667eea;
            margin-bottom: 20px;
            font-size: 20px;
        }
        .subtitle {
            text-align: center;
            color: #666;
            margin-bottom: 30px;
        }
        .tabs {
            display: flex;
            gap: 10px;
            margin-bottom: 20px;
            border-bottom: 2px solid #e0e0e0;
        }
        .tab {
            padding: 10px 20px;
            cursor: pointer;
            border: none;
            background: none;
            color: #666;
            font-weight: 600;
            transition: all 0.3s;
        }
        .tab.active {
            color: #667eea;
            border-bottom: 3px solid #667eea;
        }
        .tab-content {
            display: none;
        }
        .tab-content.active {
            display: block;
        }
        .model-info {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
            gap: 15px;
            margin-top: 20px;
        }
        .metric {
            padding: 15px;
            background: #f8f9fa;
            border-radius: 8px;
            text-align: center;
        }
        .metric-value {
            font-size: 24px;
            font-weight: bold;
            color: #667eea;
        }
        .metric-label {
            color: #666;
            font-size: 12px;
            margin-top: 5px;
        }
        .status-badge {
            display: inline-block;
            padding: 5px 15px;
            border-radius: 20px;
            font-size: 12px;
            font-weight: 600;
        }
        .status-trained {
            background: #4caf50;
            color: white;
        }
        .status-untrained {
            background: #ff9800;
            color: white;
        }
        .form-group {
            margin-bottom: 15px;
        }
        label {
            display: block;
            margin-bottom: 5px;
            color: #333;
            font-weight: 600;
            font-size: 14px;
        }
        input[type="number"] {
            width: 100%;
            padding: 10px;
            border: 2px solid #e0e0e0;
            border-radius: 8px;
            font-size: 14px;
        }
        input:focus {
            outline: none;
            border-color: #667eea;
        }
        .checkbox-group {
            display: flex;
            align-items: center;
            gap: 10px;
        }
        input[type="checkbox"] {
            width: 18px;
            height: 18px;
            cursor: pointer;
        }
        button {
            width: 100%;
            padding: 15px;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
            border: none;
            border-radius: 8px;
            font-size: 16px;
            font-weight: 600;
            cursor: pointer;
            transition: transform 0.2s;
        }
        button:hover {
            transform: translateY(-2px);
        }
        .result {
            margin-top: 20px;
            padding: 20px;
            background: #f8f9fa;
            border-radius: 10px;
            display: none;
        }
        .result.show {
            display: block;
        }
        .score {
            font-size: 48px;
            font-weight: bold;
            text-align: center;
            margin: 20px 0;
        }
        .score.low { color: #f44336; }
        .score.medium { color: #ff9800; }
        .score.high { color: #4caf50; }
        .charts-grid {
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(350px, 1fr));
            gap: 20px;
            margin-top: 20px;
        }
        .chart-card {
            background: #f8f9fa;
            border-radius: 10px;
            padding: 15px;
            cursor: pointer;
            transition: transform 0.2s;
        }
        .chart-card:hover {
            transform: scale(1.02);
        }
        .chart-card img {
            width: 100%;
            border-radius: 8px;
        }
        .chart-name {
            margin-top: 10px;
            font-weight: 600;
            color: #333;
            font-size: 14px;
        }
        .loading {
            text-align: center;
            color: #666;
            padding: 20px;
        }
        .form-columns {
            display: grid;
            grid-template-columns: repeat(2, 1fr);
            gap: 15px;
        }
        @media (max-width: 768px) {
            .form-columns {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="card">
            <h1>🎨 Sketchfab Forecasts</h1>
            <p class="subtitle">ML-прогнозирование популярности 3D-моделей</p>
            
            <div class="tabs">
                <button class="tab active" onclick="switchTab('prediction')">🔮 Прогноз</button>
                <button class="tab" onclick="switchTab('model')">🤖 Модель</button>
                <button class="tab" onclick="switchTab('eda')">📊 Анализ данных</button>
            </div>
            
            <div id="prediction-tab" class="tab-content active">
                <form id="predictionForm" class="form-columns">
                    <div class="form-group">
                        <label>Категории:</label>
                        <input type="number" id="categoryCount" min="0" max="10" value="2">
                    </div>
                    <div class="form-group">
                        <label>Теги:</label>
                        <input type="number" id="tagCount" min="0" max="50" value="8">
                    </div>
                    <div class="form-group">
                        <label>Длина описания:</label>
                        <input type="number" id="descriptionLength" min="0" max="5000" value="250">
                    </div>
                    <div class="form-group">
                        <label>Полигоны (faces):</label>
                        <input type="number" id="faceCount" min="0" value="10000">
                    </div>
                    <div class="form-group">
                        <label>Вершины (vertices):</label>
                        <input type="number" id="vertexCount" min="0" value="5000">
                    </div>
                    <div class="form-group">
                        <label>Анимации:</label>
                        <input type="number" id="animationCount" min="0" value="0">
                    </div>
                    <div class="form-group">
                        <label>Подписчики автора:</label>
                        <input type="number" id="authorFollowers" min="0" value="150">
                    </div>
                    <div class="form-group">
                        <div class="checkbox-group">
                            <input type="checkbox" id="isDownloadable" checked>
                            <label for="isDownloadable">Доступна для скачивания</label>
                        </div>
                        <div class="checkbox-group" style="margin-top:10px">
                            <input type="checkbox" id="isPremiumAuthor">
                            <label for="isPremiumAuthor">Премиум автор</label>
                        </div>
                    </div>
                </form>
                <button onclick="predict()" style="margin-top:20px">🔮 Получить прогноз</button>
                <div class="result" id="result">
                    <div style="text-align:center; font-size:20px; font-weight:600; margin-bottom:10px" id="category"></div>
                    <div class="score" id="score"></div>
                    <div style="text-align:center; color:#666" id="confidence"></div>
                </div>
            </div>
            
            <div id="model-tab" class="tab-content">
                <div id="modelStatus" class="loading">Загрузка информации о модели...</div>
                <div id="modelInfo" style="display:none">
                    <div class="model-info"></div>
                </div>
            </div>
            
            <div id="eda-tab" class="tab-content">
                <div id="chartsLoading" class="loading">Загрузка графиков...</div>
                <div id="chartsGrid" class="charts-grid" style="display:none"></div>
            </div>
        </div>
    </div>
    
    <script>
        function switchTab(tab) {
            document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            event.target.classList.add('active');
            document.getElementById(tab + '-tab').classList.add('active');
            
            if (tab === 'model') loadModelInfo();
            if (tab === 'eda') loadCharts();
        }
        
        async function predict() {
            const data = {
                category_count: parseInt(document.getElementById('categoryCount').value),
                tag_count: parseInt(document.getElementById('tagCount').value),
                description_length: parseInt(document.getElementById('descriptionLength').value),
                face_count: parseInt(document.getElementById('faceCount').value),
                vertex_count: parseInt(document.getElementById('vertexCount').value),
                animation_count: parseInt(document.getElementById('animationCount').value),
                author_followers: parseInt(document.getElementById('authorFollowers').value),
                is_downloadable: document.getElementById('isDownloadable').checked,
                is_premium_author: document.getElementById('isPremiumAuthor').checked
            };
            
            try {
                const response = await fetch('/api/predict', {
                    method: 'POST',
                    headers: {'Content-Type': 'application/json'},
                    body: JSON.stringify(data)
                });
                const result = await response.json();
                
                const categoryText = {
                    'low': '📉 Низкая популярность',
                    'medium': '📊 Средняя популярность',
                    'high': '📈 Высокая популярность'
                };
                
                document.getElementById('category').textContent = categoryText[result.category];
                document.getElementById('score').textContent = result.popularity_score.toFixed(2);
                document.getElementById('score').className = 'score ' + result.category;
                document.getElementById('confidence').textContent = 
                    'Уверенность: ' + (result.confidence * 100).toFixed(0) + '%';
                document.getElementById('result').classList.add('show');
            } catch (error) {
                alert('Ошибка: ' + error.message);
            }
        }
        
        async function loadModelInfo() {
            try {
                const response = await fetch('/api/model-info');
                const data = await response.json();
                
                const info = document.querySelector('.model-info');
                if (data.trained) {
                    info.innerHTML = 
                        '<div class="metric">' +
                            '<div class="status-badge status-trained">✓ Обучена</div>' +
                        '</div>' +
                        '<div class="metric">' +
                            '<div class="metric-value">' + (data.training_samples || 'N/A') + '</div>' +
                            '<div class="metric-label">Примеров обучения</div>' +
                        '</div>' +
                        '<div class="metric">' +
                            '<div class="metric-value">' + (data.r2_score || 0).toFixed(3) + '</div>' +
                            '<div class="metric-label">R² Score</div>' +
                        '</div>' +
                        '<div class="metric">' +
                            '<div class="metric-value">' + (data.rmse || 0).toFixed(3) + '</div>' +
                            '<div class="metric-label">RMSE</div>' +
                        '</div>' +
                        '<div class="metric">' +
                            '<div class="metric-value">' + (data.mae || 0).toFixed(3) + '</div>' +
                            '<div class="metric-label">MAE</div>' +
                        '</div>' +
                        '<div class="metric">' +
                            '<div class="metric-value">' + (data.training_date || 'N/A') + '</div>' +
                            '<div class="metric-label">Дата обучения</div>' +
                        '</div>';
                } else {
                    info.innerHTML = 
                        '<div class="metric">' +
                            '<div class="status-badge status-untrained">⚠ Не обучена</div>' +
                        '</div>' +
                        '<div style="grid-column: 1/-1; text-align:center; color:#666; margin-top:10px">' +
                            'Модель использует эвристический алгоритм. Запустите обучение для улучшения точности.' +
                        '</div>';
                }
                
                document.getElementById('modelStatus').style.display = 'none';
                document.getElementById('modelInfo').style.display = 'block';
            } catch (error) {
                document.getElementById('modelStatus').textContent = 'Ошибка загрузки: ' + error.message;
            }
        }
        
        async function loadCharts() {
            try {
                const response = await fetch('/api/eda-charts');
                const data = await response.json();
                
                const grid = document.getElementById('chartsGrid');
                if (data.count === 0) {
                    document.getElementById('chartsLoading').textContent = 
                        'Графики не найдены. Запустите полный пайплайн для генерации EDA.';
                    return;
                }
                
                grid.innerHTML = data.charts.map(function(chart) {
                    return '<div class="chart-card" onclick="window.open(\'' + chart.url + '\', \'_blank\')">' +
                        '<img src="' + chart.url + '" alt="' + chart.name + '">' +
                        '<div class="chart-name">' + chart.name.replace('.png', '').replace(/_/g, ' ') + '</div>' +
                    '</div>';
                }).join('');
                
                document.getElementById('chartsLoading').style.display = 'none';
                grid.style.display = 'grid';
            } catch (error) {
                document.getElementById('chartsLoading').textContent = 'Ошибка: ' + error.message;
            }
        }
        
        document.getElementById('predictionForm').addEventListener('submit', e => {
            e.preventDefault();
            predict();
        });
    </script>
</body>
</html>
	`

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{"error": message})
}

func (s *Server) Start(port string) error {
	s.logger.Infof("Starting server on port %s", port)
	return http.ListenAndServe(":"+port, s.router)
}

func (s *Server) handleModelInfo(w http.ResponseWriter, r *http.Request) {
	// Load model metrics from file if it exists
	type ModelMetrics struct {
		Trained         bool     `json:"trained"`
		TrainingDate    string   `json:"training_date,omitempty"`
		RMSE            float64  `json:"rmse,omitempty"`
		MAE             float64  `json:"mae,omitempty"`
		R2Score         float64  `json:"r2_score,omitempty"`
		TrainingSamples int      `json:"training_samples,omitempty"`
		Features        []string `json:"features,omitempty"`
	}

	metrics := ModelMetrics{Trained: false}

	// Try to load model info from models/model_metrics.json
	if data, err := os.ReadFile("models/model_metrics.json"); err == nil {
		json.Unmarshal(data, &metrics)
		metrics.Trained = true
	}

	respondJSON(w, http.StatusOK, metrics)
}

func (s *Server) handleEdaCharts(w http.ResponseWriter, r *http.Request) {
	// List available EDA charts
	charts := []string{
		"eda_popularity_distribution.png",
		"eda_correlation_matrix.png",
		"eda_categorical_features.png",
		"eda_complexity_analysis.png",
		"eda_temporal_trends.png",
		"eda_tags_categories.png",
		"feature_importance.png",
		"predictions_random_forest.png",
	}

	available := []map[string]string{}
	for _, chart := range charts {
		path := "data/" + chart
		if _, err := os.Stat(path); err == nil {
			available = append(available, map[string]string{
				"name": chart,
				"url":  "/data/" + chart,
			})
		}
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"charts": available,
		"count":  len(available),
	})
}

func main() {
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

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Создание и запуск сервера
	server := NewServer(logger)

	logger.Info("=== Sketchfab Forecasts Service ===")
	logger.Infof("Web интерфейс: http://localhost:%s", port)
	logger.Infof("API endpoint: http://localhost:%s/api/predict", port)
	logger.Infof("Health check: http://localhost:%s/health", port)

	if err := server.Start(port); err != nil {
		logger.Fatalf("Failed to start server: %v", err)
	}
}
