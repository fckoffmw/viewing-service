# w2g: запуск

## Локальный запуск

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

## Docker Compose

```bash
docker compose up -d --build
```

Сервис доступен на `http://<SERVER_IP>:8080`

## Проверка

### Health endpoint

```bash
curl http://localhost:8080/healthz
```

Ответ: `ok`

### Чат

1. Открыть сервис в двух вкладках
2. Отправить сообщение в одной
3. Убедиться, что оно появилось в другой

## Остановка

Docker Compose:

```bash
docker compose down
```

Go run: `Ctrl+C`

## Тесты

```bash
go test ./internal/repo/ -v
```

Все пакеты:

```bash
go test ./... -v -count=1
```
