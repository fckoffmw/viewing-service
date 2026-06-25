# Tasks

## Backend

### TODO

- [ ] full rest доделать для всех видов ресурсов
- [ ] CSV `rowsTo[T]` uses reflection** — `repo/repo.go:388-446`: fragile, slow, type-unsafe CSV deserialization via `reflect`. Already in tasks.md but still high priority.
- [ ] sync
  - [ ] чтобы страница обновлялась у всех при изменении источника
  - [ ] ws взаимодействие
  - [ ] nats
- [ ] админка
- [ ] разделение источников помимо VK
- [ ] кеширование
- [ ] auth more +

### DONE

- [x] побольше логирования с разделением по уровням
- [x] rooms: docs, review, tests
- [x] много комнат
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

## Client-side

### TODO

#### index.html

- [ ] пофиксить баг: при введении кода несуществующей комнаты в нее нельзя войти
- [ ] весь js код в файлы

#### content.html

- [ ] доделать функции editSource и deleteSource

### DONE

- [x] `web/demo/` сохранён (player-api-demo.html) — не трогать.
- [x] bug fix: deleteRoom btn on main page
- [x] убрать var контейнеры
- [x] переименовать глобальный var контейнер в roomContainer
