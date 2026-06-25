var currentUsername = "";
var currentUserId = "";
var inviteCode = "";
var roomOwnerId = "";
var ws = null;
var reconnectTimer = null;

async function checkAuth() {
  try {
    var res = await fetch("/auth/me");
    if (!res.ok) { window.location.href = "/login.html"; return false; }
    var user = await res.json();
    currentUsername = user.username;
    currentUserId = user.id;
    return true;
  } catch {
    window.location.href = "/login.html";
    return false;
  }
}

async function loadRoom() {
  try {
    var res = await fetch("/api/rooms/" + inviteCode);
    if (!res.ok) { setPlaceholder("Комната не найдена"); return; }
    var room = await res.json();

    document.title = "w2g — " + (room.name || "Комната");
    roomOwnerId = room.owner_id;
    updateMembers(room.members_online);

    if (room.current_source && room.current_source.url) {
      showVideo(room.current_source.url);
      updateSourceName(room.current_source.name || "");
    } else {
      setPlaceholder('Источник не выбран. <a href="/content.html?invite=' + inviteCode + '">Выберите на странице контента</a>.');
      updateSourceName("");
    }

    var isOwner = roomOwnerId === currentUserId;
    document.getElementById("menu-change-source").style.display = isOwner ? "" : "none";
    document.getElementById("menu-delete-room").style.display = isOwner ? "" : "none";
  } catch {
    setPlaceholder("Ошибка загрузки");
  }
}

function setPlaceholder(html) {
  var el = document.getElementById("room-placeholder");
  el.innerHTML = html;
  el.style.display = "flex";
  document.getElementById("room-player").style.display = "none";
}

function showVideo(url) {
  var decoded = decodeURIComponent(url);
  document.getElementById("room-player").src = decoded;
  document.getElementById("room-player").style.display = "block";
  document.getElementById("room-placeholder").style.display = "none";
}

function updateSourceName(name) {
  document.getElementById("room-source-name").textContent = name || "";
}

function updateMembers(n) {
  var el = document.getElementById("room-member-count");
  el.textContent = n > 0 ? "· " + n : "";
}

function setStatus(status) {
  var dot = document.getElementById("room-status-dot");
  var text = document.getElementById("room-online-text");
  dot.className = "room-status-dot " + status;
  text.textContent = status === "online" ? "онлайн" : "отключено";
}

function connectWS() {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }
  var proto = location.protocol === "https:" ? "wss:" : "ws:";
  ws = new WebSocket(proto + "//" + location.host + "/ws/" + inviteCode);

  ws.onopen = function() {
    setStatus("online");
    appendMessage("system", "", "Подключились");
    fetchMembers();
  };

  ws.onmessage = function(e) {
    var data = JSON.parse(e.data);
    if (data.type === "source_changed") {
      if (data.payload && data.payload.source_url) {
        showVideo(data.payload.source_url);
        refreshSourceName();
      }
      return;
    }
    var type = data.username === currentUsername ? "me" : "other";
    appendMessage(type, data.username, data.text);
  };

  ws.onclose = function() {
    setStatus("offline");
    appendMessage("system", "", "Отключились");
    scheduleReconnect();
  };
}

function scheduleReconnect() {
  reconnectTimer = setTimeout(connectWS, 5000);
}

async function fetchMembers() {
  try {
    var res = await fetch("/api/rooms/" + inviteCode);
    if (res.ok) {
      var room = await res.json();
      updateMembers(room.members_online);
    }
  } catch {}
}

async function refreshSourceName() {
  try {
    var res = await fetch("/api/rooms/" + inviteCode);
    if (res.ok) {
      var room = await res.json();
      var name = room.current_source ? room.current_source.name : "";
      updateSourceName(name);
    }
  } catch {}
}

function appendMessage(type, username, text) {
  var el = document.getElementById("room-messages");
  var div = document.createElement("div");
  if (type === "system") {
    div.className = "room-msg system";
    div.textContent = text;
  } else {
    div.className = "room-msg";
    var color = getUserColor(username);
    div.style.borderLeftColor = color;
    var nameSpan = document.createElement("span");
    nameSpan.className = "room-msg-username";
    nameSpan.style.color = color;
    nameSpan.textContent = username + ":";
    var textSpan = document.createElement("span");
    textSpan.className = "room-msg-text";
    textSpan.textContent = text;
    div.appendChild(nameSpan);
    div.appendChild(textSpan);
  }
  el.appendChild(div);
  el.scrollTop = el.scrollHeight;
}

