# Tasks

## Backend

### TODO

- [] sync
  - [] ws взаимодействие
  - [] nats
  - [] изучить апи разных источников *
  - [] обобщить механизм синхронизации *
- [] auth
  - [] страницы login/reg
  - [] структура хранилища
  - [] апи
  - [] сессии
- [] deploy script
- [] разделение источников помимо VK
  - [] изучить возможные основные iframe источники (VK, YT, RUT)
  - [] спланировать разделение подсистем работы с источниками
  - [] реализация

### DONE

- [x] улучшение конфигурации
- [x] config
- [x] logging
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

- [] адаптивность страницы просмотра
  - [] проверить на реальных моб. устройствах (Chrome, Firefox)
  - [] поле ввода должно занимать всю ширину чата
  - [] видео ~40% высоты, чат остальное
- [] страница выбора

### DONE

- [x] Landing page (index.html)
- [x] Room page (room.html) — плеер + чат
- [x] Content selection page (content.html)
- [x] Profile page mockup (profile.html)
- [x] CSS для всех страниц
