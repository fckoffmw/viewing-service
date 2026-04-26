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
- `{"error": ""}`

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
- `{"error": ""}`

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

**Response 200:**
```json
{"id": "new-id"}
```

---

## Комната

### GET /api/room

Информация о глобальной комнате. Защищённый endpoint.

**Response 200:**
```json
{
  "id": "1",
  "name": "global",
  "current_source": {
    "id": "1",
    "name": "Семь (1995)",
    "url": "https://vkvideo.ru/..."
  }
}
```

---

### PATCH /api/room/source

Установить активный источник. Защищённый endpoint.

**Request:**
```json
{"source_id": "1"}
```

**Response 200:**
```json
{"id": "1"}
```

---

## WebSocket

### GET /ws

Чат. Публичный endpoint.

**Сообщение:**
```json
{
  "clientId": "user_abc123",
  "text": "Привет!"
}
```

---

## Заголовки

### X-Request-ID

Идентификатор запроса (8 символов UUID) для трейсинга.