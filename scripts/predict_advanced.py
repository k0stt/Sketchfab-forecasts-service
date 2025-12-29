#!/usr/bin/env python3
"""
Расширенный скрипт прогнозирования с поддержкой:
- Текстовых признаков (теги, описание)
- Оценки качества модели
"""

import sys
import json
import joblib
import numpy as np
import pandas as pd
import os

# Добавляем путь к модулю quality_rating
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
from quality_rating import QualityRater

def load_models():
    """Загрузка всех доступных моделей"""
    models = {}
    
    # Стандартная модель
    try:
        standard_model = joblib.load('models/popularity_model.pkl')
        standard_scaler = joblib.load('models/scaler.pkl')
        models['standard'] = {
            'model_data': standard_model,
            'scaler': standard_scaler
        }
    except FileNotFoundError:
        models['standard'] = None
    
    # Расширенная модель с текстом
    try:
        advanced_model = joblib.load('models/popularity_model_advanced.pkl')
        models['advanced'] = advanced_model
    except FileNotFoundError:
        models['advanced'] = None
    
    return models

def predict_popularity_standard(input_data, model_data, scaler):
    """Стандартное прогнозирование без текста"""
    feature_columns = model_data['feature_columns']
    
    # Подготовка признаков
    features = {}
    features['category_count'] = input_data.get('category_count', 0)
    features['tag_count'] = input_data.get('tag_count', 0)
    features['description_length'] = input_data.get('description_length', 0)
    features['face_count'] = input_data.get('face_count', 0)
    features['vertex_count'] = input_data.get('vertex_count', 0)
    features['animation_count'] = input_data.get('animation_count', 0)
    features['is_downloadable'] = 1 if input_data.get('is_downloadable', False) else 0
    features['is_premium_author'] = 1 if input_data.get('is_premium_author', False) else 0
    features['author_followers'] = input_data.get('author_followers', 0)
    features['days_since_published'] = input_data.get('days_since_published', 0)
    
    # Создаем DataFrame
    X = pd.DataFrame([features], columns=feature_columns)
    
    # Нормализация
    X_scaled = scaler.transform(X)
    
    # Предсказание
    model = model_data['model']
    prediction = model.predict(X_scaled)[0]
    
    return prediction

def preprocess_text(text):
    """Предобработка текста"""
    import re
    if not text:
        return ""
    text = text.lower()
    text = re.sub(r'[^a-zA-Z0-9\s]', ' ', text)
    text = re.sub(r'\s+', ' ', text).strip()
    return text

def predict_popularity_advanced(input_data, model_data):
    """Расширенное прогнозирование с текстом"""
    # Извлекаем компоненты модели
    model = model_data['model']
    tfidf = model_data['tfidf']
    scaler = model_data['scaler']
    numeric_features = model_data['numeric_features']
    
    # Численные признаки
    features = {}
    features['category_count'] = input_data.get('category_count', 0)
    features['tag_count'] = input_data.get('tag_count', 0)
    features['description_length'] = input_data.get('description_length', 0)
    features['face_count'] = input_data.get('face_count', 0)
    features['vertex_count'] = input_data.get('vertex_count', 0)
    features['animation_count'] = input_data.get('animation_count', 0)
    features['is_downloadable'] = 1 if input_data.get('is_downloadable', False) else 0
    features['is_premium_author'] = 1 if input_data.get('is_premium_author', False) else 0
    features['author_followers'] = input_data.get('author_followers', 0)
    
    X_numeric = pd.DataFrame([features], columns=numeric_features)
    X_numeric_scaled = scaler.transform(X_numeric)
    
    # Текстовые признаки
    tags = input_data.get('tags', [])
    description = input_data.get('description', '')
    categories = input_data.get('categories', [])
    
    # Объединяем текст
    tags_text = ' '.join([str(tag) for tag in tags]) if isinstance(tags, list) else ''
    categories_text = ' '.join([str(cat) for cat in categories]) if isinstance(categories, list) else ''
    combined_text = preprocess_text(f"{tags_text} {description} {categories_text}")
    
    # Векторизация текста
    X_text_vec = tfidf.transform([combined_text]).toarray()
    
    # Объединяем признаки
    X_combined = np.hstack([X_numeric_scaled, X_text_vec])
    
    # Предсказание
    prediction = model.predict(X_combined)[0]
    
    return prediction

def calculate_quality(input_data):
    """Расчет рейтинга качества модели"""
    rater = QualityRater()
    
    # Подготовка данных для рейтера
    quality_data = {
        'tags': input_data.get('tags', []),
        'description': input_data.get('description', ''),
        'face_count': input_data.get('face_count', 0),
        'category': input_data.get('category', 'generic'),
        'account_type': input_data.get('account_type', 'basic'),
        'author_followers': input_data.get('author_followers', 0),
        'is_downloadable': input_data.get('is_downloadable', False),
        'has_textures': input_data.get('has_textures', False),
        'has_pbr': input_data.get('has_pbr', False),
        'is_rigged': input_data.get('is_rigged', False),
        'is_animated': input_data.get('is_animated', False)
    }
    
    return rater.calculate_quality_score(quality_data)

def categorize_score(score):
    """Категоризация оценки популярности"""
    if score >= 8:
        return "high"
    elif score >= 5:
        return "medium"
    else:
        return "low"

def main():
    """Основная функция"""
    # Читаем входные данные из stdin
    input_data = json.loads(sys.stdin.read())
    
    # Загружаем модели
    models = load_models()
    
    result = {}
    
    # Пытаемся использовать расширенную модель
    if models['advanced'] and ('tags' in input_data or 'description' in input_data):
        try:
            popularity_score = predict_popularity_advanced(input_data, models['advanced'])
            result['model_used'] = 'advanced'
        except Exception as e:
            print(f"Warning: Advanced model failed: {e}", file=sys.stderr)
            # Fallback на стандартную модель
            if models['standard']:
                popularity_score = predict_popularity_standard(
                    input_data, 
                    models['standard']['model_data'],
                    models['standard']['scaler']
                )
                result['model_used'] = 'standard'
            else:
                result['error'] = 'No models available'
                print(json.dumps(result))
                return
    # Используем стандартную модель
    elif models['standard']:
        popularity_score = predict_popularity_standard(
            input_data, 
            models['standard']['model_data'],
            models['standard']['scaler']
        )
        result['model_used'] = 'standard'
    else:
        result['error'] = 'No models available'
        print(json.dumps(result))
        return
    
    # Прогноз популярности
    result['popularity_score'] = float(popularity_score)
    result['popularity_category'] = categorize_score(popularity_score)
    result['confidence'] = 0.85  # Можно улучшить с помощью prediction intervals
    
    # Рейтинг качества
    try:
        quality_result = calculate_quality(input_data)
        result['quality_rating'] = {
            'score': quality_result['total_score'],
            'grade': quality_result['grade'],
            'details': quality_result['scores'],
            'recommendations': quality_result['recommendations']
        }
    except Exception as e:
        print(f"Warning: Quality rating failed: {e}", file=sys.stderr)
        result['quality_rating'] = None
    
    # Выводим результат
    print(json.dumps(result, indent=2))

if __name__ == "__main__":
    main()
