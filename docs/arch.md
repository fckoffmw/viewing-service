# Архитектура

## Обзор

```
cmd/w2g/main.go ──→ HTTP Server ──→ middleware.Logging ──→ middleware.Auth ──→ Routes
                                                                              │
                                                                              ├── Static (web/, /static/)
                                                                              ├── WebSocket (/ws)
                                                                              ├── Auth API (/auth/*) - public
                                                                              └── API (/api/*) - protected
```

## Пакеты

| Пакет | Назначение |
|-----|-----------|
| `auth` | Аутентификация (Register, Login, Logout) |
| `chat` | WebSocket: Hub + Client |
| `middleware` | HTTP middleware (logging, auth) |
| `repo` | CSV хранилище |
| `room` | Управление комнатами |
| `source` | Управление источниками |
| `response` | Утилиты для HTTP ответов |
| `errors` | Кастомные ошибки |
| `config` | Конфигурация |

## Middleware

### logging
Логирование запросов:
- Генерирует `request_id` (8 символов UUID)
- Добавляет заголовок `X-Request-ID` в ответ
- Логирует: method, path, status, duration

### auth
Аутентификация:
- Пропускает публичные маршруты: `/`, `/login.html`, `/register.html`, `/auth/*`, `/healthz`, `/ws`, `/static/*`
- Для защищённых маршрутов проверяет `session_id` cookie
- Sliding expiry: обновляет `ExpiresAt` при каждом запросе

## WebSocket чат

```
Client ──→ readPump ──→ hub.broadcast ──→ writePump ──→ Client
```

- Max 2 клиента
- Hub — одна goroutine с channel dispatch
- Keepalive: ping/pong каждые 54с

## Хранение

CSV файлы в `./storage/`:
- `users.csv` — пользователи
- `sources.csv` — источники видео
- `rooms.csv` — комнаты

Thread-safe через `sync.RWMutex`.

## Роуты

### Публичные

| Метод | Путь | Обработчик |
|-------|------|----------|
| GET | `/` | Static (web/index.html) |
| GET | `/login.html` | Static |
| GET | `/register.html` | Static |
| GET | `/healthz` | Health check |
| POST | `/auth/register` | auth.Register |
| POST | `/auth/login` | auth.Login |
| POST | `/auth/logout` | auth.Logout |
| GET | `/auth/me` | auth.Me |
| GET | `/ws` | WebSocket чат |

### Защищённые

| Метод | Путь | Обработчик |
|-------|------|----------|
| GET | `/api/sources` | source.GetAllSources |
| POST | `/api/sources` | source.AddSource |
| GET | `/api/room` | room.GetGlobalRoom |
| PATCH | `/api/room/source` | room.PatchGlobalRoomSource |

## Зависимости

| Модуль | Версия |
|--------|--------|
| `gorilla/websocket` | v1.5.3 |
| `github.com/google/uuid` | v1.6.0 |
| `golang.org/x/crypto` | bcrypt |