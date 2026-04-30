# Tasks

## Backend

### TODO

- [ ] simple deploy
- [ ] CSV storage — убрать reflect, использовать ручной маппинг или кодогенерацию
- [ ] graceful shutdown для Hub
- [ ] CORS middleware
- [ ] sync
  - [ ] ws взаимодействие
  - [ ] nats
- [ ] много комнат
- [ ] разделение источников помимо VK
- [ ] auth more +

### DONE

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