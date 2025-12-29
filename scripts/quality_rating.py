#!/usr/bin/env python3
"""
Система оценки качества модели (0-100%)
Оценивает насколько хорошо модель подготовлена для Sketchfab
"""

import re
import numpy as np

class QualityRater:
    """Оценивает качество модели для Sketchfab"""
    
    def __init__(self):
        # Оптимальные диапазоны для разных целей
        self.polygon_ranges = {
            'game_mobile': (1000, 10000),
            'game_desktop': (5000, 50000),
            'architecture': (10000, 100000),
            'showcase': (50000, 500000),
            'generic': (5000, 50000)
        }
        
        # Популярные теги для разных категорий
        self.recommended_tags = {
            'game': ['game', 'lowpoly', 'pbr', 'realtime', 'unity', 'unreal'],
            'architecture': ['architecture', 'building', 'interior', 'exterior', 'archviz'],
            'character': ['character', 'rigged', 'animated', 'humanoid', 'creature'],
            'vehicle': ['vehicle', 'car', 'transportation', 'pbr', 'lowpoly'],
            'nature': ['nature', 'environment', 'landscape', 'organic', 'plants'],
            'prop': ['prop', 'asset', 'object', '3d', 'model']
        }
        
    def calculate_quality_score(self, model_data):
        """
        Вычисляет общий рейтинг качества модели (0-100%)
        
        Args:
            model_data: dict с параметрами модели
                - tags: список тегов
                - description: описание
                - face_count: количество полигонов
                - category: категория использования (опционально)
                - account_type: тип аккаунта ('basic', 'pro', 'premium')
                - is_downloadable: доступна ли для скачивания
                - has_textures: есть ли текстуры
                - is_rigged: есть ли риг
                - is_animated: есть ли анимация
        
        Returns:
            dict с оценкой и деталями
        """
        scores = {}
        weights = {}
        
        # 1. Оценка описания (25%)
        description_score = self._rate_description(
            model_data.get('description', '')
        )
        scores['description'] = description_score
        weights['description'] = 0.25
        
        # 2. Оценка тегов (25%)
        tags_score = self._rate_tags(
            model_data.get('tags', []),
            model_data.get('category', 'generic')
        )
        scores['tags'] = tags_score
        weights['tags'] = 0.25
        
        # 3. Оценка полигонов (20%)
        polygon_score = self._rate_polygons(
            model_data.get('face_count', 0),
            model_data.get('category', 'generic')
        )
        scores['polygons'] = polygon_score
        weights['polygons'] = 0.20
        
        # 4. Оценка аккаунта (15%)
        account_score = self._rate_account(
            model_data.get('account_type', 'basic'),
            model_data.get('author_followers', 0)
        )
        scores['account'] = account_score
        weights['account'] = 0.15
        
        # 5. Оценка технических характеристик (15%)
        technical_score = self._rate_technical(model_data)
        scores['technical'] = technical_score
        weights['technical'] = 0.15
        
        # Вычисляем взвешенную сумму
        total_score = sum(scores[key] * weights[key] for key in scores)
        
        return {
            'total_score': round(total_score, 2),
            'grade': self._get_grade(total_score),
            'scores': scores,
            'recommendations': self._get_recommendations(scores, model_data)
        }
    
    def _rate_description(self, description):
        """Оценка качества описания (0-100)"""
        if not description:
            return 0
        
        score = 0
        description = description.strip()
        length = len(description)
        
        # Длина описания
        if length >= 200:
            score += 30
        elif length >= 100:
            score += 20
        elif length >= 50:
            score += 10
        
        # Наличие ключевых слов
        keywords = [
            'model', '3d', 'texture', 'polygon', 'uv', 'material',
            'pbr', 'low poly', 'high poly', 'rigged', 'animated'
        ]
        found_keywords = sum(1 for kw in keywords if kw.lower() in description.lower())
        score += min(found_keywords * 5, 25)
        
        # Структурированность (наличие пунктуации, списков)
        has_punctuation = bool(re.search(r'[.!?;:]', description))
        has_lists = bool(re.search(r'[-*•\n]', description))
        if has_punctuation:
            score += 15
        if has_lists:
            score += 10
        
        # Наличие технических деталей (числа, измерения)
        has_numbers = bool(re.search(r'\d+', description))
        if has_numbers:
            score += 10
        
        # Разнообразие слов
        words = re.findall(r'\b\w+\b', description.lower())
        unique_ratio = len(set(words)) / max(len(words), 1)
        if unique_ratio > 0.7:
            score += 10
        
        return min(score, 100)
    
    def _rate_tags(self, tags, category='generic'):
        """Оценка качества тегов (0-100)"""
        if not tags:
            return 0
        
        score = 0
        tags_lower = [str(tag).lower() for tag in tags]
        
        # Количество тегов
        tag_count = len(tags)
        if tag_count >= 10:
            score += 30
        elif tag_count >= 5:
            score += 20
        elif tag_count >= 3:
            score += 10
        
        # Релевантность категории
        recommended = []
        for cat_key, cat_tags in self.recommended_tags.items():
            if cat_key in category.lower():
                recommended = cat_tags
                break
        
        if not recommended:
            recommended = self.recommended_tags.get('prop', [])
        
        matching_tags = sum(1 for tag in tags_lower if any(rec in tag for rec in recommended))
        relevance_score = min(matching_tags * 10, 40)
        score += relevance_score
        
        # Специфичность (не только общие теги)
        generic_tags = ['3d', 'model', 'object', 'asset']
        specific_tags = sum(1 for tag in tags_lower if tag not in generic_tags)
        if specific_tags >= 5:
            score += 20
        elif specific_tags >= 3:
            score += 10
        
        # Популярные теги Sketchfab
        popular_tags = ['pbr', 'lowpoly', 'game', 'realtime', 'blender', 'maya']
        has_popular = any(tag in tags_lower for tag in popular_tags)
        if has_popular:
            score += 10
        
        return min(score, 100)
    
    def _rate_polygons(self, face_count, category='generic'):
        """Оценка оптимальности полигонов (0-100)"""
        if face_count <= 0:
            return 0
        
        # Определяем оптимальный диапазон
        optimal_range = self.polygon_ranges.get(category, self.polygon_ranges['generic'])
        min_optimal, max_optimal = optimal_range
        
        # Если в оптимальном диапазоне - высокая оценка
        if min_optimal <= face_count <= max_optimal:
            # Чем ближе к середине, тем лучше
            mid = (min_optimal + max_optimal) / 2
            distance_from_mid = abs(face_count - mid)
            max_distance = (max_optimal - min_optimal) / 2
            score = 100 - (distance_from_mid / max_distance) * 20
            return max(score, 80)
        
        # Если меньше оптимума
        if face_count < min_optimal:
            ratio = face_count / min_optimal
            return max(ratio * 70, 20)
        
        # Если больше оптимума
        excess_ratio = (face_count - max_optimal) / max_optimal
        if excess_ratio < 1:
            return 70 - (excess_ratio * 30)
        elif excess_ratio < 5:
            return 40 - (min(excess_ratio - 1, 3) * 10)
        else:
            return 10
    
    def _rate_account(self, account_type, followers):
        """Оценка аккаунта (0-100)"""
        score = 0
        
        # Тип аккаунта
        if account_type == 'premium':
            score += 50
        elif account_type == 'pro':
            score += 35
        else:  # basic
            score += 15
        
        # Количество подписчиков
        if followers >= 1000:
            score += 50
        elif followers >= 500:
            score += 40
        elif followers >= 100:
            score += 30
        elif followers >= 50:
            score += 20
        elif followers >= 10:
            score += 10
        
        return min(score, 100)
    
    def _rate_technical(self, model_data):
        """Оценка технических характеристик (0-100)"""
        score = 0
        
        # Доступность для скачивания
        if model_data.get('is_downloadable', False):
            score += 20
        
        # Наличие текстур
        if model_data.get('has_textures', False):
            score += 25
        
        # PBR материалы
        if model_data.get('has_pbr', False):
            score += 25
        
        # Риггинг
        if model_data.get('is_rigged', False):
            score += 15
        
        # Анимация
        if model_data.get('is_animated', False):
            score += 15
        
        return min(score, 100)
    
    def _get_grade(self, score):
        """Определение буквенной оценки"""
        if score >= 90:
            return 'A+'
        elif score >= 85:
            return 'A'
        elif score >= 80:
            return 'A-'
        elif score >= 75:
            return 'B+'
        elif score >= 70:
            return 'B'
        elif score >= 65:
            return 'B-'
        elif score >= 60:
            return 'C+'
        elif score >= 55:
            return 'C'
        elif score >= 50:
            return 'C-'
        elif score >= 40:
            return 'D'
        else:
            return 'F'
    
    def _get_recommendations(self, scores, model_data):
        """Генерация рекомендаций по улучшению"""
        recommendations = []
        
        # Описание
        if scores['description'] < 50:
            recommendations.append(
                "Добавьте подробное описание: укажите назначение модели, "
                "технические характеристики, использованные инструменты"
            )
        
        # Теги
        if scores['tags'] < 60:
            recommendations.append(
                "Добавьте больше релевантных тегов (минимум 5-10). "
                "Используйте популярные теги: pbr, lowpoly, game, realtime"
            )
        
        # Полигоны
        if scores['polygons'] < 60:
            face_count = model_data.get('face_count', 0)
            category = model_data.get('category', 'generic')
            optimal = self.polygon_ranges.get(category, self.polygon_ranges['generic'])
            recommendations.append(
                f"Оптимизируйте количество полигонов. Текущее: {face_count:,}, "
                f"рекомендуемое для {category}: {optimal[0]:,}-{optimal[1]:,}"
            )
        
        # Аккаунт
        if scores['account'] < 40:
            recommendations.append(
                "Рассмотрите возможность перехода на PRO аккаунт для "
                "расширенных возможностей и повышения доверия"
            )
        
        # Технические
        if scores['technical'] < 50:
            if not model_data.get('is_downloadable'):
                recommendations.append("Включите возможность скачивания модели")
            if not model_data.get('has_textures'):
                recommendations.append("Добавьте текстуры для повышения качества")
        
        return recommendations


