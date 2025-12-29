#!/usr/bin/env python3
"""
Расширенное обучение модели с поддержкой текстовых признаков (теги, описание)
"""

import json
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.ensemble import RandomForestRegressor, GradientBoostingRegressor
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import StandardScaler
from sklearn.feature_extraction.text import TfidfVectorizer
from sklearn.metrics import mean_squared_error, r2_score, mean_absolute_error
from sklearn.compose import ColumnTransformer
from sklearn.pipeline import Pipeline
import joblib
import re

def load_raw_data(filename='data/raw_models.json'):
    """Загрузка сырых данных с тегами и описанием"""
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    return data

def preprocess_text(text):
    """Предобработка текста"""
    if not text:
        return ""
    # Приводим к нижнему регистру
    text = text.lower()
    # Удаляем специальные символы
    text = re.sub(r'[^a-zA-Z0-9\s]', ' ', text)
    # Удаляем множественные пробелы
    text = re.sub(r'\s+', ' ', text).strip()
    return text

def prepare_advanced_features(raw_data):
    """Подготовка признаков с текстовыми данными"""
    df_list = []
    
    for model in raw_data:
        # Извлекаем теги
        tags = model.get('tags', [])
        if isinstance(tags, list):
            tags_text = ' '.join([str(tag) for tag in tags])
        else:
            tags_text = ""
        
        # Описание
        description = model.get('description', '')
        
        # Категории
        categories = model.get('categories', [])
        if isinstance(categories, list):
            categories_text = ' '.join([str(cat) for cat in categories])
        else:
            categories_text = ""
        
        # Объединяем все текстовые данные
        combined_text = preprocess_text(f"{tags_text} {description} {categories_text}")
        
        # Численные признаки
        face_count = model.get('faceCount', 0)
        vertex_count = model.get('vertexCount', 0)
        animation_count = model.get('animationCount', 0)
        is_downloadable = 1 if model.get('isDownloadable', False) else 0
        
        # Автор
        user = model.get('user', {})
        is_premium = 1 if user.get('account', 'basic') in ['pro', 'premium'] else 0
        followers = user.get('followerCount', 0)
        
        # Целевая переменная (популярность)
        views = model.get('viewCount', 0)
        likes = model.get('likeCount', 0)
        downloads = model.get('downloadCount', 0)
        
        # Вычисляем популярность
        popularity = (np.log1p(views) * 0.25 + 
                     np.log1p(likes) * 0.35 + 
                     np.log1p(downloads) * 0.25 +
                     calculate_polygon_score(face_count) * 0.15)
        
        df_list.append({
            'combined_text': combined_text,
            'tags_text': preprocess_text(tags_text),
            'description_text': preprocess_text(description),
            'category_count': len(categories) if isinstance(categories, list) else 0,
            'tag_count': len(tags) if isinstance(tags, list) else 0,
            'description_length': len(description),
            'face_count': face_count,
            'vertex_count': vertex_count,
            'animation_count': animation_count,
            'is_downloadable': is_downloadable,
            'is_premium_author': is_premium,
            'author_followers': followers,
            'popularity_score': popularity
        })
    
    return pd.DataFrame(df_list)

def calculate_polygon_score(face_count):
    """Оценка качества по полигонам (копия из Go кода)"""
    if face_count <= 0:
        return 0
    
    min_optimal = 5000.0
    max_optimal = 50000.0
    penalty_rate = 0.5
    
    faces = float(face_count)
    
    if faces >= min_optimal and faces <= max_optimal:
        mid = (min_optimal + max_optimal) / 2
        distance = abs(faces - mid)
        max_distance = (max_optimal - min_optimal) / 2
        return 10.0 * (1.0 - (distance / max_distance) * 0.2)
    
    if faces < min_optimal:
        ratio = faces / min_optimal
        return np.log1p(faces) * ratio * penalty_rate
    
    excess = (faces - max_optimal) / max_optimal
    penalty = 1.0 / (1.0 + excess * penalty_rate)
    return 10.0 * penalty

def create_advanced_pipeline():
    """Создание pipeline с обработкой текста"""
    # Численные признаки
    numeric_features = [
        'category_count', 'tag_count', 'description_length',
        'face_count', 'vertex_count', 'animation_count',
        'is_downloadable', 'is_premium_author', 'author_followers'
    ]
    
    # Создаем TF-IDF векторизатор для текста
    # max_features ограничивает количество признаков
    tfidf = TfidfVectorizer(
        max_features=100,
        min_df=2,
        max_df=0.8,
        ngram_range=(1, 2)
    )
    
    return numeric_features, tfidf

def train_advanced_model(X_train_numeric, X_train_text, y_train):
    """Обучение модели с текстовыми признаками"""
    # Векторизация текста
    tfidf = TfidfVectorizer(
        max_features=100,
        min_df=2,
        max_df=0.8,
        ngram_range=(1, 2)
    )
    
    X_train_text_vec = tfidf.fit_transform(X_train_text).toarray()
    
    # Объединяем численные и текстовые признаки
    X_train_combined = np.hstack([X_train_numeric, X_train_text_vec])
    
    # Обучаем модель
    model = GradientBoostingRegressor(
        n_estimators=200,
        max_depth=7,
        learning_rate=0.05,
        random_state=42,
        subsample=0.8
    )
    
    model.fit(X_train_combined, y_train)
    
    return model, tfidf, X_train_text_vec.shape[1]

