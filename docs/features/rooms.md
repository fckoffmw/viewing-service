# Мультикомнатность — дизайн-документ

> Дизайн-документ: принятые решения на момент разработки фичи.
> Текущее поведение см. в [api.md](../api.md), [arch.md](../arch.md), [ws-protocol.md](ws-protocol.md).

## Важно:

Упор фичи - мультикомнаты. Пока что БЕЗ синхронизации. Все что касается синхронизации - заглушки на будущее

## Анализ: In-Memory vs Persistent

### Почему persistent был выбран изначально

Главный аргумент — комнаты переживают рестарт. Но давай честно оценим насколько это критично для w2g:

- Сервис деплоится через systemd, рестарты редки и контролируемы
- Комната без участников бессмысленна — всё равно нужно собраться заново
- Ссылка-инвайт может быть статичной (зашита в invite_code) и пережить рестарт даже при in-memory, если код детерминирован или заранее известен пользователю

### Честные минусы in-memory для комнат

| Минус | Критичность для w2g |
|---|---|
| Потеря комнат при рестарте | Средняя — пересоздать комнату 30 секунд |
| Потеря invite_code при рестарте | Высокая — сохранённая ссылка перестаёт работать |
| Нет истории (кто создавал) | Низкая — не заявлено в требованиях |

**Ключевой вопрос был именно в invite_code** — если ссылка умирает при рестарте, UX сломан. Это и тянуло к persistent.

### Решение: Гибрид

**Invite_code персистентен (CSV/файл), всё остальное in-memory.**

Один лёгкий файл `rooms.csv` хранит только `id, name, owner_id, invite_code, created_at` — никакого SQLite, никаких миграций. При старте сервис загружает комнаты в память. Участники, состояние воспроизведения, онлайн — всё in-memory.

Это соответствует текущей философии проекта (CSV storage) и даёт стабильные ссылки.

---

# Спецификация: Мультикомнатность — In-Memory редакция

## 1. Хранилище

```
CSV (rooms.csv)          ← только метаданные комнаты, персистентно
In-Memory (HubManager)   ← участники, состояние, чат
```

### rooms.csv

```
id,name,owner_id,invite_code,created_at
uuid,Movie Night,user-uuid,X7K2PQ4M,2025-01-01T00:00:00Z
```

**При старте:** загрузить все комнаты из CSV в `map[inviteCode]*Room`.
**При создании:** записать в CSV + добавить в map.
**При удалении:** удалить из CSV + удалить из map + закрыть RoomHub если активен.

Участники (`RoomMember`) — **не персистируются**. Только in-memory на время сессии.

---

## 2. Модель данных

```go
// Персистентная часть (CSV)
type Room struct {
    ID         string
    Name       string
    OwnerID    string
    InviteCode string    // 8 символов, [A-Z0-9], crypto/rand
    CreatedAt  time.Time
}

// In-memory часть (живёт в RoomHub)
type RoomState struct {
    SourceID  string
    SourceURL string
    Playing   bool
    Position  float64
    UpdatedAt time.Time
    UpdatedBy string
}

type RoomMember struct {
    UserID    string
    Username  string
    JoinedAt  time.Time
}
```

---

## 3. Архитектура In-Memory Hub

```
RoomStore (in-memory + CSV sync)
└── map[inviteCode]*Room

HubManager (singleton, in-memory)
└── map[roomID]*RoomHub

RoomHub
├── room      *Room
├── state     RoomState
├── clients   map[*Client]bool
├── broadcast chan Message
├── register  chan *Client
└── unregister chan *Client
```

### Жизненный цикл RoomHub

```
Первый клиент подключается → HubManager.GetOrCreate(roomID) → новый RoomHub → go hub.Run()
Последний клиент уходит   → HubManager авто-удаляет RoomHub → GC
Комната удаляется owner'ом → закрыть RoomHub + удалить из CSV
```

`RoomHub` создаётся лениво — нет участников, нет аллокаций. Это важно: 100 комнат в CSV не означают 100 горутин.

---

## 4. REST API

```
POST   /api/rooms                         — создать комнату (auth required)
GET    /api/rooms/{invite_code}           — получить метаданные комнаты
DELETE /api/rooms/{invite_code}           — удалить (owner only)
PATCH  /api/rooms/{invite_code}/source    — сменить источник (owner only)
```

### POST /api/rooms

Request:
```json
{ "name": "Movie Night" }
```

Response 201:
```json
{
  "id": "uuid",
  "name": "Movie Night",
  "invite_code": "X7K2PQ4M",
  "invite_url": "/room/X7K2PQ4M",
  "owner_id": "user-uuid",
  "created_at": "2025-01-01T00:00:00Z"
}
```

