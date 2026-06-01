Отлично, всё что нужно есть. Пишу спецификацию.---

# Спецификация: синхронизация просмотра

## 1. Выбор стека и обоснование

### Почему без NATS/Redis

Для личного проекта с 2–20 участниками в комнате внешний брокер — это overengineering с реальными издержками: отдельный процесс, конфигурация, мониторинг, потенциальная точка отказа. Всё это ради задачи, которую уже решает существующий `RoomHub`.

`RoomHub` — это уже брокер событий внутри одного процесса: goroutine с channel dispatch, регистрация/дерегистрация клиентов, broadcast. Синхронизация просмотра — это просто новые типы сообщений в уже работающем пайплайне.

**Итоговый стек: чистый Go + WebSocket (gorilla/websocket, уже в зависимостях).** Никаких новых зависимостей.

---

## 2. Архитектурные решения

### 2.1 Авторитетный источник состояния — сервер

Клиент посылает **намерение** (`play`, `pause`, `seek`). Сервер обновляет `RoomState` и рассылает **факт** всем, включая отправителя. Это устраняет рассинхрон между клиентами: истина одна, она на сервере.

```
Client A: "play at 42.5s" → RoomHub → обновляет state → broadcast "playing=true, pos=42.5" → все клиенты
```

Отправителю тоже возвращается событие — это подтверждение, что его действие принято, и унифицирует логику на фронте (один обработчик для всех).

### 2.2 Новый клиент получает snapshot

При подключении клиент немедленно получает `sync` с текущим `RoomState`. Если видео играло, клиент вычисляет приблизительную позицию: `state.position + (now - state.updatedAt).seconds()` и делает `seek` + `play` локально.

### 2.3 Права — любой участник

Все три события (`play`, `pause`, `seek`) доступны любому участнику комнаты. Проверка: пользователь должен быть аутентифицирован и находиться в данной комнате (WebSocket соединение установлено).

### 2.4 Расхождение позиции — открытый вопрос

Оставляем на будущее. Возможные подходы: периодический `heartbeat` с позицией от каждого клиента + автоматический `resync` при отклонении > N секунд. Сейчас — заглушка.

---

## 3. Модель данных

### RoomState (расширение существующей структуры)

```go
type RoomState struct {
    SourceID  string
    SourceURL string

    // Новые поля синхронизации
    Playing   bool
    Position  float64   // секунды
    UpdatedAt time.Time // когда последний раз менялось состояние
    UpdatedBy string    // userID, кто последний изменил
}
```

### WebSocket-сообщения

**Клиент → сервер** (новые типы, дополняют существующий `chat`):

```json
{ "type": "play",  "payload": { "position": 42.5 } }
{ "type": "pause", "payload": { "position": 42.5 } }
{ "type": "seek",  "payload": { "position": 120.0 } }
```

**Сервер → клиент** (broadcast всем в комнате, включая отправителя):

```json
{ "type": "play",  "username": "alice", "payload": { "position": 42.5 } }
{ "type": "pause", "username": "alice", "payload": { "position": 42.5 } }
{ "type": "seek",  "username": "alice", "payload": { "position": 120.0 } }
```

**Сервер → новый клиент** (только ему при подключении):

```json
{
  "type": "sync",
  "payload": {
    "source_id":  "1",
    "source_url": "https://vk.com/...",
    "playing":    true,
    "position":   87.3,
    "updated_at": "2025-01-01T00:00:42Z"
  }
}
```

---

## 4. Серверная логика (Go)

### 4.1 Изменения в `chat/message.go`

```go
type IncomingMessage struct {
    Type    string          `json:"type"`    // "chat" | "play" | "pause" | "seek"
    Payload json.RawMessage `json:"payload"`
}

type PlayPayload  struct { Position float64 `json:"position"` }
type PausePayload struct { Position float64 `json:"position"` }
type SeekPayload  struct { Position float64 `json:"position"` }

type OutgoingMessage struct {
    Type      string      `json:"type"`
    Username  string      `json:"username,omitempty"`
    Timestamp time.Time   `json:"timestamp,omitempty"`
    Payload   interface{} `json:"payload,omitempty"`
}
```

### 4.2 Изменения в `chat/hub.go`

В `RoomHub.Run()` добавляется обработка новых типов сообщений:

```go
case msg := <-h.incoming:
    switch msg.Type {
    case "play":
        var p PlayPayload
        json.Unmarshal(msg.RawPayload, &p)
        h.state.Playing = true
        h.state.Position = p.Position
        h.state.UpdatedAt = time.Now()
        h.state.UpdatedBy = msg.SenderID
        h.broadcastAll(OutgoingMessage{
            Type:      "play",
            Username:  msg.SenderUsername,
            Timestamp: time.Now(),
            Payload:   p,
        })

    case "pause":
        // аналогично, Playing = false

    case "seek":
        // аналогично, Position = p.Position, Playing не меняется
    }
```

Ключевое решение: `broadcastAll` рассылает **всем**, включая отправителя. Это упрощает фронтенд — не нужно локально применять своё действие, оно придёт с сервера как подтверждение.

