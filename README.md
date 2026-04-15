# w2g

Минимальный сервис совместного просмотра VK Video с чатом на Go.

## Что есть в MVP

- Один общий рум без регистрации.
- Iframe-плеер VK Video.
- Реалтайм-чат для двух участников через WebSocket.
- Имена в чате: `Me` и `NotMe`.

## Локальный запуск без Docker

```bash
go mod tidy
go run ./cmd/server
```

Сервис будет доступен на `http://localhost:8080`.

## Тесты

```bash
go test ./internal/repo/ -v
```

Флаг `-v` — подробный вывод, `-count=1` — отключить кеш:

```bash
go test ./... -v -count=1
```

## Запуск через Docker Compose

```bash
docker compose up -d --build
```

Проверка:

```bash
curl http://localhost:8080/healthz
```

Остановка:

```bash
docker compose down
```
