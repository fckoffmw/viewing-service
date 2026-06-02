#!/usr/bin/env bash
set -euo pipefail

# ============================================================
#  test-ws.sh  —  Интерактивное тестирование WebSocket API w2g
#
#  Зависимости:  curl, python3 (c библиотекой websockets)
#  Использование:
#    ./test-ws.sh                  # автоматический режим
#    ./test-ws.sh --server PORT    # если сервер уже запущен
# ============================================================

RED='\033[0;31m'; GREEN='\033[0;32m'; CYAN='\033[0;36m'
YELLOW='\033[1;33m'; BOLD='\033[1m'; NC='\033[0m'
step()  { echo -e "\n${CYAN}══════════════════════════════════════════════${NC}"; echo -e "${BOLD}$1${NC}"; }
ok()    { echo -e "  ${GREEN}✓${NC} $1"; }
warn()  { echo -e "  ${YELLOW}⚠${NC} $1"; }
fail()  { echo -e "  ${RED}✗${NC} $1"; exit 1; }

cleanup() {
  step "🧹 Очистка"
  [ -n "${SERVER_PID:-}" ] && kill "$SERVER_PID" 2>/dev/null && ok "Сервер остановлен"
  rm -rf "$TMPDIR"
}
trap cleanup EXIT

# ---- Настройка ----
TMPDIR=$(mktemp -d)
PORT="${1:-8080}"
SERVER_PORT="${1:-8080}"
BASE="http://localhost:$SERVER_PORT"
WS="ws://localhost:$SERVER_PORT/ws"

# ---- 1. Запуск сервера (если не флаг --server) ----
if [ "${1:-}" = "--server" ]; then
  PORT="$2"
  BASE="http://localhost:$PORT"
  WS="ws://localhost:$PORT"
  warn "Сервер уже запущен, пропускаем запуск"
  shift 2
else
  step "🚀 Запуск сервера на порту $PORT"
  export STORAGE_DIR="$TMPDIR/storage"
  export LOG_LEVEL=error
  mkdir -p "$STORAGE_DIR"
  go run ./cmd/w2g &
  SERVER_PID=$!
  sleep 1
  for i in $(seq 1 10); do
    curl -sf "$BASE/healthz" >/dev/null 2>&1 && break
    sleep 0.5
  done
  curl -sf "$BASE/healthz" >/dev/null 2>&1 || fail "Сервер не запустился"
  ok "Сервер запущен (PID=$SERVER_PID)"
fi

# ---- 2. Регистрация ----
step "👤 Регистрация пользователей"
COOKIE1="$TMPDIR/cookie1.txt"
curl -sf -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"alice","password":"pass123"}' \
  -c "$COOKIE1" >/dev/null || fail "Не удалось зарегистрировать alice"
ok "alice зарегистрирована"

COOKIE2="$TMPDIR/cookie2.txt"
curl -sf -X POST "$BASE/auth/register" \
  -H "Content-Type: application/json" \
  -d '{"username":"bob","password":"pass123"}' \
  -c "$COOKIE2" >/dev/null || fail "Не удалось зарегистрировать bob"
ok "bob зарегистрирован"

SESSION1=$(grep session_id "$COOKIE1" | awk '{print $NF}')
SESSION2=$(grep session_id "$COOKIE2" | awk '{print $NF}')

# ---- 3. Создание комнаты ----
step "🏠 Создание комнаты"
ROOM=$(curl -sf -X POST "$BASE/api/rooms" \
  -H "Content-Type: application/json" \
  -b "$COOKIE1" \
  -d '{"name":"test-room"}') || fail "Не удалось создать комнату"
INVITE=$(echo "$ROOM" | python3 -c "import sys,json; print(json.load(sys.stdin)['invite_code'])")
ok "Комната создана, код: $INVITE"

# ---- 4. WebSocket тесты ----
step "🔌 WebSocket тесты"