function getUserColor(name) {
  var hash = 2166136261;
  for (var i = 0; i < name.length; i++) {
    hash ^= name.charCodeAt(i);
    hash = Math.imul(hash, 16777619);
  }
  return "hsl(" + Math.abs(hash % 360) + ", 80%, 65%)";
}

// --- menu ---
document.addEventListener("click", function(e) {
  var menu = document.getElementById("room-header-menu");
  var btn = document.getElementById("room-menu-btn");
  if (menu.classList.contains("open") && !btn.contains(e.target) && !menu.contains(e.target)) {
    menu.classList.remove("open");
  }
});

document.getElementById("room-menu-btn").addEventListener("click", function(e) {
  e.stopPropagation();
  document.getElementById("room-header-menu").classList.toggle("open");
});

function copyLink() {
  var url = window.location.origin + "/room.html?invite=" + inviteCode;
  navigator.clipboard.writeText(url).then(function() {
    showToast("Ссылка скопирована");
  }, function() {
    showToast("Ошибка копирования");
  });
  document.getElementById("room-header-menu").classList.remove("open");
}

function deleteRoom() {
  if (!confirm("Удалить комнату?")) return;
  fetch("/api/rooms/" + inviteCode, { method: "DELETE" }).then(function(res) {
    if (res.ok) {
      window.location.href = "/";
    } else {
      showToast("Ошибка удаления");
    }
  }).catch(function() {
    showToast("Ошибка удаления");
  });
  document.getElementById("room-header-menu").classList.remove("open");
}

// --- source modal ---
function openSourceModal() {
  document.getElementById("room-header-menu").classList.remove("open");
  document.getElementById("source-modal").style.display = "flex";
  loadSourceList();
}

function closeSourceModal() {
  document.getElementById("source-modal").style.display = "none";
}

async function loadSourceList() {
  var list = document.getElementById("source-list");
  try {
    var res = await fetch("/api/sources");
    if (!res.ok) { list.innerHTML = "<p style='color:#555;text-align:center;padding:12px;font-size:13px;'>Ошибка загрузки</p>"; return; }
    var sources = await res.json();
    if (sources.length === 0) {
      list.innerHTML = "<p style='color:#555;text-align:center;padding:12px;font-size:13px;'>Нет источников. <a href='/content.html' style='color:#888;'>Добавьте</a>.</p>";
      return;
    }
    list.innerHTML = "";
    sources.forEach(function(s) {
      var btn = document.createElement("button");
      btn.className = "modal-source-item";
      btn.textContent = s.name;
      btn.onclick = function() { selectSource(s.id); };
      list.appendChild(btn);
    });
  } catch {
    list.innerHTML = "<p style='color:#555;text-align:center;padding:12px;font-size:13px;'>Ошибка загрузки</p>";
  }
}

async function selectSource(id) {
  try {
    var res = await fetch("/api/rooms/" + inviteCode + "/source", {
      method: "PATCH",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ source_id: id }),
    });
    if (res.ok) {
      closeSourceModal();
    } else {
      showToast("Не удалось сменить источник");
    }
  } catch {
    showToast("Ошибка");
  }
}

// --- form ---
document.getElementById("room-chat-form").addEventListener("submit", function(e) {
  e.preventDefault();
  var input = document.getElementById("room-chat-input");
  var text = input.value.trim();
  if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;
  ws.send(JSON.stringify({ text: text }));
  input.value = "";
});

// --- toast ---
function showToast(msg) {
  var t = document.createElement("div");
  t.className = "toast";
  t.textContent = msg;
  document.body.appendChild(t);
  setTimeout(function() { t.remove(); }, 2500);
}

// --- start ---
window.addEventListener("DOMContentLoaded", async function() {
  var ok = await checkAuth();
  if (!ok) return;

  var params = new URLSearchParams(window.location.search);
  inviteCode = params.get("invite");
  if (!inviteCode) {
    setPlaceholder("Нет кода приглашения");
    return;
  }

  await loadRoom();
  connectWS();
});