# w2g

Сервис для совместного просмотра видеоконтента с чатом.

## Документация

| Файл | Описание |
|------|----------|
| [docs/spec.md](docs/spec.md) | Спецификация |
| [docs/arch.md](docs/arch.md) | Архитектура |
| [docs/run.md](docs/run.md) | Запуск |
| [docs/api.md](docs/api.md) | REST API |
| [docs/auth.md](docs/auth.md) | Аутентификация |
| [docs/tasks.md](docs/tasks.md) | TODO |

## Текущее состояние

- Аутентификация: register, login, logout, /auth/me
- REST API: sources, rooms (мультикомнаты)
- WebSocket чат: /ws/{invite_code}
- Источники: CRUD через API
- Комнаты: несколько комнат с invite-кодами

## Быстрый старт

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

## Деплой

```bash
sudo ./deploy.sh --default true
```

Документация: [docs/run.md](docs/run.md)