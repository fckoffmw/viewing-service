# Tasks

## Backend

### TODO

- [ ] global review
  - [ ] done in global-review section

- [ ] rooms: docs, review, tests
- [ ] побольше логирования с разделением по уровням
- [ ] sync
  - [ ] ws взаимодействие
  - [ ] nats
- [ ] чтобы страница обновлялась у всех при изменении источника
- [ ] разделение источников помимо VK
- [ ] CSV storage — убрать reflect, использовать ручной маппинг или кодогенерацию
- [ ] кеширование
- [ ] auth more +

### DONE

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

- **[INFO] `cmd/web/main.go` is a frontend dev-server** — not the main backend. It serves static files and proxies API to backend on `:8080`. `cmd/w2g/main.go` is the real backend. The two binaries have distinct purposes — worth documenting to avoid confusion.

- **[REFLECT] CSV `rowsTo[T]` uses reflection** — `repo/repo.go:388-446`: fragile, slow, type-unsafe CSV deserialization via `reflect`. Already in tasks.md but still high priority.



