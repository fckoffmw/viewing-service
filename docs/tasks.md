# Tasks

## Backend

### TODO

- [ ] sync
  - [ ] client side
  - [x] структура состояния
  - [x] типы возможных ивентов (сверяясь с sync.md)
  - [x] сервисный слой
  - [x] зачем ? extractInviteCodeFromPath(r.URL.Path) - убрать
- [ ] deploy добавить флаг nginx/web директории для копирования туда фронта
- [ ] настроить инфру: запуск taskfile, проверка тестов, линтеры
- [ ] админка
- [ ] кеширование

#### тех долг
- [ ] full rest доделать для всех видов ресурсов
- [ ] CSV `rowsTo[T]` uses reflection** — `repo/repo.go:388-446`: fragile, slow, type-unsafe CSV deserialization via `reflect`. Already in tasks.md but still high priority.
- [ ] разделение источников помимо VK


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

- [ ] завести гайд на ручное тестирование (в отдельном файле шаги и команды)

#### index.html

- [ ] сделать автообновление онлайна (и на room-page)

#### content.html

- [ ] доделать функции editSource и deleteSource

### DONE

- [x] `web/demo/` сохранён (player-api-demo.html) — не трогать.
- [x] bug fix: deleteRoom btn on main page
- [x] убрать var контейнеры
- [x] переименовать глобальный var контейнер в roomContainer
- [x] весь js код в файлы
- [x] пофиксить баг: при введении кода несуществующей комнаты в нее нельзя войти
- [x] пофиксить баг: онлайн разобраться 
- [x] пофиксить баг: после удаления комнат должно удаляться пространство
