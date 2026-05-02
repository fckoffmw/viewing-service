# REST API

## Аутентификация

### POST /auth/register

Регистрация нового пользователя.

**Request:**
```json
{
  "username": "user123",
  "password": "password123"
}
```

**Response 201:**
- `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800`
- `null`

**Response 400:**
```json
{"error": "username cannot be empty"}
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
{
  "username": "user123",
  "password": "password123"
}
```

**Response 200:**
- `Set-Cookie: session_id=...; HttpOnly; Path=/; Max-Age=604800`
- `null`

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
  "id": "user-id",
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
{
  "name": "Название",
  "url": "https://..."
}
```

**Response 201:**
```json
{"id": "new-id"}
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
  "created_at": "2025-01-01T00:00:00Z"
}
```

---

### DELETE /api/rooms/{invite_code}

Удалить комнату. Только owner. Защищённый endpoint.

**Response 200:**
```json
{"status": "deleted"}
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

---

## WebSocket

### GET /ws/{invite_code}

Чат комнаты. Защищённый endpoint (требует session_id cookie).

**Сообщение:**
```json
{
  "username": "user123",
  "text": "Привет!"
}
```

---

## Заголовки

### X-Request-ID

Идентификатор запроса (8 символов UUID) для трейсинга.