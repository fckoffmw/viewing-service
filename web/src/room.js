var currentUsername = "";
var currentUserId = "";
var inviteCode = "";
var roomOwnerId = "";
var ws = null;
var reconnectTimer = null;

var vkPlayer = null;
var pollTimer = null;
var suppressCount = 0;
var lastKnownState = { playing: false, position: 0 };
var pendingSync = null;

const MEMBERS_ONLINE_POLLING_INTERVAL = 10000 // in ms

const STICKER_MODULES = import.meta.glob('/static/assets/stickers/**/*.{png,webp,jpg,gif}');
const STICKER_PATHS = Object.keys(STICKER_MODULES);
const STICKER_COUNT = STICKER_PATHS.length;

const STICKER_BTN = document.getElementById("room-chat-stickers-btn");
const STICKER_PANEL = document.getElementById("room-chat-sticker-panel");
const STICKER_GRID = STICKER_PANEL.querySelector(".room-chat-sticker-grid");

if (STICKER_GRID) {
  STICKER_GRID.addEventListener("click", function (e) {
    var sticker = e.target.closest(".room-chat-sticker-item");
    if (!sticker) return;

    if (!ws || ws.readyState !== WebSocket.OPEN) {
      if (typeof showToast === "function") showToast("Ошибка: нет соединения с сервером");
      return;
    }

    var stickerID = sticker.getAttribute("data-id");

    wsSend("sticker", { id: stickerID });

    appendMessage("sticker", currentUsername, stickerID); // когда появится логика ws для "sticker" - удалить

  }
)
}

window.addEventListener("beforeunload", function () {
  if (reconnectTimer) { clearTimeout(reconnectTimer); reconnectTimer = null; }

  if (ws){
    ws.onclose = null
    ws.close()
  }
})

function wsSend(type, payload) {
  var msg = JSON.stringify({type: type, payload: payload});
  console.log("[WS] send", type, JSON.stringify(payload));
  if (ws && ws.readyState === WebSocket.OPEN) ws.send(msg);
}

function pollPlayer() {
  if (!vkPlayer) return;

  try {
    var currentTime = vkPlayer.getCurrentTime() || 0;
    var stateStr = vkPlayer.getState();
    if (!stateStr || stateStr === 'uninited') return;

    var currentPlaying = stateStr === 'playing';

    if (suppressCount > 0) {
      suppressCount--;
      lastKnownState.playing = currentPlaying;
      lastKnownState.position = currentTime;
      return;
    }

    if (currentPlaying !== lastKnownState.playing) {
      console.log("[POLL] play state change", currentPlaying ? "play" : "pause", "@", currentTime);
      wsSend(currentPlaying ? "play" : "pause", { position: currentTime });
      lastKnownState.playing = currentPlaying;
      lastKnownState.position = currentTime;
      return;
    }

    var diff = currentTime - lastKnownState.position;
    lastKnownState.position = currentTime;
    if (diff > 3.0 || diff < -3.0) {
      console.log("[POLL] seek detected @", currentTime, "diff:", diff);
      wsSend("seek", { position: currentTime, playing: currentPlaying });
    }
  } catch(e) {
    console.log("[POLL] error:", e);
  }
}

function startPoll() {
  if (pollTimer) return;
  console.log("[POLL] starting");
  pollTimer = setInterval(pollPlayer, 200);
}

function stopPoll() {
  if (pollTimer) {
    clearInterval(pollTimer);
    pollTimer = null;
  }
}

function applyRemoteCommand(type, data) {
  var p = data.payload || data;
  if (!p || !vkPlayer) return;

  console.log("[CMD] apply", type, JSON.stringify(p));
  suppressCount = 4;

  switch (type) {
    case "sync":
      var pos = p.position;
      if (p.playing && p.updated_at) {
        pos = Math.max(0, pos + (Date.now() - new Date(p.updated_at).getTime()) / 1000);
      }
      vkPlayer.seek(pos);
      if (p.playing) { vkPlayer.play(); } else { vkPlayer.pause(); }
      lastKnownState = { playing: !!p.playing, position: pos };
      break;
    case "play":
      vkPlayer.seek(p.position);
      vkPlayer.play();
      lastKnownState = { playing: true, position: p.position };
      break;
    case "pause":
      vkPlayer.seek(p.position);
      vkPlayer.pause();
      lastKnownState = { playing: false, position: p.position };
      break;
    case "seek":
      vkPlayer.seek(p.position);
      if (p.playing === true) {
        vkPlayer.play();
        lastKnownState = { playing: true, position: p.position };
      } else if (p.playing === false) {
        vkPlayer.pause();
        lastKnownState = { playing: false, position: p.position };
      } else {
        lastKnownState.position = p.position;
      }
      break;
  }
}

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
  stopPoll();
  if (vkPlayer) {
    try { vkPlayer.destroy(); } catch(e) {}
    vkPlayer = null;
  }
  lastKnownState = { playing: false, position: 0 };
  suppressCount = 0;

  var decoded = decodeURIComponent(url);
  var iframe = document.getElementById("room-player");

  try {
    var u = new URL(decoded);
    if (u.hostname.indexOf("vk") !== -1 && !u.searchParams.has("js_api")) {
      u.searchParams.set("js_api", "1");
      decoded = u.toString();
    }
  } catch (e) {}

  iframe.src = decoded;
  iframe.style.display = "block";
  document.getElementById("room-placeholder").style.display = "none";

  iframe.addEventListener("load", function initVK() {
    iframe.removeEventListener("load", initVK);
    initPlayer();
  });
}

