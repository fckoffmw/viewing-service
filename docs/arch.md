# Архитектура

## Обзор

```
cmd/w2g/main.go ──→ HTTP Server ──→ middleware.Logging ──→ http.ServeMux ──→ Routes
                                                                    │
                                                                    ├── Static (web/)
                                                                    ├── WebSocket (/ws) ──→ chat.Hub
                                                                    └── API (/api/*) ──→ Handlers
```

## Пакеты

| Пакет | Назначение |
|-------|------------|
| `auth` | Аутентификация (stub) |
| `chat` | WebSocket: Hub + Client |
| `middleware` | HTTP middleware (logging) |
| `repo` | CSV хранилище |
| `room` | Управление комнатами |
| `source` | Управление источниками |

## HTTP Middleware

**logging** — логирование запросов:
- Генерирует `request_id` (8 символов UUID)
- Добавляет в контекст запроса
- Добавляет заголовок `X-Request-ID` в ответ
- Логирует: method, path, status, duration

## WebSocket чат

```
Client ──→ readPump ──→ hub.broadcast ──→ writePump ──→ Client
```

- Max 2 клиента
- Hub — одна goroutine с channel dispatch
- Keepalive: ping/pong каждые 54с

## Хранение

CSV файлы в `./storage/`:

| Файл | Описание |
|------|----------|
| `users.csv` | Пользователи |
| `sources.csv` | Источники видео |
| `rooms.csv` | Комнаты |

Thread-safe через `sync.RWMutex`. Каждый запрос читает файл целиком.

## Роуты

| Метод | Путь | Обработчик |
|-------|------|------------|
| GET | `/` | Static (web/index.html) |
| GET | `/ws` | WebSocket чат |
| GET | `/healthz` | Health check |
| GET | `/api/sources` | source.GetAllSources |
| POST | `/api/sources` | source.AddSource |
| GET | `/api/room` | room.GetGlobalRoom |
| PATCH | `/api/room/source` | room.PatchGlobalRoomSource |

## Зависимости

| Модуль | Версия |
|--------|--------|
| `gorilla/websocket` | v1.5.3 |
| `github.com/google/uuid` | v1.6.0 |
| `github.com/joho/godotenv` | v1.5.1 |