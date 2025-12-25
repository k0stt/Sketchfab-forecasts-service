#!/usr/bin/env python3
"""
Разведочный анализ данных (EDA) для 3D-моделей Sketchfab
Визуализация и статистический анализ
"""

import json
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
import seaborn as sns
from collections import Counter
from datetime import datetime
import os

# Настройка стиля графиков
sns.set_style("whitegrid")
plt.rcParams['figure.figsize'] = (12, 6)

def load_data(filename='data/preprocessed_data.json'):
    """Загрузка обработанных данных"""
    with open(filename, 'r', encoding='utf-8') as f:
        data = json.load(f)
    return pd.DataFrame(data)

def analyze_popularity_distribution(df):
    """Анализ распределения популярности"""
    print("\n=== Распределение популярности ===")
    print(df['popularity_score'].describe())
    
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))
    
    # Гистограмма
    axes[0].hist(df['popularity_score'], bins=50, edgecolor='black', alpha=0.7)
    axes[0].set_xlabel('Показатель популярности')
    axes[0].set_ylabel('Частота')
    axes[0].set_title('Распределение показателя популярности')
    
    # Box plot
    axes[1].boxplot(df['popularity_score'])
    axes[1].set_ylabel('Показатель популярности')
    axes[1].set_title('Box Plot популярности')
    
    plt.tight_layout()
    plt.savefig('data/eda_popularity_distribution.png', dpi=300)
    print("График сохранен: data/eda_popularity_distribution.png")

def analyze_features_correlation(df):
    """Анализ корреляции признаков"""
    print("\n=== Корреляция признаков ===")
    
    # Выбираем числовые признаки
    numeric_cols = [
        'category_count', 'tag_count', 'description_length',
        'face_count', 'vertex_count', 'animation_count',
        'author_followers', 'days_since_published', 'popularity_score'
    ]
    
    correlation_matrix = df[numeric_cols].corr()
    print(correlation_matrix['popularity_score'].sort_values(ascending=False))
    
    # Тепловая карта корреляций
    plt.figure(figsize=(10, 8))
    sns.heatmap(correlation_matrix, annot=True, fmt='.2f', cmap='coolwarm', center=0)
    plt.title('Матрица корреляции признаков')
    plt.tight_layout()
    plt.savefig('data/eda_correlation_matrix.png', dpi=300)
    print("График сохранен: data/eda_correlation_matrix.png")

def analyze_categorical_features(df):
    """Анализ влияния категориальных признаков"""
    print("\n=== Анализ категориальных признаков ===")
    
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))
    
    # Влияние downloadable
    downloadable_stats = df.groupby('is_downloadable')['popularity_score'].mean()
    axes[0].bar(['Не доступна', 'Доступна'], downloadable_stats.values)
    axes[0].set_ylabel('Средняя популярность')
    axes[0].set_title('Влияние доступности скачивания')
    
    # Влияние премиум аккаунта автора
    premium_stats = df.groupby('is_premium_author')['popularity_score'].mean()
    axes[1].bar(['Обычный', 'Премиум'], premium_stats.values)
    axes[1].set_ylabel('Средняя популярность')
    axes[1].set_title('Влияние премиум аккаунта автора')
    
    plt.tight_layout()
    plt.savefig('data/eda_categorical_features.png', dpi=300)
    print("График сохранен: data/eda_categorical_features.png")

def analyze_tags_and_categories(df):
    """Анализ тегов и категорий"""
    print("\n=== Анализ количества тегов и категорий ===")
    
    fig, axes = plt.subplots(1, 2, figsize=(14, 5))
    
    # Распределение количества тегов
    axes[0].hist(df['tag_count'], bins=30, edgecolor='black', alpha=0.7)
    axes[0].set_xlabel('Количество тегов')
    axes[0].set_ylabel('Частота')
    axes[0].set_title('Распределение количества тегов')
    
    # Распределение количества категорий
    axes[1].hist(df['category_count'], bins=20, edgecolor='black', alpha=0.7)
    axes[1].set_xlabel('Количество категорий')
    axes[1].set_ylabel('Частота')
    axes[1].set_title('Распределение количества категорий')
    
    plt.tight_layout()
    plt.savefig('data/eda_tags_categories.png', dpi=300)
    print("График сохранен: data/eda_tags_categories.png")
    
    print(f"\nСреднее количество тегов: {df['tag_count'].mean():.2f}")
    print(f"Среднее количество категорий: {df['category_count'].mean():.2f}")

