# Запуск

## Локальный запуск

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

## Docker Compose

```bash
cp .env.example .env
docker compose up -d --build
```

## Проверка

Health endpoint:
```bash
curl http://localhost:8080/healthz
```
Ответ: `ok`

## Остановка

- Docker Compose: `docker compose down`
- Go run: `Ctrl+C`

## Тесты

```bash
go test ./internal/... -v -count=1
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `PORT` | `8080` | Порт сервера |
| `STORAGE_DIR` | `./storage/` | Директория CSV-файлов |
| `MAX_CLIENTS` | `2` | Макс. клиентов в чате |
| `LOG_LEVEL` | `debug` | Уровень логирования |
| `LOG_FILE` | — | Путь к лог-файлу (опционально) |
| `SESSIONS_CLEANUP_INTERVAL` | `300` | Интервал очистки сессий (сек) |