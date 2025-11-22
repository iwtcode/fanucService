<div align="center">

# Fanuc Focas Service

![alt text](https://img.shields.io/badge/Go-1.19+-00ADD8?logo=go)
![alt text](https://img.shields.io/badge/Fanuc-Focas-yellow)
![alt text](https://img.shields.io/badge/Apache%20Kafka-Integrated-blue?logo=apachekafka)
![alt text](https://img.shields.io/badge/PostgreSQL-Supported-336791?logo=postgresql)
![alt text](https://img.shields.io/badge/Docker-Ready-2496ED?logo=docker)
![alt text](https://img.shields.io/badge/License-MIT-green)

*Сервис для сбора данных со станков Fanuc по протоколу Focas, отправки в Apache Kafka и управления через REST API*

</div>

### ✨ Ключевые возможности
- 🚀 **Потоковая передача в Kafka**: Данные в реальном времени отправляются в топик Apache Kafka.
- 🔐 **Безопасность**: Доступ к API защищен с помощью `X-API-Key`.
- 🕹️ **Управляемый опрос**: Запуск и остановка мониторинга для каждого станка через API.
- 💾 **Персистентность**: Состояния подключений сохраняются в PostgreSQL для автоматического восстановления после перезагрузки.
- 🏭 **Fanuc Focas Integration**: Использование обертки над библиотой Fanuc (Fwlib).
- 🐳 **Простота развертывания**: Готовая конфигурация docker-compose.

## 🏗️ Архитектура

```
┌─────────────────┐      ┌─────────────────┐      ┌──────────────────┐
│   Управляющий   ├─────▸│     Сервис      │◂─────┤    Fanuc CNC     │
│    REST API     │      │  fanucService   │      │     Adapter      │
│   (Gin-Gonic)   │      │    (Go App)     │      │     (Focas)      │
└─────────────────┘      └───────┬───┬─────┘      └──────────────────┘
        ▴                        │   │      (Polling)
        │                        │   └─────────────────────┐
        │                        ▾                         ▾
┌───────┴─────────┐      ┌─────────────────┐      ┌──────────────────┐
│  Пользователь / │      │   PostgreSQL    │      │   Apache Kafka   │
│     Система     │      │   (Состояния    │      │   (Потоковая     │
│   (Управление)  │      │   Подключений)  │      │   обработка)     │
└─────────────────┘      └─────────────────┘      └──────────────────┘
```

## 📦 Установка

1️⃣ **Клонирование репозитория**

```bash
git clone https://github.com/iwtcode/fanucService.git
cd fanucService
```

2️⃣ **Конфигурация приложения**

Откройте файл `.env` и при необходимости измените его

```dotenv
# App
APP_PORT=8080
GIN_MODE=debug
API_KEY=secret_key

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=1234
DB_NAME=fanuc_db

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=fanuc_data
```

3️⃣ **Запуск Apache Kafka**

```bash
docker compose up -d
```

4️⃣ **Запуск приложения**

```bash
# Linux
./build/fanuc_service

# Golang
go run cmd/app/main.go
```

## 🔌 API

🔒 **Аутентификация**: Все запросы должны содержать заголовок `X-API-Key`.

## Создание подключения

```http
POST /api/v1/connect
```

```bash
curl -X 'POST' \
  'http://localhost:8080/api/v1/connect' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key' \
  -H 'Content-Type: application/json' \
  -d '{
    "endpoint": "192.168.56.1:8193",
    "timeout": 5000,
    "model": "FS0i-D",
    "series": "0i"
}'
```

```json
{
  "status": "ok",
  "data": {
    "id": "90e09ee9-7d39-4a15-8a00-b7fb351b27ee",
    "endpoint": "192.168.56.1:8193",
    "timeout": 5000,
    "model": "FS0i-D",
    "series": "0i",
    "interval": 0,
    "status": "connected",
    "mode": "static",
    "created_at": "2025-11-22T21:40:17.465186444+03:00",
    "updated_at": "2025-11-22T21:40:17.465186629+03:00"
  }
}
```

## Получение списка подключений и проверка их актуальности

```http
GET /api/v1/connect
```

```bash
curl -X 'GET' \
  'http://localhost:8080/api/v1/connect' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key'
```

```json
{
  "status": "ok",
  "data": [
    {
      "id": "90e09ee9-7d39-4a15-8a00-b7fb351b27ee",
      "endpoint": "192.168.56.1:8193",
      "timeout": 5000,
      "model": "FS0i-D",
      "series": "0i",
      "interval": 0,
      "status": "connected",
      "mode": "static",
      "created_at": "2025-11-22T21:40:17.465186+03:00",
      "updated_at": "2025-11-22T21:40:17.465186+03:00"
    },
    {
      "id": "667204be-5e3c-433f-9700-ea931ee14f63",
      "endpoint": "192.168.56.1:8195",
      "timeout": 5000,
      "model": "FS30i-D",
      "series": "30i",
      "interval": 0,
      "status": "reconnecting",
      "mode": "polled",
      "created_at": "2025-11-22T21:48:17.087876+03:00",
      "updated_at": "2025-11-22T21:48:17.087876+03:00"
    }
  ]
}
```

## Получение конкретного подключения и проверка его актуальности

```http
POST /api/v1/connect?id={uuid}
```

```bash
curl -X 'GET' \
  'http://localhost:8080/api/v1/connect?id=90e09ee9-7d39-4a15-8a00-b7fb351b27ee' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key'
```

```json
{
  "status": "ok",
  "data": {
    "id": "90e09ee9-7d39-4a15-8a00-b7fb351b27ee",
    "endpoint": "192.168.56.1:8193",
    "timeout": 5000,
    "model": "FS0i-D",
    "series": "0i",
    "interval": 0,
    "status": "connected",
    "mode": "static",
    "created_at": "2025-11-22T21:40:17.465186+03:00",
    "updated_at": "2025-11-22T21:40:17.465186+03:00"
  }
}
```

## Запуск сбора данных

```http
POST /api/v1/polling/start
```

```bash
curl -X 'POST' \
  'http://localhost:8080/api/v1/polling/start' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": "90e09ee9-7d39-4a15-8a00-b7fb351b27ee",
  "interval": 10000
}'
```

```json
{
  "status": "ok",
  "message": "Polling started for session 90e09ee9-7d39-4a15-8a00-b7fb351b27ee"
}
```

## Остановка сбора данных

```http
POST /api/v1/polling/stop
```

```bash
curl -X 'POST' \
  'http://localhost:8080/api/v1/polling/stop' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key' \
  -H 'Content-Type: application/json' \
  -d '{
  "id": "90e09ee9-7d39-4a15-8a00-b7fb351b27ee"
}'
```

```json
{
  "status": "ok",
  "message": "Polling stopped for session 90e09ee9-7d39-4a15-8a00-b7fb351b27ee"
}
```

## Удаление подключения

```http
DELETE /api/v1/connect?id={uuid}
```

```bash
curl -X 'DELETE' \
  'http://localhost:8080/api/v1/connect?id=90e09ee9-7d39-4a15-8a00-b7fb351b27ee' \
  -H 'accept: application/json' \
  -H 'X-API-Key: secret_key'
```

```json
{
  "status": "ok",
  "message": "Session 90e09ee9-7d39-4a15-8a00-b7fb351b27ee successfully deleted"
}
```

## 🔧 Структура проекта

```
fanucService/
├── cmd/
│   └── app/                # Точка входа в приложение
├── internal/               # Приватный код приложения
│   ├── app/                # Сборка зависимостей
│   ├── domain/             # Основные сущности и модели данных
│   │   ├── entities/       # Структуры базы данных
│   │   └── models/         # DTO для API и ошибки
│   ├── handlers/           # HTTP слой
│   ├── interfaces/         # Абстракции для развязывания слоев
│   ├── middleware/         # Обёртки над функциями
│   ├── repository/         # Слой доступа к базе данных
│   ├── services/           # Инфраструктурные сервисы и логика работы с оборудованием
│   │   ├── fanuc/          # Логика соединения со станками и опроса
│   │   └── kafka/          # Логика отправки данных в Kafka
│   └── usecases/           # Бизнес-логика
├── .env                    # Конфигурация переменных окружения
├── client.go               # SDK для взаимодействия с этим сервисом
├── config.go               # Загрузка конфигурации приложения
├── models.go               # Общие модели, экспортируемые для клиента SDK
└── docker-compose.yml      # Запуск Kafka-UI
```

## 🆘 Поддержка

- 🐛 [Создайте issue](https://github.com/iwtcode/fanucService/issues)
- 📧 Напишите на email: iwtcode@gmail.com

## 📝 Лицензия

Проект распространяется под [лицензией MIT](LICENSE)

Copyright (c) 2025 iwtcode