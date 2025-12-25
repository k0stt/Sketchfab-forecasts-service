#!/usr/bin/env python3
"""
Обучение модели машинного обучения для прогнозирования популярности 3D-моделей
"""

import json
import pandas as pd
import numpy as np
from sklearn.model_selection import train_test_split, cross_val_score
from sklearn.ensemble import RandomForestRegressor, GradientBoostingRegressor
from sklearn.linear_model import LinearRegression
from sklearn.preprocessing import StandardScaler
from sklearn.metrics import mean_squared_error, r2_score, mean_absolute_error
import joblib
import matplotlib.pyplot as plt
import seaborn as sns

def load_data(filename='data/preprocessed_data.json'):
    """Загрузка обработанных данных"""
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    return pd.DataFrame(data)

def prepare_features(df):
    """Подготовка признаков для обучения"""
    # Выбираем признаки
    feature_columns = [
        'category_count',
        'tag_count',
        'description_length',
        'face_count',
        'vertex_count',
        'animation_count',
        'is_downloadable',
        'is_premium_author',
        'author_followers',
        'days_since_published'
    ]
    
    X = df[feature_columns].copy()
    y = df['popularity_score'].values
    
    # Преобразуем булевы значения в числовые
    X['is_downloadable'] = X['is_downloadable'].astype(int)
    X['is_premium_author'] = X['is_premium_author'].astype(int)
    
    # Обработка пропущенных значений
    X = X.fillna(0)
    
    return X, y, feature_columns

def train_models(X_train, y_train):
    """Обучение нескольких моделей"""
    models = {
        'Linear Regression': LinearRegression(),
        'Random Forest': RandomForestRegressor(
            n_estimators=100,
            max_depth=15,
            min_samples_split=5,
            random_state=42,
            n_jobs=-1
        ),
        'Gradient Boosting': GradientBoostingRegressor(
            n_estimators=100,
            max_depth=5,
            learning_rate=0.1,
            random_state=42
        )
    }
    
    trained_models = {}
    scores = {}
    
    print("\n=== Обучение моделей ===")
    for name, model in models.items():
        print(f"\nОбучение {name}...")
        model.fit(X_train, y_train)
        trained_models[name] = model
        
        # Cross-validation
        cv_scores = cross_val_score(model, X_train, y_train, cv=5, 
                                     scoring='neg_mean_squared_error')
        scores[name] = {
            'cv_mse': -cv_scores.mean(),
            'cv_std': cv_scores.std()
        }
        print(f"  CV MSE: {scores[name]['cv_mse']:.4f} (+/- {scores[name]['cv_std']:.4f})")
    
    return trained_models, scores

def evaluate_models(models, X_test, y_test):
    """Оценка моделей на тестовой выборке"""
    print("\n=== Оценка моделей на тестовой выборке ===")
    
    results = {}
    for name, model in models.items():
        y_pred = model.predict(X_test)
        
        mse = mean_squared_error(y_test, y_pred)
        rmse = np.sqrt(mse)
        mae = mean_absolute_error(y_test, y_pred)
        r2 = r2_score(y_test, y_pred)
        
        results[name] = {
            'mse': mse,
            'rmse': rmse,
            'mae': mae,
            'r2': r2,
            'predictions': y_pred
        }
        
        print(f"\n{name}:")
        print(f"  MSE: {mse:.4f}")
        print(f"  RMSE: {rmse:.4f}")
        print(f"  MAE: {mae:.4f}")
        print(f"  R²: {r2:.4f}")
    
    return results

def plot_feature_importance(model, feature_names):
    """Визуализация важности признаков"""
    if hasattr(model, 'feature_importances_'):
        importances = model.feature_importances_
        indices = np.argsort(importances)[::-1]
        
        plt.figure(figsize=(10, 6))
        plt.bar(range(len(importances)), importances[indices])
        plt.xticks(range(len(importances)), [feature_names[i] for i in indices], rotation=45, ha='right')
        plt.xlabel('Признаки')
        plt.ylabel('Важность')
        plt.title('Важность признаков (Random Forest)')
        plt.tight_layout()
        plt.savefig('data/feature_importance.png', dpi=300)
        print("\nГрафик важности признаков сохранен: data/feature_importance.png")
        
        print("\nТоп-5 важных признаков:")
        for i in range(min(5, len(indices))):
            print(f"  {feature_names[indices[i]]}: {importances[indices[i]]:.4f}")

