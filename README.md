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
- Страницы: index, room, content, login, register
- Чат: WebSocket (max 2 клиента)
- Источники: CRUD через API
- Комната: одна глобальная

## Быстрый старт

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

Документация: [docs/run.md](docs/run.md)