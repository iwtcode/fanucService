# fanucService

```bash
go run main.go
```

```
структура проекта:
fanucService/
├── cmd/
│   └── app/
│       └── main.go             # Точка входа в приложение. Инициализирует конфигурацию и запускает fx.App.
│                               # Здесь вызывается app.New().Run().
├── docs/                       # Генерируемые файлы Swagger (swagger.json/yaml).
│
├── internal/
│   ├── app/
│   │   └── app.go              #  Определяет fx.Options, модули (Repository, Service, Handler)
│   │                           # и Lifecycle hooks (OnStart/OnStop) для запуска HTTP сервера и восстановления соединений.
│   │
│   ├── handlers/               # Слой HTTP (Presentation Layer). Использует Gin.
│   │   ├── connection.go       # Обработчики для POST/GET/DELETE /connect. Парсят JSON, валидируют и вызывают Usecase.
│   │   ├── polling.go          # Обработчики для POST /polling/start|stop.
│   │   ├── response.go         # Хелперы для унифицированных JSON ответов (success, error wrapper).
│   │   └── router.go           # Настройка роутера Gin, регистрация групп API (/api/v1) и подключение Middleware.
│   │
│   ├── domain/                 # Слой домена. Содержит структуры данных.
│   │   ├── entities/           # Модели базы данных (GORM).
│   │   │   └── machine.go      # Таблица для хранения настроек подключения (ID сессии, IP, Port, timeout, Model) и состояния опроса.
│   │   └── models/             # DTO (Data Transfer Objects) для API и внутренней логики.
│   │       ├── requests.go     # Структуры для биндинга входящих JSON (например, CreateConnectionRequest).
│   │       ├── responses.go    # Структуры ответов API.
│   │       └── errors.go       # Кастомные ошибки приложения (код, сообщение, http статус).
│   │
│   ├── interfaces/             # Интерфейсы.
│   │   ├── repository.go       # Интерфейс репозитория.
│   │   ├── service.go          # Интерфейс сервисов (FanucService, KafkaService).
│   │   └── usecase.go          # Интерфейс бизнес-логики (CreateConnection, StartPolling, Restore).
│   │
│   ├── middleware/             # Прослойки HTTP запросов.
│   │   ├── auth.go             # Проверка заголовка X-API-Key на соответствие ключу из конфига.
│   │   ├── logger.go           # Логирование времени выполнения и статуса запросов.
│   │   └── recovery.go         # Перехват паник для предотвращения падения сервера.
│   │
│   ├── repository/             # Слой доступа к данным. Реализация interfaces.Repository.
│   │   ├── connection.go       # Реализация методов работы с таблицей подключений через GORM.
│   │   ├── polling.go          # Методы обновления статусов опроса.
│   │   └── repository.go       # Инициализация подключения к Postgres, выполнение AutoMigrate.
│   │
│   ├── services/               # Слой инфраструктурных сервисов.
│   │   ├── fanuc/              # Обертка над библиотекой fanucAdapter.
│   │   │   ├── service.go      # Реализация интерфейса управления станками. Хранит пул активных клиентов fanucAdapter.
│   │   │   ├── connector.go    # Логика физического подключения (fanuc.New) и проверки связи (Ping/SystemInfo).
│   │   │   ├── poller.go       # Логика циклических горутин (Ticker), вызывающих adapter.AggregateAllData().
│   │   │   └── manager.go      # Управление map[SessionID]*fanuc.Client (добавление/удаление из памяти).
│   │   └── kafka/              # Реализация Kafka Producer.
│   │       └── producer.go     # Инициализация writer и метод SendMessage(topic, key, value).
│   │
│   └── usecases/               # Слой бизнес-логики (Application Layer).
│       ├── connection.go       # Use cases для управления подключениями
│       ├── polling.go          # Use cases для управления опросом
│       └── restore.go          # Восстановление состояния при старте
│
├── .env                        # Переменные окружения.
├── client.go                   # Публичный Go-клиент для HTTP API этого сервиса (чтобы использовать сервис как lib в других Go программах).
├── config.go                   # Публичная структура конфигурации (для использования client.go или инициализации app).
├── models.go                   # Публичные структуры запросов/ответов (шаринг между client.go и internal/handlers).
├── docker-compose.yml          # Описание контейнеров: Postgres, Kafka, Zookeeper, (опционально Kafka-UI).
├── go.mod                      # Зависимости модуля.
└── go.sum                      # Хэши зависимостей.
```