#!/usr/bin/env python3
"""
Скрипт для прогнозирования популярности модели
Используется обученная ML модель
"""

import sys
import json
import joblib
import numpy as np
import pandas as pd

def load_model():
    """Загрузка обученной модели"""
    try:
        model_data = joblib.load('models/popularity_model.pkl')
        scaler = joblib.load('models/scaler.pkl')
        return model_data, scaler
    except FileNotFoundError:
        # Если модель не найдена, возвращаем None
        return None, None

def predict(input_data, model_data, scaler):
    """Выполнение предсказания"""
    # Подготовка признаков в правильном порядке
    feature_columns = model_data['feature_columns']
    
    # Создаем DataFrame с одной строкой
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
    
    # Категоризация
    if prediction < 2.0:
        category = "low"
        confidence = 0.7
    elif prediction < 4.0:
        category = "medium"
        confidence = 0.8
    else:
        category = "high"
        confidence = 0.85
    
    return {
        'popularity_score': float(prediction),
        'category': category,
        'confidence': confidence
    }

def main():
    """Основная функция"""
    if len(sys.argv) < 2:
        print(json.dumps({'error': 'No input data provided'}))
        sys.exit(1)
    
    try:
        # Парсим входные данные
        input_json = sys.argv[1]
        input_data = json.loads(input_json)
        
        # Загружаем модель
        model_data, scaler = load_model()
        
        if model_data is None or scaler is None:
            # Если модель не найдена, используем простую эвристику
            result = simple_predict(input_data)
        else:
            # Используем обученную модель
            result = predict(input_data, model_data, scaler)
        
        # Выводим результат в JSON
        print(json.dumps(result))
        
    except Exception as e:
        print(json.dumps({'error': str(e)}))
        sys.exit(1)

def simple_predict(input_data):
    """Простое предсказание без ML модели"""
    score = 0.0
    
    # Простая эвристика
    score += input_data.get('tag_count', 0) * 0.15
    score += input_data.get('category_count', 0) * 0.2
    score += input_data.get('description_length', 0) * 0.001
    score += input_data.get('author_followers', 0) * 0.0005
    
    # ВАЖНО: Учитываем полигоны!
    face_count = input_data.get('face_count', 0)
    vertex_count = input_data.get('vertex_count', 0)
    
    # Оптимальный диапазон полигонов: 5000-50000
    if 5000 <= face_count <= 50000:
        score += 1.0
    elif face_count > 0:
        if face_count < 5000:
            score += 0.5 * (face_count / 5000)
        else:
            score += 0.5 * max(0, 1 - (face_count - 50000) / 100000)
    
    # Учитываем вершины
    if 2500 <= vertex_count <= 25000:
        score += 0.5
    elif vertex_count > 0:
        score += 0.2
    
    if input_data.get('is_downloadable', False):
        score += 0.5
    
    if input_data.get('is_premium_author', False):
        score += 0.3
    
    animation_count = input_data.get('animation_count', 0)
    if animation_count > 0:
        score += animation_count * 0.2
    
    # Нормализация (шкала 0-10)
    score = min(max(score, 0), 10.0)
    
    if score < 2.0:
        category = "low"
    elif score < 4.0:
        category = "medium"
    else:
        category = "high"
    
    return {
        'popularity_score': score,
        'category': category,
        'confidence': 0.6
    }

if __name__ == "__main__":
    main()
