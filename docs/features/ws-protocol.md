# WebSocket Protocol — w2g

## Подключение

```
ws://localhost:8080/ws/{invite_code}
```

Аутентификация — через cookie `session_id` (HTTP upgrade request). Без валидной сессии — 401.

## Входящие сообщения (клиент → сервер)

```json
{ "type": "play",  "payload": { "position": 42.5 } }
{ "type": "pause", "payload": { "position": 42.5 } }
{ "type": "seek",  "payload": { "position": 120.0 } }
{ "type": "chat",  "payload": { "text": "hello" } }
{ "type": "sticker", "payload": { "id": "0" } }
```

| type | Payload | Описание |
|------|---------|----------|
| `play` | `{ "position": float }` | Запуск видео с позиции |
| `pause` | `{ "position": float }` | Пауза на позиции |
| `seek` | `{ "position": float }` | Перемотка на позицию |
| `chat` | `{ "text": string }` | Чат-сообщение (макс 1000 символов) |
| `sticker` | `{ "id": string }` | ID стикера |

## Исходящие сообщения (сервер → клиент)

### sync — отправляется ТОЛЬКО новому клиенту при подключении
```json
{
  "type": "sync",
  "payload": {
    "source_id":  "1",
    "source_url": "https://vk.com/...",
    "playing":    true,
    "position":   87.3,
    "updated_at": "2025-01-01T00:00:42Z"
  }
}
```

### play / pause / seek — broadcast всем в комнате (включая отправителя)
```json
{ "type": "play",  "username": "alice", "payload": { "position": 42.5 } }
{ "type": "pause", "username": "alice", "payload": { "position": 42.5 } }
{ "type": "seek",  "username": "alice", "payload": { "position": 120.0 } }
```

### source_changed — broadcast всем (отправитель = nil, т.к. приходит через HTTP)
```json
{ "type": "source_changed", "payload": { "source_id": "2", "source_url": "https://..." } }
```

### chat — broadcast всем КРОМЕ отправителя
```json
{ "type": "chat", "username": "alice", "payload": { "text": "hello" } }
```

### sticker — broadcast всем КРОМЕ отправителя
```json
{ "type": "sticker", "username": "alice", "payload": { "id": "0" } }
```

## Ключевые правила

1. **Сервер — источник истины.** Клиент шлёт намерение, сервер обновляет `State` и рассылает факт всем (включая отправителя).
2. **sync** приходит в момент подключения — содержит снапшот текущего состояния комнаты.
3. **chat** и **sticker** не идут отправителю (остальные — всем).
4. **source_changed** приходит через HTTP-эндпоинт, затем проксируется в WS.
5. **State (Position, Playing, UpdatedAt) in-memory**, не сохраняется в CSV.

## Полезные HTTP-эндпоинты

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/auth/register` | Регистрация + session_id |
| POST | `/auth/login` | Логин + session_id |
| GET | `/auth/me` | Текущий пользователь |
| GET | `/api/sources` | Список источников |
| POST | `/api/sources` | Добавить источник |
| PATCH | `/api/sources/{id}` | Обновить источник |
| DELETE | `/api/sources/{id}` | Удалить источник |
| POST | `/api/rooms` | Создать комнату |
| GET | `/api/rooms` | Список комнат |
| GET | `/api/rooms/{invite_code}` | Детали комнаты |
| PATCH | `/api/rooms/{invite_code}/source` | Сменить источник |
| DELETE | `/api/rooms/{invite_code}` | Удалить комнату |
| GET | `/healthz` | Проверка сервера |