def plot_predictions(y_test, predictions, model_name):
    """Визуализация предсказаний vs реальные значения"""
    plt.figure(figsize=(10, 6))
    plt.scatter(y_test, predictions, alpha=0.5)
    plt.plot([y_test.min(), y_test.max()], [y_test.min(), y_test.max()], 'r--', lw=2)
    plt.xlabel('Реальные значения')
    plt.ylabel('Предсказанные значения')
    plt.title(f'Предсказания vs Реальность ({model_name})')
    plt.tight_layout()
    plt.savefig(f'data/predictions_{model_name.replace(" ", "_").lower()}.png', dpi=300)
    print(f"График предсказаний сохранен: data/predictions_{model_name.replace(' ', '_').lower()}.png")

def save_best_model(models, results, feature_columns, data_size):
    """Сохранение лучшей модели"""
    # Находим модель с наименьшим RMSE
    best_model_name = min(results.keys(), key=lambda x: results[x]['rmse'])
    best_model = models[best_model_name]
    
    print(f"\n=== Лучшая модель: {best_model_name} ===")
    print(f"RMSE: {results[best_model_name]['rmse']:.4f}")
    print(f"R²: {results[best_model_name]['r2']:.4f}")
    
    # Сохраняем модель
    model_data = {
        'model': best_model,
        'model_name': best_model_name,
        'feature_columns': feature_columns,
        'metrics': results[best_model_name]
    }
    
    joblib.dump(model_data, 'models/popularity_model.pkl')
    print("\nМодель сохранена: models/popularity_model.pkl")
    
    # Сохраняем метрики для веб-интерфейса
    from datetime import datetime
    metrics_json = {
        'trained': True,
        'training_date': datetime.now().strftime('%Y-%m-%d %H:%M:%S'),
        'rmse': results[best_model_name]['rmse'],
        'mae': results[best_model_name]['mae'],
        'r2_score': results[best_model_name]['r2'],
        'training_samples': data_size,
        'model_type': best_model_name,
        'features': feature_columns
    }
    
    with open('models/model_metrics.json', 'w') as f:
        json.dump(metrics_json, f, indent=2)
    
    print("Метрики сохранены: models/model_metrics.json")
    
    return best_model, best_model_name

def main():
    """Основная функция"""
    print("Запуск обучения модели машинного обучения...")
    
    # Загрузка данных
    df = load_data()
    print(f"Загружено {len(df)} записей")
    
    # Подготовка признаков
    X, y, feature_columns = prepare_features(df)
    print(f"\nПризнаков: {len(feature_columns)}")
    print(f"Целевая переменная: popularity_score")
    
    # Разделение на train/test
    X_train, X_test, y_train, y_test = train_test_split(
        X, y, test_size=0.2, random_state=42
    )
    print(f"\nТренировочная выборка: {len(X_train)}")
    print(f"Тестовая выборка: {len(X_test)}")
    
    # Нормализация признаков
    scaler = StandardScaler()
    X_train_scaled = scaler.fit_transform(X_train)
    X_test_scaled = scaler.transform(X_test)
    
    # Обучение моделей
    models, cv_scores = train_models(X_train_scaled, y_train)
    
    # Оценка моделей
    results = evaluate_models(models, X_test_scaled, y_test)
    
    # Визуализация важности признаков
    plot_feature_importance(models['Random Forest'], feature_columns)
    
    # Визуализация предсказаний лучшей модели
    best_model_name = min(results.keys(), key=lambda x: results[x]['rmse'])
    plot_predictions(y_test, results[best_model_name]['predictions'], best_model_name)
    
    # Сохранение лучшей модели и scaler
    data_size = len(X_train) + len(X_test)
    save_best_model(models, results, feature_columns, data_size)
    joblib.dump(scaler, 'models/scaler.pkl')
    print("Scaler сохранен: models/scaler.pkl")
    
    print("\n=== Обучение завершено! ===")

if __name__ == "__main__":
    main()
