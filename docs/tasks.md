# Tasks

## Backend

### TODO


- [ ] много комнат
- [ ] побольше логирования с разделением по уровням
- [ ] sync
  - [ ] ws взаимодействие
  - [ ] nats
- [ ] разделение источников помимо VK
- [ ] CSV storage — убрать reflect, использовать ручной маппинг или кодогенерацию
- [ ] graceful shutdown для Hub
- [ ] CORS middleware
- [ ] кеширование
- [ ] auth more +
- [ ] CSV storage — убрать reflect, использовать ручной маппинг или кодогенерацию

#### rooms feature

- [ ] csv storage в room нужно переименовать в repo
- [ ] не нравится что id, err := s.csvStorage.AddRoom(*room) указатель разыменовываем (решить надо как передавать структуры в репо)
- [ ] перенести GetAllRooms AddRoom UpdateRoom DeleteRoom выше чем локальные методы
- [ ] не делать составные имена по типу hub_manager
- [ ] дублрирование extractInviteCode убрать
- [ ] убрать логику старого хаба

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