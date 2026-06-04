# Запуск

## Разработка (Taskfile)

```bash
# Запустить всё (тесты + линтер + сборка + сервис)
task run

# Остановить
task stop
```

Сервис: `http://localhost:8080`
Dev-прокси (статика + прокси API/WS): `http://localhost:8081`

## Локальный запуск (без Taskfile)

```bash
# Backend
go run ./cmd/w2g

# Dev-прокси (опционально, для разработки фронта)
go run ./cmd/web -dir ./web -backend http://localhost:8080
```

Сервис доступен на `http://localhost:8080`

## Docker Compose

```bash
cp .env.example .env
docker compose up -d --build
```

## Деплой (systemd + nginx)

### Требования
- Go 1.26+
- nginx
- systemd

### Быстрый старт

```bash
sudo ./deploy.sh --default true
```

### Кастомный деплой

```bash
sudo ./deploy.sh \
  --deploy-root ~/w2g \
  --port 8080 \
  --log-file /var/log/w2g/w2g.log \
  --storage /var/lib/w2g/storage
```

### Управление сервисом

```bash
systemctl status w2g   # Статус
journalctl -u w2g -f   # Логи
sudo systemctl restart w2g  # Перезапуск
sudo systemctl stop w2g     # Остановка
```

## Проверка

```bash
curl http://localhost:8080/healthz
```
Ответ: `ok`

## Тесты

```bash
go test ./internal/... -v -count=1
```

## Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|-------------|----------|
| `PORT` | `8080` | Порт сервера |
| `STORAGE_DIR` | `./storage/` | Директория CSV-файлов |
| `LOG_LEVEL` | `debug` | Уровень логирования |
| `LOG_FILE` | — | Путь к лог-файлу |
| `SESSIONS_CLEANUP_INTERVAL` | `300` | Интервал очистки сессий (сек) |
| `MAX_ROOMS_PER_USER` | `10` | Макс. комнат на пользователя |