### GET /api/rooms/{invite_code}

Response 200:
```json
{
  "id": "uuid",
  "name": "Movie Night",
  "invite_code": "X7K2PQ4M",
  "owner_id": "user-uuid",
  "members_online": 3,
  "current_source": { "id": "1", "name": "...", "url": "..." },
  "created_at": "..."
}
```

`members_online` — берётся из `HubManager` если `RoomHub` активен, иначе 0.
`current_source` — берётся из `RoomHub.state` если активен, иначе null.

### PATCH /api/rooms/{invite_code}/source

Owner only. Обновляет `RoomHub.state.SourceID` + рассылает `source_changed` WS-событие всем участникам.

```json
{ "source_id": "1" }
```

Response 200:
```json
{ "source_id": "1" }
```

---

## 5. WebSocket

### Endpoint

```
GET /ws/{invite_code}
```

Флоу подключения:
1. Проверить `session_id` cookie → получить `userID`, `username`
2. Найти комнату по `invite_code` в `RoomStore`
3. `HubManager.GetOrCreate(room.ID)` → получить `RoomHub`
4. Зарегистрировать клиента в хабе
5. Отправить новому клиенту `sync` с текущим `RoomState`
6. Broadcast `join` остальным участникам

### Типы сообщений

Клиент → сервер:
```json
{ "type": "chat",  "payload": { "text": "Привет!" } }
{ "type": "play",  "payload": { "position": 42.5 } }
{ "type": "pause", "payload": { "position": 42.5 } }
{ "type": "seek",  "payload": { "position": 120.0 } }
```

Сервер → клиент:
```json
{ "type": "chat",           "username": "user123", "timestamp": "...", "payload": { "text": "Привет!" } }
{ "type": "join",           "username": "user123", "timestamp": "..." }
{ "type": "leave",          "username": "user123", "timestamp": "..." }
{ "type": "sync",           "payload": { "source_id": "1", "source_url": "...", "playing": true, "position": 87.3 } }
{ "type": "play",           "username": "user123", "payload": { "position": 42.5 } }
{ "type": "pause",          "username": "user123", "payload": { "position": 42.5 } }
{ "type": "seek",           "username": "user123", "payload": { "position": 120.0 } }
{ "type": "source_changed", "payload": { "source_id": "2", "source_url": "..." } }
```

### Права на WS-события

| Событие | Кто может слать |
|---|---|
| `chat` | Все участники |
| `play`, `pause`, `seek` | Owner only (проверка на сервере) |

Если не-owner шлёт `play/pause/seek` — сервер молча игнорирует (no error, no relay).

---

## 6. Изменения в архитектуре проекта

```
internal/
├── room/
│   ├── model.go      — Room, RoomState, RoomMember
│   ├── store.go      — RoomStore: CSV load/save + in-memory map
│   ├── service.go    — Create, GetByInviteCode, Delete, PatchSource
│   ├── handler.go    — HTTP хендлеры
│   └── *_test.go
├── chat/
│   ├── manager.go    — HubManager (новый файл)
│   ├── hub.go        — рефакторинг: RoomHub вместо глобального Hub
│   ├── client.go     — добавить roomID, userID, isOwner
│   ├── message.go    — типизированные сообщения (новый файл)
│   └── *_test.go
└── http/
    └── router.go     — новые роуты, /ws/{invite_code}
```

`repo` пакет (CSV) — `RoomStore` реализует аналогичный паттерн самостоятельно внутри `room` пакета, не расширяя общий `repo` (у комнат своя логика in-memory слоя).

---

## 7. Конфигурация

```env
MAX_ROOMS_PER_USER=10   # защита от спама, 0 = unlimited
```

`MAX_ROOM_MEMBERS` не нужен — убираем хардкод лимита в 2 клиента из `chat`.

---

## 8. Порядок разработки

| Этап | Задача |
|---|---|
| 1 | `room` пакет: модели, `RoomStore` (CSV + in-memory map), сервис, хендлеры |
| 2 | Роутер: `/api/rooms/*`, депрекация `/api/room` |
| 3 | `chat`: `message.go` — типизированные сообщения |
| 4 | `chat`: рефакторинг `Hub` → `RoomHub` + `HubManager` |
| 5 | `client.go`: добавить `roomID`, `userID`, `isOwner` |
| 6 | `/ws/{invite_code}` — новый endpoint, флоу подключения |
| 7 | Синхронизация: `RoomState`, `play/pause/seek/sync` |
| 8 | Фронтенд |