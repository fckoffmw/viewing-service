# sources

## GET /api/sources
response:
```json
[
    {
        "id":"1",
        "name":"Семь (1995)",
        "url":"https://kek.ru"
    }
]
```

# rooms

## GET /api/room
response:
```json
{
    "id":"1",
    "name":"global",
    "current_source": {
        "id":"1",
        "name":"Семь (1995)",
        "url":"https://kek.ru"
    }
}
```

## PATCH /api/room/source
request:
```json
{
    "source_id": "1",
}
```
response:

status 200
```json
{
    "id": "1",
    "message": "ok"
}

status 400
```json
{
    "message": "some error message"
}
```

status 500
```json
{
    "message": "some error message"
}
```



