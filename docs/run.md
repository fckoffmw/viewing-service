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
# Статус
systemctl status w2g

# Логи
journalctl -u w2g -f

# Перезапуск
sudo systemctl restart w2g

# Остановка
sudo systemctl stop w2g
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
- Systemd: `sudo systemctl stop w2g`

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