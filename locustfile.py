from locust import HttpUser, task, between
import random
import time
import json

class MetricsUser(HttpUser):
    wait_time = between(0.01, 0.05)  # Очень маленькая задержка для высокой нагрузки

    def on_start(self):
        self.device_id = f"device_{random.randint(1, 1000)}"

    @task(3)
    def send_metric(self):
        """Отправка метрики IoT устройства"""
        payload = {
            "device_id": self.device_id,
            "timestamp": int(time.time()),
            "cpu": random.uniform(0.1, 0.9),
            "memory": random.uniform(0.2, 0.8),
            "rps": random.uniform(10, 1000),
            "network": random.uniform(1, 100)
        }

        headers = {'Content-Type': 'application/json'}
        self.client.post("/metrics", json=payload, headers=headers)

    @task(1)
    def get_analytics(self):
        """Получение аналитики"""
        self.client.get(f"/analytics/{self.device_id}")

    @task(1)
    def get_summary(self):
        """Получение сводной статистики"""
        self.client.get("/summary")

    @task(1)
    def health_check(self):
        """Проверка здоровья сервиса"""
        self.client.get("/health")


class AdminUser(HttpUser):
    wait_time = between(1, 5)

    @task(3)
    def get_all_analytics(self):
        """Получение всей аналитики (для админов)"""
        self.client.get("/analytics")

    @task(1)
    def get_cache_metrics(self):
        """Получение метрик кэша"""
        self.client.get("/cache-metrics")