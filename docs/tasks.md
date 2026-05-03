# Tasks

## Backend

### TODO

- [ ] review
  - [ ] CRITICAL: убрать глобальную map currentSourceID (room/service.go:49)
  - [ ] CRITICAL: добавить файловую блокировку для CSV
  - [ ] CRITICAL: добавить graceful shutdown для Hub
  - [ ] CRITICAL: валидация Source URL
  - [ ] важно: sessions persistence (сейчас in-memory)
  - [ ] важно: добавить rate limiting
  - [ ] важно: добавить ReadTimeout/WriteTimeout для http.Server
  - [ ] важно: использовать filepath.Join для путей
  - [ ] важно: добавить CORS middleware
- [ ] много комнат
- [ ] побольше логирования с разделением по уровням
- [ ] sync
  - [ ] ws взаимодействие
  - [ ] nats
- [ ] разделение источников помимо VK
- [ ] CSV storage — убрать reflect, использовать ручной маппинг или кодогенерацию
- [ ] кеширование
- [ ] auth more +

### DONE

- [x] simple deploy
- [x] разгрузить main
- [x] auth
- [x] config
- [x] logging
- [x] CSV storage layer с тестами
- [x] Room entity + API
- [x] Sources API
- [x] WebSocket chat (max 2)
- [x] Content selection page
- [x] handler/service/repo разделение слоёв
- [x] thread safety
- [x] Docker Compose
- [x] auth — register/login/logout/me endpoints
- [x] auth middleware
- [x] login/register pages

## Frontend

### TODO

- [ ] страница выбора контента — добавить проверку auth

### DONE

- [x] Landing page (index.html)
- [x] Room page (room.html) — плеер + чат
- [x] Content selection page (content.html)
- [x] Login page (login.html)
- [x] Register page (register.html)
- [x] CSS для всех страниц