function initPlayer() {
  if (!window.VK) {
    setTimeout(initPlayer, 1000);
    return;
  }

  var iframe = document.getElementById("room-player");
  if (!iframe || !iframe.src) return;
  if (vkPlayer) return;

  try {
    vkPlayer = VK.VideoPlayer(iframe);
    console.log("[VK] player initialized");

    if (pendingSync) {
      applyRemoteCommand("sync", pendingSync);
      pendingSync = null;
    }

    startPoll();
  } catch (e) {
    console.log("[VK] init error:", e);
  }
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

    setInterval(fetchMembers, MEMBERS_ONLINE_POLLING_INTERVAL);
  };

  ws.onmessage = function(e) {
    var data = JSON.parse(e.data);
    console.log("[WS] recv", data.type, JSON.stringify(data.payload));

    switch (data.type) {
      case "sync":
        pendingSync = data;
        if (vkPlayer && data.payload) {
          applyRemoteCommand("sync", data);
        }
        return;

      case "play":
      case "pause":
      case "seek":
        if (vkPlayer && data.payload) {
          applyRemoteCommand(data.type, data);
        }
        return;

      case "source_changed":
        if (data.payload && data.payload.source_url) {
          showVideo(data.payload.source_url);
          refreshSourceName();
        }
        return;

      case "chat":
        if (data.payload) {
          var type = data.username === currentUsername ? "me" : "other";
          appendMessage(type, data.username, data.payload.text || "");
        }
        return;

      case "sticker":
        if (data.payload) {
          appendMessage("sticker", data.username, data.payload.id || "");
        }
    }
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

  var now = new Date();
  var timeString = now.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  
  var timeSpan = document.createElement("span");
  timeSpan.className = "room-msg-time";
  timeSpan.textContent = timeString;
  
  if (type === "system") {
    div.className = "room-msg system";
    div.textContent = text;
    div.appendChild(timeSpan);
  } else if (type === "sticker") {
    div.className = "room-msg";

    var color = getUserColor(username);
    
    var headerDiv = document.createElement("div");
    headerDiv.className = "room-msg-header";
    headerDiv.style.borderLeftColor = color;

    var nameSpan = document.createElement("span");
    nameSpan.className = "room-msg-username";
    nameSpan.style.color = color;
    nameSpan.textContent = username + ":";
    
    var img  = document.createElement("img");
    img.src = "public/assets/stickers/" + text + ".webp";
    img.className = "room-msg-sticker";
    
    headerDiv.appendChild(nameSpan);
    headerDiv.appendChild(img);
    
    var contentWrapper = document.createElement("div");
    contentWrapper.className = "room-msg-content";
    
    var timeSpanClone = timeSpan.cloneNode(true);
    timeSpanClone.style.color = color;
    contentWrapper.appendChild(timeSpanClone);
    
    div.appendChild(headerDiv);
    div.appendChild(contentWrapper);
  } else {
    div.className = "room-msg";

    var color = getUserColor(username);
    
    var headerDiv = document.createElement("div");
    headerDiv.className = "room-msg-header";
    headerDiv.style.borderLeftColor = color;

    var nameSpan = document.createElement("span");
    nameSpan.className = "room-msg-username";
    nameSpan.style.color = color;
    nameSpan.textContent = username + ":";
    
    var textSpan = document.createElement("span");
    textSpan.className = "room-msg-text";
    textSpan.textContent = text;
    
    headerDiv.appendChild(nameSpan);
    headerDiv.appendChild(textSpan);
    
    var contentWrapper = document.createElement("div");
    contentWrapper.className = "room-msg-content";
    
    var timeSpanClone = timeSpan.cloneNode(true);
    timeSpanClone.style.color = color;
    contentWrapper.appendChild(timeSpanClone);
    
    div.appendChild(headerDiv);
    div.appendChild(contentWrapper);
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

  if (STICKER_PANEL.classList.contains("open") && !STICKER_BTN.contains(e.target) && !STICKER_PANEL.contains(e.target)) {
    STICKER_PANEL.classList.remove("open");
  }
});

document.getElementById("room-menu-btn").addEventListener("click", function(e) {
  e.stopPropagation();
  document.getElementById("room-header-menu").classList.toggle("open");
});

STICKER_BTN.addEventListener("click", function(e) {
  if (e.detail === 0) return; 
  e.stopPropagation();
  STICKER_PANEL.classList.toggle("open");
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
window.copyLink = copyLink;

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
window.deleteRoom = deleteRoom;

// --- source modal ---
function openSourceModal() {
  document.getElementById("room-header-menu").classList.remove("open");
  document.getElementById("source-modal").style.display = "flex";
  loadSourceList();
}
window.openSourceModal = openSourceModal;

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
  wsSend("chat", { text: text });
  appendMessage("me", currentUsername, text);
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
