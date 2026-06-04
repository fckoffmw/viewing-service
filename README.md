# w2g

Сервис для совместного просмотра видеоконтента с чатом.

## Документация

| Файл | Описание |
|------|----------|
| [docs/spec.md](docs/spec.md) | Спецификация |
| [docs/arch.md](docs/arch.md) | Архитектура |
| [docs/run.md](docs/run.md) | Запуск |
| [docs/api.md](docs/api.md) | REST API |
| [docs/features/auth.md](docs/features/auth.md) | Аутентификация |
| [docs/features/ws-protocol.md](docs/features/ws-protocol.md) | WebSocket протокол |
| [docs/tasks.md](docs/tasks.md) | TODO |

## Текущее состояние

- Аутентификация: register, login, logout, /auth/me
- REST API: sources, rooms (мультикомнаты)
- WebSocket чат и синхронизация: /ws/{invite_code}
- Синхронизация просмотра (play/pause/seek)
- Источники: CRUD через API
- Комнаты: несколько комнат с invite-кодами

## Быстрый старт

```bash
task run
```

Сервис доступен на `http://localhost:8080`

## Деплой

```bash
sudo ./deploy.sh --default true
```

Документация: [docs/run.md](docs/run.md)
