# Tasks

## Backend

### TODO

- [ ] Auth
  - [ ] Регистрация
  - [ ] Логин / сессии (Redis)
  - [ ] Логаут
- [ ] Профиль
  - [ ] API для получения профиля
  - [ ] API для обновления профиля
- [ ] Смотреть позже
  - [ ] Модель данных
  - [ ] CRUD endpoints
- [ ] История просмотра
  - [ ] Модель данных
  - [ ] Запись при просмотре
  - [ ] API для получения
- [ ] Заметки
  - [ ] Модель данных
  - [ ] CRUD endpoints
- [ ] Комнаты
  - [ ] Перейти от глобальной комнаты к пользовательским
  - [ ] Создание комнат
  - [ ] Приглашения

### DONE

- [x] CSV storage layer с тестами
- [x] Room entity + API (`/api/room`, `/api/room/source`)
- [x] Sources API (`/api/sources`)
- [x] WebSocket chat (Hub + Client, max 2)
- [x] Content selection page
- [x] Handler/service/repo разделение слоёв
- [x] Thread safety в repo
- [x] Docker Compose конфигурация

## Frontend

### TODO

- [ ] Авторизация / регистрация формы
- [ ] Профиль с реальными данными
- [ ] Watch later UI
- [ ] История просмотра UI
- [ ] Заметки UI

### DONE

- [x] Landing page (index.html)
- [x] Room page (room.html) — плеер + чат
- [x] Content selection page (content.html)
- [x] Profile page mockup (profile.html)
- [x] CSS для всех страниц
