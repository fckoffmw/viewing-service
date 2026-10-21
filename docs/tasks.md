# Tasks

## Backend

### TODO

- [ ] SPA fallback для фронта: фронт переезжает на React Router (одна точка входа `index.html`, маршруты `/login`, `/register`, дальше будут другие). Таких файлов в `dist` нет, поэтому при F5 или прямом заходе на `/login` сейчас будет 404. Нужно: если запрошенного файла нет и путь не `/auth/*`, `/api/*`, `/ws/*`, `/healthz` — отдавать `index.html` со статусом 200. Существующие файлы (`rooms.html`, `content.html`, `room.html`, `profile.html`, `/assets/*`) отдавать как раньше.
  - [ ] nginx на сервере (прод, `deploy.sh` кладёт `dist` в `/var/www/w2g`): в `location /` что-то вроде `try_files $uri $uri/ /index.html;`
  - [ ] `cmd/web/main.go` (локальный `task run`): то же поведение вместо голого `http.FileServer`, чтобы локально было как на проде
  - проверка: после `npm run build` открываются `/login`, `/register`, `/abracadabra` (React-страница 404), `/rooms.html`; `/api/...` по-прежнему уходит на бэкенд

- [ ] убрать `cp -r static dist/static` из `deploy.sh` и `Taskfile.yml` (`build:web`)
  папку `web/static` переименовали в `public` (коммит 28732b6), команда падает → из-за `set -e` деплой останавливается, новая версия фронта не выкатывается; `task build:web` тоже падает.
  копировать вручную больше не нужно: Vite сам кладёт содержимое `public` в `dist` (`publicDir` в vite.config.js исправлен).
  проверка: `task build:web` и `deploy.sh` проходят до конца, в `dist` есть `img/` и `assets/stickers/`



- [ ] full rest доделать для всех видов ресурсов
- [ ] тесты везде до 75+ %
- [ ] приватные комнаты
- [ ] линтер для if, for, return и др (wsl) + на длину строки
- [ ] админка
- [ ] кеширование
- [ ] postgresql
- [ ] redis
- [ ] расширить текущую логику вебсокет-сервера для синхронизации количества пользователей в реальном времени и мгновенного уведомления участников при удалении комнаты,
добавить в сокет отправку двух новых типов событий (data.type), когда кто-то подключается к комнате или отключается от неё:При входе:{ "type": "user_join", "username": "Ник_Пользователя" }При выходе:{ "type": "user_leave", "username": "Ник_Пользователя" }»

#### тех долг

- [ ] продумать структуру апи (всего)
- [ ] CSV `rowsTo[T]` uses reflection** — `repo/repo.go:388-446`: fragile, slow, type-unsafe CSV deserialization via `reflect`. Already in tasks.md but still high priority.
- [ ] разделение источников помимо VK
- [ ] логер в пакет ?

### DONE

- [x] реализовать обработку WebSocket-сообщения { "type": "sticker", "payload": { "id": "stickerID" } } на сервере, при получении сервер должен определить автора и разослать всем участникам комнаты ответный JSON вида { "type": "sticker", "username": "...", "payload": { "id": "..." } }.
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

- [x] завести гайд на ручное тестирование (в отдельном файле шаги и команды)

- [ ] переписать автообновление на страницах index.js и room.js с помощью ws

- [ ] определить как лучше доставать объекты в константы или в переменные

- [ ] bug: zoom: 1.33 на body в app.css:18, он увеличивает всё на треть, и на узком экране страница 
перестаёт помещаться

- [ ] в room.js:380 стикеры в чате подставляются по пути "public/assets/stickers/" + text + ".webp". Но в Vite файлы из public доступны от корня, без слова public: на сервере стикер лежит по адресу /assets/stickers/0.webp, а не /public/assets/....  Тот же старый путь /static/... есть в room.js:16.

#### index.html

### room.html

- [ ] сделать чтобы было видно кто присоединился и кто отсоединился
- [ ] починить баг со скроллом после отправки стикеров (скроллится раз через раз)

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
- [x] автообновление  на room-page
- [x] починить баг со сменой источника, удалением комнаты и копированием ссылки
- [x] переделать устройство папок под стандарты vite
- [x] починить баг с открывающейся панелью стикеров после отправки текстового сообщения
- [x] починить баг с вылетающей картинкой стикеров при переходе с index на room
- [x] сделать чтобы было видно время сообщений
- [x] починить баг: функция copyInvite не работает