PYTHON_SCRIPT=$(cat << 'PYEOF'
import asyncio, json, sys, time
import websockets

async def run_test(ws_url, session, invite_code, scenario):
    headers = {"Cookie": f"session_id={session}"}
    async with websockets.connect(f"{ws_url}/{invite_code}", additional_headers=headers) as ws:
        # Ждём sync
        sync = json.loads(await ws.recv())
        assert sync["type"] == "sync", f"Ожидался sync, получен {sync['type']}"
        print(f"  → sync получен: playing={sync['payload']['playing']}, position={sync['payload']['position']}")

        results = {}
        for action, payload, expected_type, expect_broadcast in scenario:
            await ws.send(json.dumps({"type": action, "payload": payload}))
            print(f"\n  🡒 send: {action} {json.dumps(payload)}")

            # Читаем свой ответ
            resp = json.loads(await ws.recv())
            assert resp["type"] == expected_type, f"Ожидался {expected_type}, получен {resp['type']}"
            print(f"  🡐 recv: {resp['type']}  username={resp.get('username','-')}  payload={resp.get('payload')}")
            results[action] = resp

            # Если нужен broadcast — читаем второе сообщение
            if expect_broadcast:
                bc = json.loads(await ws.recv())
                print(f"  🡐 recv: {bc['type']}  username={bc.get('username','-')}  payload={bc.get('payload')}")

        return results

async def two_users(ws_url, s1, s2, invite_code):
    h1 = {"Cookie": f"session_id={s1}"}
    h2 = {"Cookie": f"session_id={s2}"}

    async with websockets.connect(f"{ws_url}/{invite_code}", additional_headers=h1) as alice, \
               websockets.connect(f"{ws_url}/{invite_code}", additional_headers=h2) as bob:
        # sync для обоих
        for name, ws in [("alice", alice), ("bob", bob)]:
            sync = json.loads(await ws.recv())
            print(f"  {name} → sync получен (playing={sync['payload']['playing']})")

        # alice шлёт play
        await alice.send(json.dumps({"type": "play", "payload": {"position": 30}}))
        resp_a = json.loads(await alice.recv())
        resp_b = json.loads(await bob.recv())
        print(f"\n  alice play → alice: {resp_a['type']} pos={resp_a['payload']['position']}")
        print(f"  alice play → bob:   {resp_b['type']} pos={resp_b['payload']['position']}")
        assert resp_a["type"] == "play"
        assert resp_b["type"] == "play"

        # alice шлёт chat (bob должен получить, alice — нет)
        await alice.send(json.dumps({"type": "chat", "payload": {"text": "привет!"}}))
        # alice не получает свой chat
        # bob получает chat
        chat_b = json.loads(await bob.recv())
        print(f"\n  alice chat → alice: (не получает)")
        print(f"  alice chat → bob:   {chat_b['type']} username={chat_b['username']} text={chat_b['payload']['text']}")
        assert chat_b["type"] == "chat"
        assert chat_b["username"] == "alice"

        # bob шлёт seek
        await bob.send(json.dumps({"type": "seek", "payload": {"position": 120}}))
        sb = json.loads(await bob.recv())
        sa = json.loads(await alice.recv())
        print(f"\n  bob seek → bob:   {sb['type']} pos={sb['payload']['position']}")
        print(f"  bob seek → alice: {sa['type']} pos={sa['payload']['position']}")
        assert sb["type"] == "seek"
        assert sa["type"] == "seek"

        # проверка state через HTTP
        return

scenario = sys.argv[1] if len(sys.argv) > 1 else "all"
ws_url = sys.argv[2]
session1 = sys.argv[3]
session2 = sys.argv[4]
invite_code = sys.argv[5]

if scenario == "two_users":
    asyncio.run(two_users(ws_url, session1, session2, invite_code))
else:
    # Базовый сценарий: одиночный клиент
    base_scenario = [
        ("play",  {"position": 42.5}, "play",  False),
        ("pause", {"position": 55.0}, "pause", False),
        ("seek",  {"position": 120},  "seek",  False),
    ]
    asyncio.run(run_test(ws_url, session1, invite_code, base_scenario))
PYEOF
)

echo ""
python3 -c "$PYTHON_SCRIPT" "single" "$WS" "$SESSION1" "$SESSION2" "$INVITE" && ok "Одиночный тест пройден" || fail "Одиночный тест упал"

step "👥 Тест: два пользователя"
python3 -c "$PYTHON_SCRIPT" "two_users" "$WS" "$SESSION1" "$SESSION2" "$INVITE" && ok "Тест с двумя пользователями пройден" || fail "Тест с двумя пользователями упал"

# ---- 5. Проверка HTTP API после WS ----
step "🔍 Проверка HTTP: состояние комнаты"
ROOM_STATE=$(curl -sf "$BASE/api/rooms/$INVITE" -b "$COOKIE1") || warn "Не удалось получить комнату"
MEMBERS=$(echo "$ROOM_STATE" | python3 -c "import sys,json; print(json.load(sys.stdin)['members_online'])")
ok "Участников онлайн: $MEMBERS"

