# Tasks

## Backend

### TODO

- [ ] sync — синхронизация воспроизведения
  - [ ] изучить API источников (VK, YT, RUT)
  - [ ] WebSocket взаимодействие
  - [ ] обобщить механизм синхронизации
- [ ] auth — аутентификация
  - [ ] страницы login/reg
  - [ ] структура хранилища
  - [ ] API
  - [ ] сессии
- [ ] deploy script
- [ ] мультирум (сейчас только 1 глобальная комната)

### DONE

- [x] config, logging
- [x] CSV storage layer + тесты
- [x] Room entity + API (`/api/room`, `/api/room/source`)
- [x] Sources API (`GET /api/sources`, `POST /api/sources`)
- [x] AddSource handler + service тесты
- [x] WebSocket чат (Hub + Client, max 2)
- [x] Content selection page
- [x] Handler/service/repo разделение слоёв
- [x] Thread safety в repo
- [x] Docker Compose конфигурация

## Frontend

### TODO

- [ ] адаптивность страницы просмотра
  - [ ] проверить на мобильных устройствах
  - [ ] поле ввода на всю ширину чата
  - [ ] видео ~40% высоты, чат остальное

### DONE

- [x] Landing page (index.html)
- [x] Room page (room.html) — плеер + чат
- [x] Content selection page (content.html)
- [x] Profile page mockup (profile.html)
- [x] CSS для всех страниц