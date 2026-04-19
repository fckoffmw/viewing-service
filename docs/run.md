# Запуск

## Локальный запуск

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

## Docker Compose

```bash
docker compose up -d --build
```

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
go test ./internal/... -v -count=1
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|-----------|
| `PORT` | `8080` | Порт сервера |
| `STORAGE_DIR` | `./storage/` | Директория CSV-файлов |
| `LOG_LEVEL` | `debug` | Уровень логирования (`debug`, `info`, `warn`, `error`) |

`MaxClients` захардкожен в `2` (ограничение чата).