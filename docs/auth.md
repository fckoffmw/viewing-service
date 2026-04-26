Вот полная архитектурная спецификация. (Claude gen)

> **Статус:** in-memory сессии реализованы (register, login — stub). Redis — не реализовано.

---

## Архитектурные решения

### 1. Общая архитектура

**Хранилище сессий — in-memory (sync.Map) с персистентностью через CSV**

Для вашего кейса (VPS, один инстанс, Go, CSV-based storage) Redis — избыточен. JWT — исключён по требованию. Оптимальный выбор: in-memory store с опциональным dump в файл при рестарте.

Сессия — это непрозрачный токен (криптографически случайный, 32 байта, base64url), который хранится на сервере и привязан к пользователю. Клиент получает его через httpOnly cookie.

**Создание:** `POST /auth/login` → сервер валидирует credentials → генерирует session ID → сохраняет в map → ставит cookie.

**Проверка:** каждый запрос → middleware читает cookie → ищет session ID в map → если найден и не истёк → пропускает, иначе 401.

---

### 2. Потоки регистрации и логина

**Регистрация:**
1. `POST /auth/register` принимает `{username, password}`
2. Валидация: username уникальность (поиск по `users.csv`), минимальная длина
3. Хэширование пароля через `bcrypt` (cost=12)
4. Запись в `users.csv`: `id,username,password_hash,created_at`
5. Автоматический логин: создаётся сессия, ставится cookie
6. Ответ: `201 Created + Set-Cookie`

**Логин:**
1. `POST /auth/login` принимает `{username, password}`
2. Поиск пользователя по username в CSV
3. `bcrypt.CompareHashAndPassword(hash, password)`
4. При успехе: `crypto/rand` генерирует 32 байта → base64url → session ID
5. `SessionStore.Set(sessionID, Session{UserID, CreatedAt, LastSeenAt, ExpiresAt})`
6. `Set-Cookie: session_id=...; HttpOnly; Secure; SameSite=Lax; Path=/; Max-Age=604800`

---

### 3. Структуры данных

**User (users.csv):**
```
id          string    // UUID v4
username    string    // уникальный, lowercase
password    string    // bcrypt hash ($2a$12$...)
created_at  time.Time
```

**Session (in-memory):**
```
session_id   string    // 32 bytes, base64url (ключ map)
user_id      string    // FK → User.ID
created_at   time.Time
last_seen_at time.Time // обновляется при каждом запросе
expires_at   time.Time // created_at + 7 дней (sliding или fixed)
```

**SessionStore:**
```go
type SessionStore struct {
    mu       sync.RWMutex
    sessions map[string]*Session
}
```

---

### 4. Logout, истечение, несколько устройств

**Logout:** `POST /auth/logout` → удаляет session_id из map → `Set-Cookie: session_id=; Max-Age=0`. Сессия исчезает мгновенно.

**Истечение:** фоновая горутина каждые N минут проходит по map и удаляет сессии с `expires_at < now()`. Дополнительно middleware проверяет `expires_at` при каждом запросе — expired сессия = 401 + clear cookie.

**Sliding vs Fixed expiry:** рекомендую sliding (обновляем `expires_at` при каждом запросе, пока пользователь активен). Для `watch2gether` это логично — пользователь смотрит видео часами.

**Несколько устройств:** один пользователь → несколько записей в map (разные session_id). Logout на одном устройстве удаляет только текущую сессию. `POST /auth/logout-all` → фильтрует map по `user_id`, удаляет все.

---

### 5. Безопасность

**Cookie flags:**
- `HttpOnly` — JS не может читать (XSS protection)
- `Secure` — только HTTPS (на VPS за Nginx — обязательно)
- `SameSite=Lax` — достаточно для большинства сценариев, защищает от CSRF при навигации
- `Path=/` — cookie отправляется на все endpoints

**CSRF:** при `SameSite=Lax` большинство атак заблокированы. Для WS-handshake дополнительно проверяйте `Origin` header в Nginx или в Go upgrader. Если нужны мутирующие POST-запросы из JS (fetch), — добавьте Double Submit Cookie или custom header `X-Requested-With: XMLHttpRequest` (достаточно для проверки что запрос не cross-site form submission).

**Хранение паролей:** `golang.org/x/crypto/bcrypt`, cost=12. Никакого MD5/SHA1. Никакого хранения plaintext.