def analyze_complexity(df):
    """Анализ сложности моделей (полигоны, вершины)"""
    print("\n=== Анализ сложности моделей ===")
    
    fig, axes = plt.subplots(2, 2, figsize=(14, 10))
    
    # Распределение количества полигонов
    axes[0, 0].hist(df[df['face_count'] > 0]['face_count'], bins=50, edgecolor='black', alpha=0.7)
    axes[0, 0].set_xlabel('Количество полигонов')
    axes[0, 0].set_ylabel('Частота')
    axes[0, 0].set_title('Распределение количества полигонов')
    axes[0, 0].set_xscale('log')
    
    # Распределение количества вершин
    axes[0, 1].hist(df[df['vertex_count'] > 0]['vertex_count'], bins=50, edgecolor='black', alpha=0.7)
    axes[0, 1].set_xlabel('Количество вершин')
    axes[0, 1].set_ylabel('Частота')
    axes[0, 1].set_title('Распределение количества вершин')
    axes[0, 1].set_xscale('log')
    
    # Scatter: полигоны vs популярность
    axes[1, 0].scatter(df['face_count'], df['popularity_score'], alpha=0.5)
    axes[1, 0].set_xlabel('Количество полигонов')
    axes[1, 0].set_ylabel('Популярность')
    axes[1, 0].set_title('Полигоны vs Популярность')
    axes[1, 0].set_xscale('log')
    
    # Scatter: подписчики автора vs популярность
    axes[1, 1].scatter(df['author_followers'], df['popularity_score'], alpha=0.5)
    axes[1, 1].set_xlabel('Подписчики автора')
    axes[1, 1].set_ylabel('Популярность')
    axes[1, 1].set_title('Подписчики vs Популярность')
    axes[1, 1].set_xscale('log')
    
    plt.tight_layout()
    plt.savefig('data/eda_complexity_analysis.png', dpi=300)
    print("График сохранен: data/eda_complexity_analysis.png")

def analyze_temporal_trends(df):
    """Анализ временных закономерностей"""
    print("\n=== Временной анализ ===")
    
    # Группируем по возрасту модели
    df['age_group'] = pd.cut(df['days_since_published'], 
                              bins=[0, 30, 90, 180, 365, float('inf')],
                              labels=['<1 мес', '1-3 мес', '3-6 мес', '6-12 мес', '>1 года'])
    
    age_popularity = df.groupby('age_group')['popularity_score'].mean()
    
    plt.figure(figsize=(10, 6))
    age_popularity.plot(kind='bar', color='skyblue', edgecolor='black')
    plt.xlabel('Возраст модели')
    plt.ylabel('Средняя популярность')
    plt.title('Зависимость популярности от возраста модели')
    plt.xticks(rotation=45)
    plt.tight_layout()
    plt.savefig('data/eda_temporal_trends.png', dpi=300)
    print("График сохранен: data/eda_temporal_trends.png")

def generate_summary_report(df):
    """Генерация итогового отчета"""
    print("\n" + "="*60)
    print("ИТОГОВЫЙ ОТЧЕТ - РАЗВЕДОЧНЫЙ АНАЛИЗ ДАННЫХ")
    print("="*60)
    
    print(f"\nОбщая информация:")
    print(f"  Всего моделей: {len(df)}")
    print(f"  Признаков: {len(df.columns)}")
    
    print(f"\nПопулярность:")
    print(f"  Средняя: {df['popularity_score'].mean():.4f}")
    print(f"  Медиана: {df['popularity_score'].median():.4f}")
    print(f"  Стд. откл.: {df['popularity_score'].std():.4f}")
    
    print(f"\nКлючевые находки:")
    # Exclude non-numeric columns before computing correlation
    numeric_df = df.select_dtypes(include=[np.number])
    corr = numeric_df.corr()['popularity_score'].sort_values(ascending=False)
    print(f"  Топ-3 коррелирующих признака:")
    for i, (feature, value) in enumerate(list(corr.items())[1:4], 1):
        print(f"    {i}. {feature}: {value:.3f}")
    
    downloadable_impact = df.groupby('is_downloadable')['popularity_score'].mean()
    premium_impact = df.groupby('is_premium_author')['popularity_score'].mean()
    
    print(f"\n  Влияние доступности скачивания: " + 
          f"{((downloadable_impact[True] / downloadable_impact[False] - 1) * 100):.1f}%")
    print(f"  Влияние премиум аккаунта: " + 
          f"{((premium_impact[True] / premium_impact[False] - 1) * 100):.1f}%")
    
    print("\n" + "="*60)

def main():
    """Основная функция"""
    print("Запуск разведочного анализа данных (EDA)...")
    
    # Создаем директорию для графиков если не существует
    os.makedirs('data', exist_ok=True)
    
    # Загрузка данных
    df = load_data()
    print(f"Загружено {len(df)} записей")
    
    # Проводим анализы
    analyze_popularity_distribution(df)
    analyze_features_correlation(df)
    analyze_categorical_features(df)
    analyze_tags_and_categories(df)
    analyze_complexity(df)
    analyze_temporal_trends(df)
    
    # Генерируем итоговый отчет
    generate_summary_report(df)
    
    print("\nРазведочный анализ завершен!")
    print("Все графики сохранены в директории 'data/'")

if __name__ == "__main__":
    main()