def evaluate_advanced_model(model, tfidf, X_test_numeric, X_test_text, y_test):
    """Оценка модели с текстовыми признаками"""
    X_test_text_vec = tfidf.transform(X_test_text).toarray()
    X_test_combined = np.hstack([X_test_numeric, X_test_text_vec])
    
    y_pred = model.predict(X_test_combined)
    
    mse = mean_squared_error(y_test, y_pred)
    rmse = np.sqrt(mse)
    mae = mean_absolute_error(y_test, y_pred)
    r2 = r2_score(y_test, y_pred)
    
    return {
        'mse': mse,
        'rmse': rmse,
        'mae': mae,
        'r2': r2,
        'predictions': y_pred
    }

def save_advanced_model(model, tfidf, scaler, numeric_features, text_features_count, metrics):
    """Сохранение расширенной модели"""
    model_data = {
        'model': model,
        'tfidf': tfidf,
        'scaler': scaler,
        'model_name': 'Advanced Gradient Boosting with Text Features',
        'numeric_features': numeric_features,
        'text_features_count': text_features_count,
        'metrics': metrics
    }
    
    joblib.dump(model_data, 'models/popularity_model_advanced.pkl')
    print("\nРасширенная модель сохранена: models/popularity_model_advanced.pkl")
    
    # Метрики
    from datetime import datetime
    metrics_json = {
        'trained': True,
        'training_date': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        'rmse': metrics['rmse'],
        'mae': metrics['mae'],
        'r2_score': metrics['r2'],
        'model_type': 'Advanced Gradient Boosting with Text Features',
        'features': {
            'numeric': numeric_features,
            'text_features_count': text_features_count,
            'total': len(numeric_features) + text_features_count
        }
    }
    
    with open('models/model_metrics_advanced.json', 'w') as f:
        json.dump(metrics_json, f, indent=2)
    
    print("Метрики сохранены: models/model_metrics_advanced.json")

def main():
    """Основная функция"""
    print("=" * 60)
    print("Обучение расширенной модели с текстовыми признаками")
    print("=" * 60)
    
    # Загрузка данных
    print("\nЗагрузка сырых данных...")
    raw_data = load_raw_data()
    
    # Применяем ограничение если задано
    if 'limit' in globals() and limit:
        raw_data = raw_data[:limit]
    
    print(f"Загружено {len(raw_data)} моделей")
    
    # Подготовка признаков
    print("\nПодготовка признаков (включая текст)...")
    df = prepare_advanced_features(raw_data)
    print(f"Подготовлено {len(df)} записей")
    
    # Разделение на признаки и целевую переменную
    numeric_features = [
        'category_count', 'tag_count', 'description_length',
        'face_count', 'vertex_count', 'animation_count',
        'is_downloadable', 'is_premium_author', 'author_followers'
    ]
    
    X_numeric = df[numeric_features]
    X_text = df['combined_text']
    y = df['popularity_score'].values
    
    # Разделение на train/test
    X_train_num, X_test_num, X_train_text, X_test_text, y_train, y_test = train_test_split(
        X_numeric, X_text, y, test_size=0.2, random_state=42
    )
    
    print(f"\nТренировочная выборка: {len(X_train_num)}")
    print(f"Тестовая выборка: {len(X_test_num)}")
    
    # Нормализация численных признаков
    scaler = StandardScaler()
    X_train_num_scaled = scaler.fit_transform(X_train_num)
    X_test_num_scaled = scaler.transform(X_test_num)
    
    # Обучение модели
    print("\nОбучение модели с текстовыми признаками...")
    model, tfidf, text_features_count = train_advanced_model(
        X_train_num_scaled, X_train_text, y_train
    )
    
    print(f"Численных признаков: {len(numeric_features)}")
    print(f"Текстовых признаков (TF-IDF): {text_features_count}")
    print(f"Всего признаков: {len(numeric_features) + text_features_count}")
    
    # Оценка модели
    print("\n" + "=" * 60)
    print("Оценка модели на тестовой выборке")
    print("=" * 60)
    results = evaluate_advanced_model(
        model, tfidf, X_test_num_scaled, X_test_text, y_test
    )
    
    print(f"\nMSE: {results['mse']:.4f}")
    print(f"RMSE: {results['rmse']:.4f}")
    print(f"MAE: {results['mae']:.4f}")
    print(f"R²: {results['r2']:.4f}")
    
    # Сохранение модели
    save_advanced_model(
        model, tfidf, scaler, numeric_features, 
        text_features_count, results
    )
    
    # Пример важных слов из TF-IDF
    print("\n" + "=" * 60)
    print("Топ-20 важных слов/фраз для популярности:")
    print("=" * 60)
    feature_names = tfidf.get_feature_names_out()
    # Получаем важность признаков
    if hasattr(model, 'feature_importances_'):
        importances = model.feature_importances_
        # Берем только текстовые признаки
        text_importances = importances[len(numeric_features):]
        # Сортируем
        indices = np.argsort(text_importances)[::-1][:20]
        for i, idx in enumerate(indices, 1):
            if idx < len(feature_names):
                print(f"{i}. {feature_names[idx]}: {text_importances[idx]:.4f}")
    
    print("\n" + "=" * 60)
    print("Обучение завершено!")
    print("=" * 60)

if __name__ == "__main__":
    import sys
    # Поддержка параметра для ограничения количества данных
    limit = None
    if len(sys.argv) > 1:
        try:
            limit = int(sys.argv[1])
            print(f"Ограничение данных: {limit} моделей")
        except ValueError:
            pass
    
    main()
