# w2g (watch together)

Сервис для совместного просмотра видеоконтента с чатом.

## Функциональность

- Iframe-плеер (VK)
- Чат через WebSocket (2 участника)
- Аутентификация (session-based)

## Стек

| Компонент | Технология |
|-----------|------------|
| Backend | Go |
| Frontend | HTML, CSS, JS (vanilla) |
| Хранилище | CSV файлы |
| Сессии | in-memory |

## Модели данных

### User
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string | Порядковый номер |
| `username` | string | Логин |
| `password_hash` | string | bcrypt |
| `created_at` | time | Время создания |

### Source
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string | Порядковый номер |
| `name` | string | Название |
| `url` | string | URL iframe |

### Room
| Поле | Тип | Описание |
|------|-----|----------|
| `id` | string | ID комнаты |
| `source_id` | string | ID источника |

### Session
| Поле | Тип | Описание |
|------|-----|----------|
| `session_id` | string | 32 bytes base64 |
| `user_id` | string | FK → User.ID |
| `created_at` | time | Создание |
| `last_seen_at` | time | Последний запрос |
| `expires_at` | time | Истечение |

## Требования (TODO)

- Синхронизация просмотра
- Профиль пользователя
- Поддержка других источников (YT, RUT)

## API

См. [api.md](api.md)

## Аутентификация

См. [auth.md](auth.md)