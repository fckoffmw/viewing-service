# REST API

Все ошибки возвращают `{"error": "..."}` (кроме WebSocket upgrade).

---

## Аутентификация

### POST /auth/register

Регистрация нового пользователя.

**Request:**
```json
{"username": "user123", "password": "password123"}
```

**Response 201:** `null` + `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800; SameSite=Lax`

**Response 400:**
```json
{"error": "cannot read req body"}
{"error": "username cannot be empty"}
{"error": "username must be at least 3 characters"}
{"error": "password must be at least 4 characters"}
{"error": "user already exists"}
```

**Response 500:**
```json
{"error": "internal server error"}
```

---

### POST /auth/login

**Request:**
```json
{"username": "user123", "password": "password123"}
```

**Response 200:** `null` + `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800; SameSite=Lax`

**Response 400:**
```json
{"error": "cannot read req body"}
```

**Response 401:**
```json
{"error": "invalid credentials"}
```

**Response 500:**
```json
{"error": "internal server error"}
```

---

### POST /auth/logout

Выход из системы. Удаляет сессию из памяти, очищает cookie.

**Response 200:** пустое тело + `Set-Cookie: session_id=; Path=/; Max-Age=-1`

---

### GET /auth/me

Информация о текущем пользователе. Требует `session_id` cookie.

**Response 200:**
```json
{"id": "1", "username": "user123"}
```

**Response 401:**
```json
{"error": "session not found"}
```

---

## Источники

### GET /api/sources

Список источников. Требует `session_id` cookie.

**Response 200:**
```json
[{"id": "1", "name": "Семь (1995)", "url": "https://vkvideo.ru/..."}]
```

Пустой список: `null` (не `[]`).

**Response 500:**
```json
{"error": "cannot get sources"}
```

---

### POST /api/sources

Добавить источник. Требует `session_id` cookie.

**Request:**
```json
{"name": "Название", "url": "https://..."}
```

**Response 201:**
```json
{"id": "1"}
```

**Response 400:**
```json
{"error": "cannot read req body"}
{"error": "name and url are required"}
```

**Response 500:**
```json
{"error": "cannot add source"}
```

---

## Комнаты

### POST /api/rooms

Создать комнату. Требует `session_id` cookie.

**Request:**
```json
{"name": "Movie Night"}
```

**Response 201:**
```json
{
  "id": "1",
  "name": "Movie Night",
  "invite_code": "X7K2PQ4M",
  "invite_url": "/room/X7K2PQ4M",
  "owner_id": "1",
  "created_at": "2025-01-01T00:00:00Z"
}
```

**Response 400:**
```json
{"error": "invalid request body"}
{"error": "name is required"}
{"error": "max rooms reached"}
```

**Response 500:**
```json
{"error": "cannot create room"}
```

---

### GET /api/rooms

Список всех комнат. Требует `session_id` cookie.

**Response 200:**
```json
[
  {
    "id": "1",
    "name": "Movie Night",
    "invite_code": "X7K2PQ4M",
    "owner_id": "1",
    "members_online": 0,
    "current_source": {"id": "1", "name": "Название", "url": "https://..."},
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

`current_source` — присутствует только если для комнаты выбран источник (`omitempty`).

---

### GET /api/rooms/{invite_code}

Информация о комнате. Требует `session_id` cookie.

**Response 200:**
```json
{
  "id": "1",
  "name": "Movie Night",
  "invite_code": "X7K2PQ4M",
  "owner_id": "1",
  "members_online": 0,
  "current_source": {"id": "1", "name": "Название", "url": "https://..."},
  "created_at": "2025-01-01T00:00:00Z"
}
```

**Response 404:**
```json
{"error": "room not found"}
```

---

### DELETE /api/rooms/{invite_code}

Удалить комнату. Только owner. Требует `session_id` cookie.

**Response 200:**
```json
{"status": "deleted"}
```

**Response 403:**
```json
{"error": "not owner"}
```

**Response 404:**
```json
{"error": "room not found"}
```

**Response 500:**
```json
{"error": "cannot delete room"}
```

---

### PATCH /api/rooms/{invite_code}/source

Установить активный источник. Только owner. Требует `session_id` cookie.

**Request:**
```json
{"source_id": "1"}
```

**Response 200:**
```json
{"source_id": "1"}
```

**Response 400:**
```json
{"error": "invalid request body"}
{"error": "source not found"}
```

**Response 403:**
```json
{"error": "not owner"}
```

**Response 404:**
```json
{"error": "room not found"}
```

**Response 500:**
```json
{"error": "cannot change source"}
```

---

## Заголовки

### X-Request-ID

Идентификатор запроса (8 символов, первая часть UUID) для трейсинга. Добавляется в каждый ответ.

---

## WebSocket

### ws://localhost:8080/ws/{invite_code}

Требует `session_id` cookie (HTTP upgrade request). Без валидной сессии — `401 Unauthorized` (plain text).

Ошибки до upgrade возвращаются plain text:
- Нет cookie: `401 unauthorized`
- Невалидная сессия: `401 unauthorized`
- Нет invite_code: `400 missing invite code`

#### Формат фреймов

Каждый фрейм — JSON:
```json
{"type": "<message_type>", "username": "...", "timestamp": "...", "payload": {...}}
```

Поля `username` и `timestamp` присутствуют не для всех типов.

#### Входящие (клиент → сервер)

```json
{"type": "chat",            "payload": {"text": "hello"}}
{"type": "play",            "payload": {"position": 42.5}}
{"type": "pause",           "payload": {"position": 42.5}}
{"type": "seek",            "payload": {"position": 120.0}}
```

#### Исходящие (сервер → клиент)

| type | username | payload | Кому |
|------|----------|---------|------|
| `sync` | — | `{"source_id":"...","source_url":"...","playing":true,"position":0.0,"updated_at":"..."}` | только новому клиенту |
| `chat` | `username` | `{"text":"..."}` | всем, кроме отправителя |
| `play` | `username` | `{"position":42.5}` | всем |
| `pause` | `username` | `{"position":42.5}` | всем |
| `seek` | `username` | `{"position":120.0}` | всем |
| `source_changed` | — | `{"source_id":"...","source_url":"..."}` | всем |

#### Правила

- `username` заполнен для `chat`, `play`, `pause`, `seek`; отсутствует для `sync` и `source_changed`.
- `timestamp` — поле зарезервировано, пока не используется (`omitempty`).
- Чат: текст обрезается до 1000 символов.
- `source_changed` приходит через HTTP PATCH и проксируется в WS (sender не указан).