# ---- 6. Демонстрация sync с актуальным стейтом (пока alice online) ----
step "🔄 Демо: sync для нового клиента пока alice в комнате"
python3 -c "
import asyncio, json, websockets

async def demo():
    # alice подключается и шлёт play
    h1 = {'Cookie': 'session_id=$SESSION1'}
    async with websockets.connect('$WS/$INVITE', additional_headers=h1) as alice:
        sync = json.loads(await asyncio.wait_for(alice.recv(), timeout=5))
        print(f'  alice sync: playing={sync[\"payload\"][\"playing\"]} position={sync[\"payload\"][\"position\"]}')

        await alice.send(json.dumps({'type': 'play', 'payload': {'position': 42.5}}))
        await alice.recv()
        print(f'  alice play sent')

        # теперь подключается новичок — пока alice в комнате
        h2 = {'Cookie': 'session_id=$SESSION2'}
        async with websockets.connect('$WS/$INVITE', additional_headers=h2) as bob:
            sync2 = json.loads(await asyncio.wait_for(bob.recv(), timeout=5))
            p = sync2['payload']
            print(f'  bob sync:  playing={p[\"playing\"]} position={p[\"position\"]}')
            assert p['playing'] == True, f'Ожидался playing=true, получен {p[\"playing\"]}'
            assert p['position'] == 42.5, f'Ожидался position=42.5, получен {p[\"position\"]}'
            print(f'  ✓ Новый клиент получил актуальный стейт комнаты')

asyncio.run(demo())
" && ok "Новый клиент получил актуальный sync (playing=true, position=42.5)" || fail "sync не корректен"

# ---- 7. source_changed через HTTP ----
step "📺 Тест source_changed"
# Создаём источник
SRC=$(curl -sf -X POST "$BASE/api/sources" \
  -H "Content-Type: application/json" \
  -b "$COOKIE1" \
  -d '{"name":"Test Video","url":"https://example.com/video.mp4"}') || fail "Не удалось создать источник"
SRC_ID=$(echo "$SRC" | python3 -c "import sys,json; print(json.load(sys.stdin)['id'])")
ok "Источник создан: $SRC_ID"

# Меняем источник через HTTP и проверяем, что WS получает source_changed
python3 -c "
import asyncio, json, websockets
import subprocess, time

BASE = '$BASE'
INVITE = '$INVITE'
SESSION = '$SESSION1'
SRC_ID = '$SRC_ID'

async def check():
    h1 = {'Cookie': f'session_id={SESSION}'}
    async with websockets.connect(f'{BASE.replace(\"http\",\"ws\")}/ws/{INVITE}', additional_headers=h1) as alice:
        sync = json.loads(await asyncio.wait_for(alice.recv(), timeout=5))
        print(f'  sync получен (source_id={sync[\"payload\"][\"source_id\"]})')

        time.sleep(0.3)
        r = subprocess.run(
            ['curl', '-sf', '-X', 'PATCH', f'{BASE}/api/rooms/{INVITE}/source',
             '-H', 'Content-Type: application/json',
             '-b', f'session_id={SESSION}',
             '-d', '{\"source_id\":\"' + SRC_ID + '\"}'],
            capture_output=True, text=True
        )
        if r.returncode == 0:
            print(f'  → источник изменён через HTTP')
        else:
            print(f'  ✗ не удалось сменить источник: {r.stderr}')

        sc = json.loads(await asyncio.wait_for(alice.recv(), timeout=5))
        print(f'  → WS получил: {sc[\"type\"]} id={sc[\"payload\"][\"source_id\"]}')
        assert sc['type'] == 'source_changed'
        assert sc['payload']['source_id'] == SRC_ID
        print(f'  ✓ source_changed корректен')

asyncio.run(check())
" && ok "source_changed проверен" || fail "source_changed не прошёл"

# ---- Готово ----
step "✅ Все тесты пройдены"
echo -e "  Сервер:  ${GREEN}$BASE${NC}"
echo -e "  Сессии:  alice (${GREEN}$SESSION1${NC})  bob (${GREEN}$SESSION2${NC})"
echo -e "  Комната: ${GREEN}$INVITE${NC}"
echo ""
echo -e "  Попробуйте подключиться вручную:"
echo -e "    ${YELLOW}websocat -H=\"Cookie: session_id=$SESSION1\" $WS/$INVITE${NC}"
echo ""