**Session fixation:** генерировать новый session_id при каждом логине (не переиспользовать старый).

**Timing attacks:** сравнение паролей через bcrypt уже constant-time. Session lookup по map — O(1), не раскрывает информацию.

---

### 6. Альтернативы

| Вариант | За | Против | Вердикт |
|---|---|---|---|
| **In-memory sessions** | Просто, быстро, нет зависимостей | Теряются при рестарте | ✅ Выбран |
| **Redis sessions** | Персистентно, масштабируется | Лишняя зависимость для 1 VPS | При росте |
| **JWT** | Stateless, масштабируемо | Исключён требованием, сложнее logout | ❌ |
| **DB sessions (CSV)** | Персистентно | Медленно, CSV не для concurrent write | ❌ |

Для одного VPS с ожидаемой нагрузкой в десятки-сотни пользователей — in-memory оптимально. При росте до нескольких инстансов — вставляется Redis без изменения интерфейса (меняется только реализация `SessionStore`).

---

Теперь финальная спецификация для реализации:---

## Финальная спецификация

### Overview

Session-based аутентификация без JWT для Go-сервиса. Сессии хранятся in-memory (`sync.RWMutex` + `map`). Cookie — единственный транспорт для session ID. Пароли — bcrypt. Middleware перехватывает все защищённые маршруты.

---

### Components

**`internal/auth/model.go`**
```go
type User struct {
    ID           string
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}

type Session struct {
    SessionID   string
    UserID      string
    CreatedAt   time.Time
    LastSeenAt  time.Time
    ExpiresAt   time.Time
}
```

**`internal/auth/store.go`** — SessionStore
```go
type SessionStore struct {
    mu       sync.RWMutex
    sessions map[string]*Session
}

func (s *SessionStore) Set(sess *Session)
func (s *SessionStore) Get(id string) (*Session, bool)
func (s *SessionStore) Delete(id string)
func (s *SessionStore) DeleteByUserID(userID string)  // logout-all
func (s *SessionStore) Cleanup()                       // вызывается тикером
```

**`internal/auth/service.go`**
```go
func (s *Service) Register(username, password string) (*Session, error)
func (s *Service) Login(username, password string) (*Session, error)
func (s *Service) Logout(sessionID string)
func (s *Service) LogoutAll(userID string)
func (s *Service) ValidateSession(sessionID string) (*Session, error)
```

**`internal/middleware/auth.go`** — извлекает cookie, вызывает `ValidateSession`, кладёт `userID` в контекст. Обновляет `LastSeenAt` + `ExpiresAt` (sliding).

---

### API Endpoints

```
POST   /auth/register     body: {username, password}          → 201 + Set-Cookie
POST   /auth/login        body: {username, password}          → 200 + Set-Cookie
POST   /auth/logout       cookie required                     → 200 + clear cookie
POST   /auth/logout-all   cookie required                     → 200 + clear cookie
GET    /auth/me           cookie required                     → 200 {id, username}
```

Все остальные маршруты (`/room/*`, `/ws/*`) — за `AuthMiddleware`.

---

### Data Models

**`users.csv`:** `id,username,password_hash,created_at`

**`sessions` (in-memory только):** не персистируются (при рестарте пользователи логинятся заново — приемлемо).

Если персистентность нужна: сериализовать map в `sessions.csv` через `encoding/csv` при graceful shutdown (`os.Signal` → `SIGTERM`), загружать при старте (фильтруя expired).

---

### Flow diagrams (текст)

```
REGISTER:
Client → POST /auth/register
       → validate (unique username, len >= 3)
       → bcrypt.GenerateFromPassword(cost=12)
       → repo.CreateUser(user)
       → crypto/rand 32 bytes → base64url = sessionID
       → store.Set(session, TTL=7d)
       → Set-Cookie: session_id=... HttpOnly Secure SameSite=Lax
       → 201 Created

LOGIN:
Client → POST /auth/login
       → repo.GetUserByUsername(username)
       → bcrypt.CompareHashAndPassword
       → [fail] → 401 Unauthorized (одинаковое сообщение при wrong user/pass)
       → [ok]  → crypto/rand → sessionID → store.Set → Set-Cookie → 200

PROTECTED REQUEST:
Client (Cookie) → AuthMiddleware
               → cookie.Value = sessionID
               → store.Get(sessionID)
               → [not found / expired] → 401
               → [ok] → update LastSeenAt + ExpiresAt (sliding)
               → ctx = context.WithValue(ctx, "userID", session.UserID)
               → next handler

LOGOUT:
Client → POST /auth/logout → store.Delete(sessionID)
       → Set-Cookie: session_id=; Max-Age=0 → 200

CLEANUP (goroutine):
time.NewTicker(5 * time.Minute) → store.Cleanup()
→ range sessions: if ExpiresAt.Before(now()) → delete
```

