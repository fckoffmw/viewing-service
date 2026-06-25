var currentUserId = "";
var currentUsername = "";
var inviteCode = "";
var activeSourceId = null;

var params = new URLSearchParams(window.location.search);
inviteCode = params.get("invite");

if (inviteCode) {
    document.getElementById("room-link").href = "/room.html?invite=" + inviteCode;
    document.getElementById("room-link").style.display = "inline-flex";
}

// --- auth ---
checkAuth();

async function checkAuth() {
    try {
    var res = await fetch("/auth/me");
    if (!res.ok) { window.location.href = "/login.html"; return; }
    var user = await res.json();
    currentUserId = user.id;
    currentUsername = user.username;
    document.getElementById("nav-username").textContent = user.username;
    loadSources();
    } catch {
    window.location.href = "/login.html";
    }
}

async function logout() {
    await fetch("/auth/logout", { method: "POST" });
    window.location.href = "/login.html";
}

// --- load sources ---
async function loadSources() {
    var container = document.getElementById("sources-container");

    if (inviteCode) {
    try {
        var roomRes = await fetch("/api/rooms/" + inviteCode);
        if (roomRes.ok) {
        var room = await roomRes.json();
        if (room.current_source) {
            activeSourceId = room.current_source.id;
        }
        }
    } catch {}
    }

    try {
    var res = await fetch("/api/sources");
    if (!res.ok) {
        container.innerHTML = '<div class="empty">Ошибка загрузки</div>';
        return;
    }
    var sources = await res.json();
    if (sources.length === 0) {
        container.innerHTML = '<div class="empty">Нет источников. Добавьте первый.</div>';
        return;
    }
    container.innerHTML = sources.map(function(s) {
        var isActive = s.id === activeSourceId;
        var activeClass = isActive ? " source-item-active" : "";
        var activeBtn = "";
        if (inviteCode) {
        var disabled = isActive ? " disabled" : "";
        var label = isActive ? "Активен" : "Сделать активным";
        activeBtn = '<button onclick="setActive(\'' + s.id + '\')" class="btn btn-primary btn-sm"' + disabled + ">" + label + "</button>";
        }
        return (
        '<div class="source-item' + activeClass + '">' +
            '<div class="source-item-info">' +
            '<div class="source-item-name">' + esc(s.name) + '</div>' +
            '<div class="source-item-url">' + esc(s.url) + '</div>' +
            '</div>' +
            '<div class="source-item-actions">' +
            activeBtn +
            '<button onclick="editSource(\'' + s.id + '\')" class="btn btn-secondary btn-sm">✎</button>' +
            '<button onclick="deleteSource(\'' + s.id + '\', \'' + esc(s.name) + '\')" class="btn btn-danger btn-sm">✕</button>' +
            '</div>' +
        '</div>'
        );
    }).join("");
    } catch {
    container.innerHTML = '<div class="empty">Ошибка загрузки</div>';
    }
}

// --- add source ---
function openAddModal() {
    document.getElementById("add-modal").style.display = "flex";
    document.getElementById("source-name").value = "";
    document.getElementById("source-url").value = "";
    document.getElementById("source-name").focus();
}

function closeAddModal() {
    document.getElementById("add-modal").style.display = "none";
}

document.getElementById("add-form").addEventListener("submit", async function(e) {
    e.preventDefault();
    var name = document.getElementById("source-name").value.trim();
    var url = document.getElementById("source-url").value.trim();
    if (!name || !url) return;

    try {
    var res = await fetch("/api/sources", {
        method: "POST",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ name: name, url: url }),
    });
    if (res.ok) {
        closeAddModal();
        loadSources();
    } else {
        var data = await res.json();
        showToast("Ошибка: " + (data.error || "неизвестная"));
    }
    } catch {
    showToast("Ошибка сети");
    }
});

// --- set active ---
async function setActive(id) {
    if (!inviteCode) return;
    try {
    var res = await fetch("/api/rooms/" + inviteCode + "/source", {
        method: "PATCH",
        headers: { "Content-Type": "application/json" },
        body: JSON.stringify({ source_id: id }),
    });
    if (res.ok) {
        loadSources();
    } else {
        var data = await res.json();
        showToast("Ошибка: " + (data.error || "неизвестная"));
    }
    } catch {
    showToast("Ошибка сети");
    }
}

// --- stubs ---
function editSource(id) {
    showToast("Редактирование — будет позже");
}

function deleteSource(id, name) {
    showToast("Удаление — будет позже");
}

// --- utils ---
function esc(s) {
    var d = document.createElement("div");
    d.textContent = s;
    return d.innerHTML;
}

function showToast(msg) {
    var t = document.createElement("div");
    t.className = "toast";
    t.textContent = msg;
    document.body.appendChild(t);
    setTimeout(function() { t.remove(); }, 2500);
}