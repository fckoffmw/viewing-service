# Аутентификация

## Стек

- Сессии: in-memory (`sync.RWMutex`)
- Хранение пользователей: CSV
- Пароли: bcrypt (cost=12)
- Cookie: httpOnly, Path=/, Max-Age=604800 (7 дней)

## Потоки

### Регистрация
1. Валидация: username (≥3), password (≥4), уникальность
2. bcrypt hash пароля
3. Создание сессии
4. Set-Cookie

### Логин
1. Поиск пользователя
2. bcrypt verify
3. Создание сессии
4. Set-Cookie

### Защищённый запрос
1. Проверка cookie
2. Проверка сессии в store
3. Sliding expiry
4. Пропуск или 401

## Endpoints

| Метод | Путь | Требует auth |
|-------|-----|-------------|
| POST | `/auth/register` | Нет |
| POST | `/auth/login` | Нет |
| POST | `/auth/logout` | Нет |
| GET | `/auth/me` | Да |

## Валидация

### Запуск
```bash
go run ./cmd/w2g
```

### Регистрация
```bash
curl -v -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'
```
Ожидается: `201` + Set-Cookie

### Логин
```bash
curl -v -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}'
```
Ожидается: `200` + Set-Cookie

### Защищённый endpoint (без cookie)
```bash
curl http://localhost:8080/api/sources
```
Ожидается: `401` + `{"error":"session not found"}`

### Защищённый endpoint (с cookie)
```bash
# Login and save cookie to file
curl -v -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"testpass123"}' \
  -c cookies.txt

# Use cookie for protected request
curl http://localhost:8080/api/sources -b cookies.txt
```
Ожидается: `200` + `[]`

### Валидация полей

Пустой username:
```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"","password":"1234"}'
```
Ответ: `400` + `{"error":"username cannot be empty"}`

Короткий username (<3):
```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"ab","password":"1234"}'
```
Ответ: `400` + `{"error":"username must be at least 3 characters"}`

Короткий password (<4):
```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"123"}'
```
Ответ: `400` + `{"error":"password must be at least 4 characters"}`

Неверный пароль:
```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","password":"wrong"}'
```
Ответ: `401` + `{"error":"invalid credentials"}`

## Зависимости

```
golang.org/x/crypto   // bcrypt
```