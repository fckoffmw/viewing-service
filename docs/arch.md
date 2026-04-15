# Архитектура

## Обзор

```
cmd/w2g/main.go ──→ HTTP Server ──→ http.ServeMux ──→ Routes
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
| `repo` | CSV хранилище |
| `room` | Управление комнатами |
| `source` | Управление источниками |

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

Thread-safe через `sync.RWMutex`. Каждый запрос читает файл целиком.

## Роуты

| Метод | Путь | Обработчик |
|-------|------|------------|
| `GET` | `/` | Static (web/index.html) |
| `GET` | `/ws` | WebSocket чат |
| `GET` | `/healthz` | Health check |
| `POST` | `/api/login` | auth.Login |
| `GET` | `/api/sources` | source.GetAllSources |
| `GET` | `/api/room` | room.GetGlobalRoom |
| `PATCH` | `/api/room/source` | room.PatchGlobalRoomSource |

## Зависимости

| Модуль | Версия |
|--------|--------|
| `gorilla/websocket` | v1.5.3 |
| `golang.org/x/crypto` | v0.49.0 |
