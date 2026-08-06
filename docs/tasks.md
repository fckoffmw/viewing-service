# Tasks

## Backend

### TODO

- [ ] full rest доделать для всех видов ресурсов
- [ ] тесты везде до 75+ %
- [ ] приватные комнаты
- [ ] линтер для if, for, return и др (wsl) + на длину строки
- [ ] админка
- [ ] кеширование
- [ ] postgresql
- [ ] redis

#### тех долг

- [ ] продумать структуру апи (всего)
- [ ] CSV `rowsTo[T]` uses reflection** — `repo/repo.go:388-446`: fragile, slow, type-unsafe CSV deserialization via `reflect`. Already in tasks.md but still high priority.
- [ ] разделение источников помимо VK
- [ ] логер в пакет ?

### DONE

- [x] логирование проработать / переработать
- [x] deploy добавить флаг nginx/web директории для копирования туда фронта
- [x] настроить инфру: запуск taskfile, проверка тестов, линтеры
- [x] sync
  - [x] client side: CRITICAL BUG: после сика + плей 1го - у 2го не делается плей а сик виден только после паус 1
  - [x] docs
  - [x] структура состояния
  - [x] типы возможных ивентов (сверяясь с sync.md)
  - [x] сервисный слой
  - [x] зачем ? extractInviteCodeFromPath(r.URL.Path) - убрать
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

- [ ] сделать автообновление для онлайна и для удаления комнат

#### index.html

- [ ] сделать автообновление  на room-page

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
- [x] автообновление на index.js
