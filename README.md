# w2g

Сервис для совместного просмотра видеоконтента с чатом.

## Документация

| Файл | Описание |
|------|----------|
| [docs/spec.md](docs/spec.md) | Спецификация проекта |
| [docs/arch.md](docs/arch.md) | Архитектура |
| [docs/run.md](docs/run.md) | Запуск, проверка, остановка |
| [docs/api.md](docs/api.md) | REST API endpoints |
| [docs/vk-iframe-api.md](docs/vk-iframe-api.md) | VK Video iframe JavaScript API |
| [docs/tasks.md](docs/tasks.md) | TODO и выполненные задачи |

## Текущее состояние

Сервис работает в режиме: одна общая комната, чат на двоих через WebSocket, выбор видео из каталога. Регистрация и авторизация — не реализованы.

## Release Notes

| Commit | Описание |
|--------|----------|
| `087d2c3` | Repo layer refactoring, добавлены тесты, thread safety |
| `92decc0` | CSS для profile page, тестовый контент |
| `b99e16f` | Исправлен .gitignore: web директория доступна |
| `934cea5` | Storage вынесен из Docker image на host |
| `63ae4bf` | Content selection feature, обновлено API |
| `1f7cf9a` | Добавлена сущность Room, endpoints `/api/room`, `/api/sources` |
| `59d40b7` | Room API: getAll sources |
| `fd91005` | CSS для sidebar |
| `9c37597` | Profile page: user info, секции (watch later, reviews, history) |
| `2cba0e2` | Разделение на handler/service/repo слои |
| `682d047` | Базовый CSS |
| `34903a2` | Начата работа над profile page |
| `6d4c864` | Обновлены требования к проекту |
| `8ad2fde` | Initial backend |

## Быстрый старт

```bash
go run ./cmd/w2g
```

Сервис доступен на `http://localhost:8080`

Полные инструкции — в [docs/run.md](docs/run.md)
