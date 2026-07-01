# Архитектура

## Обзор

```
cmd/w2g/main.go ──→ HTTP Server ──→ middleware.Logging ──→ middleware.Auth ──→ Routes
                                                                                │
                                                                                ├── WebSocket (/ws/*)
                                                                                ├── Auth API (/auth/*)
                                                                                ├── Health (/healthz)
                                                                                └── API (/api/*)
```

Статика (HTML, CSS, JS) раздаётся через nginx (production) или `cmd/web` (dev-прокси).

## Taskfile

Проект использует [Taskfile](https://taskfile.dev) для автоматизации:

| Команда | Описание |
|---------|----------|
| `task run` | Тесты, линтер, сборка, запуск сервисов |
| `task stop` | Остановка всех сервисов |
| `task test` | Тесты |
| `task lint` | Линтер |
| `task build` | Сборка бинарников |

Все сервисы запускаются через `nohup`, PID сохраняются в `.task/pids/`, логи в `.task/logs/`.

## Пакеты

| Пакет | Назначение |
|-------|-----------|
| `auth` | Аутентификация (Register, Login, Logout, Me) |
| `realtime` | WebSocket: HubManager, RoomHub, Client |
| `room` | Управление комнатами (CRUD + source) |
| `source` | Управление источниками видео (CRUD) |
| `middleware` | HTTP middleware (logging, auth) |
| `repo` | CSV хранилище (общая реализация) |
| `http` | Настройка роутера |
| `http/response` | Утилиты для JSON-ответов |
| `apperrors` | Типизированные ошибки |
| `config` | Конфигурация из .env |
| `utils/ctx` | Контекст запроса (request_id, user_id) |
| `utils/str` | Генерация invite-кодов |

## Middleware

### logging

- Генерирует `request_id` (8 символов UUID)
- Добавляет заголовок `X-Request-ID` в ответ
- Логирует: method, path, status, duration

### auth

- Пропускает публичные пути: `/auth/*`, `/healthz`, `/ws/*`, `/static/*`, `/`
- Для защищённых маршрутов проверяет `session_id` cookie
- Sliding expiry: обновляет `ExpiresAt` при каждом запросе
- Устанавливает `user_id` в контекст запроса

## WebSocket

```
Client ──→ readPump ──→ hub.Incoming() ──→ handleEvent() ──→ broadcastAll() ──→ writePump ──→ Client
```

- HubManager — реестр RoomHub по invite_code
- RoomHub — одна goroutine с select на register/unregister/incoming
- Keepalive: ping/pong каждые 54с
- Синхронизация: сервер — источник истины (play/pause/seek → broadcast всем)

## Хранение

CSV файлы в `STORAGE_DIR`:

| Файл | Назначение | Мьютекс |
|------|-----------|---------|
| `users.csv` | Пользователи | `repo` (sync.RWMutex) |
| `sources.csv` | Источники видео | `repo` (sync.RWMutex) |
| `rooms.csv` | Комнаты (метаданные) | `room/store.go` (sync.RWMutex) |

Состояние плеера (playing, position) — in-memory, без персистентности.

## Роуты

### Публичные

| Метод | Путь | Обработчик |
|-------|------|-----------|
| GET | `/healthz` | health check |
| POST | `/auth/register` | auth.Register |
| POST | `/auth/login` | auth.Login |
| POST | `/auth/logout` | auth.Logout |
| GET | `/auth/me` | auth.Me |
| GET | `/ws/{invite_code}` | realtime.ServeWS (self-auth) |

### Защищённые

| Метод | Путь | Обработчик |
|-------|------|-----------|
| GET | `/api/sources` | source.GetAll |
| POST | `/api/sources` | source.Add |
| PATCH | `/api/sources/{id}` | source.Patch |
| DELETE | `/api/sources/{id}` | source.Delete |
| POST | `/api/rooms` | room.Create |
| GET | `/api/rooms` | room.GetAll |
| GET | `/api/rooms/{invite_code}` | room.Get |
| DELETE | `/api/rooms/{invite_code}` | room.Delete |
| PATCH | `/api/rooms/{invite_code}/source` | room.PatchSource |

## Зависимости

| Модуль | Версия | Назначение |
|--------|--------|-----------|
| `gorilla/websocket` | v1.5.3 | WebSocket |
| `google/uuid` | v1.6.0 | UUID |
| `golang.org/x/crypto` | v0.51.0 | bcrypt |
| `joho/godotenv` | v1.5.1 | .env загрузка |
