# Order Service

Микросервис для отображения информации о заказах с использованием:
- PostgreSQL для хранения данных
- Kafka для обработки сообщений
- In-memory кэша для быстрого доступа

## Функционал

- Приём заказов через Kafka
- Сохранение в PostgreSQL
- Кэширование в памяти
- REST API для доступа к данным
- Веб-интерфейс для просмотра заказов

## Технологии

- **Backend**: Go
- **Database**: PostgreSQL
- **Message Broker**: Apache Kafka
- **Web Framework**: chi
- **Frontend**: HTML/CSS/JS

## Запуск проекта

### Требования
- Docker и Docker Compose
- Go 1.20+

### 1. Запуск инфраструктуры
```bash
docker-compose up -d
```

### 2. Настройка БД
Создайте базу данных и таблицы (SQL-скрипты в `init.sql`)

### 3. Запуск сервиса
```bash
go run cmd/main/main.go
```

### 4. Запуск продюсера (отправка тестовых данных)
```bash
go run cmd/producer/main.go
```

## Доступ

- **API**: `http://localhost:8064/api/orders/{order_uid}`
- **Web UI**: `http://localhost:8064`

## Структура проекта

```
.
├── cmd
│   ├── initdb        # Создание БД и пользователя
│   ├── main          # Основной сервис
│   ├── migrator      # Запуск миграций
│   └── producer      # Тестовый продюсер
├── config
├── internal
│   ├── cache         # Кэш в памяти
│   ├── config        # Конфигурация
│   ├── http-server   # HTTP handlers
│   ├── kafka         # Kafka consumer
│   ├── lib           # Дополнительные логгеры
│   ├── models        # Модели данных
│   ├── service       # Основная работа с данными
│   └── storage       # PostgreSQL хранилище
├── migrations        # SQL-миграции
├── scripts           # SQL-запрос для создания БД и пользователя
└── static            # Веб-интерфейс
```

## Тестирование

1. Отправьте тестовый заказ:
```bash
go run cmd/producer/main.go
```

2. Проверьте в веб-интерфейсе:
```
http://localhost:8064
```

3. Или через API:
```bash
curl http://localhost:8081/api/orders/{orderUID}
```

## Особенности реализации

- Автоматическое восстановление кэша из БД при старте
- Обработка некорректных сообщений Kafka
- Гибкая конфигурация через yaml-файл
- Логирование всех операций