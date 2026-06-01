# REST API

Все ошибки возвращают `{"error": "..."}`.

---

## Аутентификация

### POST /auth/register

Регистрация нового пользователя.

**Request:**
```json
{"username": "user123", "password": "password123"}
```

**Response 201:**
- `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800`
- `null`

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

Вход в систему.

**Request:**
```json
{"username": "user123", "password": "password123"}
```

**Response 200:**
- `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800`
- `null`

**Response 400:**
```json
{"error": "cannot read req body"}
```

**Response 401:**
```json
{"error": "invalid credentials"}
```

---

### POST /auth/logout

Выход из системы. Требует `session_id` cookie.

**Response 200:**
- `Set-Cookie: session_id=; Max-Age=-1`

---

### GET /auth/me

Информация о текущем пользователе. Требует `session_id` cookie.

**Response 200:**
```json
{
  "id": "1",
  "username": "user123"
}
```

**Response 401:**
```json
{"error": "session not found"}
```

---

## Источники

### GET /api/sources

Список источников. Защищённый endpoint.

**Response 200:**
```json
[
  {
    "id": "1",
    "name": "Семь (1995)",
    "url": "https://vkvideo.ru/..."
  }
]
```

---

### POST /api/sources

Добавить источник. Защищённый endpoint.

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

---

## Комнаты

### POST /api/rooms

Создать комнату. Защищённый endpoint.

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

---

### GET /api/rooms

Список всех комнат. Защищённый endpoint.

**Response 200:**
```json
[
  {
    "id": "1",
    "name": "Movie Night",
    "invite_code": "X7K2PQ4M",
    "owner_id": "1",
    "members_online": 0,
    "current_source": {
      "id": "1",
      "name": "Название",
      "url": "https://..."
    },
    "created_at": "2025-01-01T00:00:00Z"
  }
]
```

Поле `current_source` присутствует только если для комнаты выбран источник.

---

### GET /api/rooms/{invite_code}

Информация о комнате. Защищённый endpoint.

**Response 200:**
```json
{
  "id": "1",
  "name": "Movie Night",
  "invite_code": "X7K2PQ4M",
  "owner_id": "1",
  "members_online": 0,
  "current_source": {
    "id": "1",
    "name": "Название",
    "url": "https://..."
  },
  "created_at": "2025-01-01T00:00:00Z"
}
```

Поле `current_source` присутствует только если для комнаты выбран источник.

**Response 404:**
```json
{"error": "room not found"}
```

---

### DELETE /api/rooms/{invite_code}

Удалить комнату. Только owner. Защищённый endpoint.

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

---

### PATCH /api/rooms/{invite_code}/source

Установить активный источник. Только owner. Защищённый endpoint.

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

---

## WebSocket

### GET /ws/{invite_code}

Чат комнаты. Защищённый endpoint (требует `session_id` cookie).

**Клиент → сервер:**
```json
{"text": "Привет!"}
```

**Сервер → клиент (чат):**
```json
{"username": "user123", "text": "Привет!"}
```

**Сервер → клиент (смена источника):**
```json
{
  "type": "source_changed",
  "payload": {
    "source_id": "1",
    "source_url": "https://..."
  }
}
```

**Response 401 (неавторизован):**
- `401 Unauthorized`

---

## Заголовки

### X-Request-ID

Идентификатор запроса (8 символов, первая часть UUID) для трейсинга. Добавляется в каждый ответ.
