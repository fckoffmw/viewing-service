# w2g: REST API

## Источники

### `GET /api/sources`

Возвращает список всех источников.

**Response 200:**

```json
[
  {
    "id": "1",
    "name": "Семь (1995)",
    "url": "https://vk.ru/video_ext.php?oid=-231263435&id=456240185&hash=d1213ab8896c93b2&hd=4"
  }
]
```

## Комната

### `GET /api/room`

Возвращает информацию о глобальной комнате.

**Response 200:**

```json
{
  "id": "1",
  "name": "global",
  "current_source": {
    "id": "1",
    "name": "Семь (1995)",
    "url": "https://vk.ru/video_ext.php?oid=-231263435&id=456240185&hash=d1213ab8896c93b2&hd=4"
  }
}
```

### `PATCH /api/room/source`

Устанавливает активный источник в глобальной комнате.

**Request:**

```json
{
  "source_id": "1"
}
```

**Response 200:**

```json
{
  "id": "1",
  "message": "ok"
}
```

**Response 400:**

```json
{
  "message": "some error message"
}
```

**Response 500:**

```json
{
  "message": "some error message"
}
```

## WebSocket

### `GET /ws`

WebSocket для чата.

**Сообщение (JSON):**

```json
{
  "clientId": "user_abc123",
  "text": "Привет!"
}
```
