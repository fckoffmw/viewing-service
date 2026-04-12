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