---

### Edge Cases

**Неверный логин:** всегда отвечать одинаково — `"invalid credentials"`. Не раскрывать "пользователь не существует" vs "неверный пароль". Добавить rate limiting (простой: `sync.Map[ip]attempts` или middleware-счётчик).

**Истёкшая сессия в запросе:** middleware удаляет её из store + шлёт `401` + `Max-Age=0` в cookie. Клиент редиректит на `/login`.

**WS-соединение:** при WS upgrade cookie отправляется браузером автоматически. В Go upgrader проверять Origin: `upgrader.CheckOrigin = func(r *http.Request) bool { return isAllowedOrigin(r.Header.Get("Origin")) }`. Session validate вызывать до upgrade.

**Race condition:** `sync.RWMutex` покрывает все случаи — RLock для Get, Lock для Set/Delete/Cleanup.

**Пустой или невалидный cookie:** `http.ErrNoCookie` → 401. Невалидный base64 → 401. Не паниковать, логировать как debug.

**Logout при уже удалённой сессии:** idempotent — `store.Delete` на несуществующий ключ — no-op, ответ всё равно `200`.

**Несколько устройств — logout-all:** итерация по store с RLock, collect IDs with matching userID, затем Lock + delete batch.

---

### Зависимости

```
golang.org/x/crypto   // bcrypt
```

Всё остальное — stdlib: `crypto/rand`, `encoding/base64`, `sync`, `net/http`, `time`.

## Что меняется относительно предыдущей спецификации

In-memory `sync.Map` полностью заменяется Redis. Интерфейс `SessionStore` остаётся тем же — меняется только реализация. Остальная архитектура (bcrypt, cookie, middleware, endpoints) идентична.

---

## Почему Redis здесь уместен

| Свойство | In-memory | Redis |
|---|---|---|
| TTL сессий | ручной тикер | `EXPIRE` из коробки |
| Персистентность | нет (рестарт = logout all) | RDB/AOF |
| Несколько инстансов Go | ❌ | ✅ |
| Сложность деплоя | нет | +1 контейнер |
| Sliding expiry | вручную | `EXPIRE` при каждом Get |

Для вашего стека (`docker-compose` + VPS) Redis добавляется одной строкой в `docker-compose.yml`.

------

## Финальная спецификация (Redis-редакция)

### Overview

Session-based аутентификация с Redis как единственным хранилищем сессий. Go-сервис не держит сессии в памяти — всё делегируется Redis. TTL управляется через `EXPIRE`/`EXPIREAT`. Sliding expiry реализуется сбросом TTL при каждом валидном запросе. Пароли — bcrypt.

---

### Components

**`internal/auth/model.go`**
```go
type User struct {
    ID           string
    Username     string
    PasswordHash string
    CreatedAt    time.Time
}

// сериализуется в Redis как JSON или hash
type Session struct {
    SessionID string
    UserID    string
    CreatedAt time.Time
}
// ExpiresAt не хранится явно — TTL управляется Redis
```

**`internal/auth/store.go`** — RedisSessionStore
```go
type RedisSessionStore struct {
    client *redis.Client
    ttl    time.Duration // 7 * 24 * time.Hour
}

func (s *RedisSessionStore) Set(sess *Session) error
// SET sess:<sessionID> <json> EX 604800

func (s *RedisSessionStore) Get(id string) (*Session, error)
// GET sess:<sessionID>
// → если nil → ErrSessionNotFound
// → если ok → EXPIRE sess:<sessionID> 604800  (sliding)

func (s *RedisSessionStore) Delete(id string) error
// DEL sess:<sessionID>

func (s *RedisSessionStore) DeleteAllForUser(userID string) error
// SMEMBERS user_sessions:<userID> → DEL каждого + DEL user_sessions:<userID>
```

**Дополнительная структура в Redis для logout-all:**
```
user_sessions:<userID>  →  Set { sessionID1, sessionID2, ... }
```
При `Set()` → `SADD user_sessions:<userID> <sessionID>`.
При `Delete()` → `SREM user_sessions:<userID> <sessionID>`.

