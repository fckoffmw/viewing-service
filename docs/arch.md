# Архитектура

## Обзор

```
cmd/w2g/main.go ──→ HTTP Server ──→ middleware.Logging ──→ middleware.Auth ──→ Routes
                                                                              │
                                                                              ├── WebSocket (/ws)
                                                                              ├── Auth API (/auth/*) - public
                                                                              ├── Health (/healthz) - public
                                                                              └── API (/api/*) - protected
```

Статика (HTML, CSS, JS) раздается через nginx.

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
- Пропускает публичные маршруты: `/auth/*`, `/healthz`, `/ws/*`
- Для защищённых маршрутов проверяет `session_id` cookie
- Sliding expiry: обновляет `ExpiresAt` при каждом запросе

## WebSocket чат

```
Client ──→ readPump ──→ hub.broadcast ──→ writePump ──→ Client
```

- HubManager — управляет RoomHub-ами по invite_code
- RoomHub — одна goroutine с channel dispatch
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
| GET | `/healthz` | Health check |
| POST | `/auth/register` | auth.Register |
| POST | `/auth/login` | auth.Login |
| POST | `/auth/logout` | auth.Logout |
| GET | `/auth/me` | auth.Me |

### Защищённые

| Метод | Путь | Обработчик |
|-------|------|----------|
| GET | `/api/sources` | source.GetAllSources |
| POST | `/api/sources` | source.AddSource |
| POST | `/api/rooms` | room.CreateRoom |
| GET | `/api/rooms/{invite_code}` | room.GetRoom |
| DELETE | `/api/rooms/{invite_code}` | room.DeleteRoom |
| PATCH | `/api/rooms/{invite_code}/source` | room.PatchRoomSource |

### WebSocket

| Метод | Путь | Обработчик |
|-------|------|----------|
| GET | `/ws/{invite_code}` | chat.ServeWS |

## Зависимости

| Модуль | Версия |
|--------|--------|
| `gorilla/websocket` | v1.5.3 |
| `github.com/google/uuid` | v1.6.0 |
| `golang.org/x/crypto` | bcrypt |