### 4.3 Подключение нового клиента — snapshot

В `chat/hub.go`, при `register`:

```go
case client := <-h.register:
    h.clients[client] = true
    // Немедленно отправить текущий state новому клиенту
    client.send <- OutgoingMessage{
        Type: "sync",
        Payload: SyncPayload{
            SourceID:  h.state.SourceID,
            SourceURL: h.state.SourceURL,
            Playing:   h.state.Playing,
            Position:  h.state.Position,
            UpdatedAt: h.state.UpdatedAt,
        },
    }
    // broadcast "join" остальным (уже есть)
```

---

## 5. Клиентская логика (JS + VK iframe API)

### 5.1 Инициализация

```javascript
const player = VK.VideoPlayer(iframe);
let isSyncing = false; // флаг: применяем серверную команду

player.on('started', () => { /* не слать ничего, это наш seek при sync */ });
```

### 5.2 Обработка входящих событий с сервера

```javascript
ws.onmessage = (e) => {
  const msg = JSON.parse(e.data);

  switch (msg.type) {
    case 'sync':
      isSyncing = true;
      player.seek(estimatedPosition(msg.payload));
      if (msg.payload.playing) player.play();
      else player.pause();
      isSyncing = false;
      break;

    case 'play':
      isSyncing = true;
      player.seek(msg.payload.position);
      player.play();
      isSyncing = false;
      break;

    case 'pause':
      isSyncing = true;
      player.pause();
      isSyncing = false;
      break;

    case 'seek':
      isSyncing = true;
      player.seek(msg.payload.position);
      isSyncing = false;
      break;
  }
};

// Учёт сетевого времени для sync snapshot
function estimatedPosition(payload) {
  const elapsed = (Date.now() - new Date(payload.updated_at)) / 1000;
  return payload.playing
    ? payload.position + elapsed
    : payload.position;
}
```

### 5.3 Отправка событий пользователем

```javascript
// Вешаем UI-кнопки, НЕ события плеера (иначе петля: server→player→event→server)
document.getElementById('btn-play').onclick = () => {
  ws.send(JSON.stringify({ type: 'play', payload: { position: player.getCurrentTime() } }));
};

document.getElementById('btn-pause').onclick = () => {
  ws.send(JSON.stringify({ type: 'pause', payload: { position: player.getCurrentTime() } }));
};

seekBar.oninput = debounce((e) => {
  ws.send(JSON.stringify({ type: 'seek', payload: { position: parseFloat(e.target.value) } }));
}, 300);
```

Важно: пользователь жмёт **кнопку** → событие уходит на сервер → сервер возвращает broadcast → `isSyncing = true` → плеер управляется. Плеер не является источником событий — только приёмником. Это предотвращает event loop.

---

## 6. Изменения в API

Новых HTTP-эндпоинтов нет. Все события проходят через существующий WebSocket `/ws/{invite_code}`.

`GET /api/rooms/{invite_code}` — поле `current_source` уже есть. Добавляем состояние воспроизведения для отображения на странице комнаты:

```json
{
  "current_source": { "id": "1", "name": "...", "url": "..." },
  "playback": {
    "playing":    true,
    "position":   87.3,
    "updated_at": "2025-01-01T00:00:42Z"
  }
}
```

Поле `playback` присутствует только если `RoomHub` активен (есть участники).

---

## 7. План разработки

| Этап | Задача | Файлы |
|------|--------|-------|
| 1 | Типизированные WS-сообщения | `chat/message.go` |
| 2 | Обработка `play/pause/seek` в RoomHub + обновление RoomState | `chat/hub.go` |
| 3 | Snapshot при подключении нового клиента | `chat/hub.go` |
| 4 | Обновление HTTP-ответа `/api/rooms/{code}` с полем `playback` | `room/handler.go` |
| 5 | Фронт: обработчик входящих sync-событий + `isSyncing` флаг | `frontend/room.js` |
| 6 | Фронт: кнопки play/pause/seek → WS | `frontend/room.js` |
| 7 | Фронт: `estimatedPosition` при `sync` snapshot | `frontend/room.js` |
| 8 | Ручное тестирование: 2+ вкладки, проверить все события | — |

---

## 8. Открытые вопросы (backlog)

| Вопрос | Варианты решения |
|--------|-----------------|
| Расхождение позиции при лагах | Периодический `heartbeat` с позицией → автоmatический `resync` при отклонении > 5s |
| Поддержка других источников (YT, RuTube) | Абстракция `PlayerAdapter` с методами `play/pause/seek/getPosition` |
| Конфликт команд (два пользователя одновременно) | Last-write-wins (текущий подход) — достаточно для 2–20 чел. |

---

Спецификация намеренно не вводит новых зависимостей и сохраняет философию проекта (простота, in-memory, CSV). Весь новый код — это расширение уже существующих структур `RoomHub` и `RoomState`, плюс клиентский JS поверх VK iframe API, который в проекте уже задокументирован.