**`internal/auth/service.go`** — без изменений относительно предыдущей спецификации, работает через интерфейс `SessionStore`.

**`internal/middleware/auth.go`**
```go
// читает cookie → вызывает store.Get() → Get сам сбрасывает TTL (sliding)
// при ошибке → 401 + clear cookie
// при успехе → context.WithValue(ctx, CtxUserID, session.UserID)
```

---

### Redis ключи

```
sess:<sessionID>          STRING  JSON(Session)   TTL=7d (sliding)
user_sessions:<userID>    SET     {sessionID...}  TTL не ставим — очищаем вручную
```

Для `user_sessions` TTL не устанавливается намеренно: сессии в Set могут протухнуть в Redis раньше, поэтому при `DeleteAllForUser` нужно игнорировать `DEL` по уже-несуществующим ключам (Redis это делает молча).

---

### API Endpoints

```
POST  /auth/register     {username, password}  → 201 + Set-Cookie
POST  /auth/login        {username, password}  → 200 + Set-Cookie
POST  /auth/logout       cookie required       → 200 + clear cookie
POST  /auth/logout-all   cookie required       → 200 + clear cookie
GET   /auth/me           cookie required       → 200 {id, username}
```

---

### docker-compose изменения

```yaml
services:
  app:
    build: .
    environment:
      REDIS_ADDR: redis:6379
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    command: redis-server --save 60 1 --loglevel warning
    volumes:
      - redis_data:/data
    restart: unless-stopped

volumes:
  redis_data:
```

`--save 60 1` — RDB snapshot каждые 60 секунд если было хотя бы 1 изменение. Достаточно для сессий.

---

### Config

```go
// internal/config/config.go
type Config struct {
    // ... existing fields
    RedisAddr     string        // "redis:6379"
    RedisPassword string        // "" если без auth
    RedisDB       int           // 0
    SessionTTL    time.Duration // 7 * 24 * time.Hour
}
```

---

### Зависимости

```
github.com/redis/go-redis/v9
golang.org/x/crypto              // bcrypt
```

Инициализация клиента:
```go
rdb := redis.NewClient(&redis.Options{
    Addr:     cfg.RedisAddr,
    Password: cfg.RedisPassword,
    DB:       cfg.RedisDB,
})
// проверка при старте:
if err := rdb.Ping(ctx).Err(); err != nil {
    log.Fatal("redis unavailable:", err)
}
```

---

### Flow (текст)

```
SET session:
  key   = "sess:" + sessionID          (32 bytes, base64url)
  value = json.Marshal(Session{...})
  cmd   = SET key value EX 604800

GET session (+ sliding TTL):
  val = GET key
  if val == nil → ErrNotFound
  json.Unmarshal → Session
  EXPIRE key 604800                    // сброс TTL при каждом запросе

LOGOUT:
  DEL sess:<sessionID>
  SREM user_sessions:<userID> <sessionID>

LOGOUT-ALL:
  ids = SMEMBERS user_sessions:<userID>
  DEL sess:<id1> sess:<id2> ...        // один вызов, variadic
  DEL user_sessions:<userID>
```

---

### Edge Cases

**Redis недоступен:** `store.Get()` вернёт ошибку → middleware ответит `503 Service Unavailable` (не 401 — это инфраструктурная ошибка, не ошибка аутентификации). Логировать как `ERROR`.

**TTL истёк в Redis, но sessionID есть в `user_sessions` Set:** `GET` вернёт nil → middleware вернёт 401. `SREM` при следующем `Delete` — no-op. Set самоочистится при `logout-all` (DEL несуществующего ключа безопасен).

**WS-соединение:** валидация сессии до `upgrader.Upgrade()`. После upgrade сессия в контексте WS-хендлера не обновляется (нет HTTP-запросов) — это нормально. Если нужен re-check — отдельный тикер внутри WS-горутины каждые N минут.

**Конкурентные логины с одного аккаунта:** разрешены по умолчанию. Каждый логин создаёт новый `sess:*` ключ и добавляет его в `user_sessions`. `logout-all` убивает все.

**Replay атака с украденным session ID:** не решается на уровне сессий — это зона HTTPS + `Secure` cookie. Дополнительно можно хранить в Session `UserAgent` и `IP` и сверять при каждом запросе (опционально, усложняет UX при смене IP).