def example_usage():
    """Пример использования"""
    rater = QualityRater()
    
    # Пример данных модели
    model = {
        'tags': ['lowpoly', 'game', 'character', 'pbr', 'rigged', 'unity'],
        'description': '''
            Low-poly game character with PBR textures.
            - Optimized for real-time rendering
            - Includes diffuse, normal, and roughness maps
            - Fully rigged and ready for animation
            - 8,500 polygons
            - UV unwrapped
        ''',
        'face_count': 8500,
        'category': 'game_desktop',
        'account_type': 'pro',
        'author_followers': 150,
        'is_downloadable': True,
        'has_textures': True,
        'has_pbr': True,
        'is_rigged': True,
        'is_animated': False
    }
    
    result = rater.calculate_quality_score(model)
    
    print("=" * 60)
    print("ОЦЕНКА КАЧЕСТВА МОДЕЛИ")
    print("=" * 60)
    print(f"\nОбщая оценка: {result['total_score']}/100 (Grade: {result['grade']})")
    print("\nДетальные оценки:")
    for key, value in result['scores'].items():
        print(f"  - {key.capitalize()}: {value:.1f}/100")
    
    if result['recommendations']:
        print("\nРекомендации по улучшению:")
        for i, rec in enumerate(result['recommendations'], 1):
            print(f"  {i}. {rec}")
    else:
        print("\nОтлично! Модель соответствует высоким стандартам.")
    print("=" * 60)


if __name__ == "__main__":
    example_usage()
