var currentUserId = ''
var currentUsername = ''
var roomContainer = document.getElementById('rooms-container');
var rooms = []

roomContainer.addEventListener('click', function(event) {
    if (event.target.classList.contains('btn-delete-room')) {
    var code = event.target.getAttribute('data-code');

    if (code) {
        deleteRoom(code);
    }
    }
});

checkAuth()

async function checkAuth() {
    try {
    var res = await fetch('/auth/me')
    if (!res.ok) { window.location.href = '/login.html'; return }
    var user = await res.json()
    currentUserId = user.id
    currentUsername = user.username
    document.getElementById('nav-username').textContent = user.username
    loadRooms()
    } catch {
    window.location.href = '/login.html'
    }
}

async function logout() {
    await fetch('/auth/logout', { method: 'POST' })
    window.location.href = '/login.html'
}

// --- rooms ---
async function loadRooms() {
    roomContainer.innerHTML = '<div class="empty">Загрузка...</div>'

    try {
    var res = await fetch('/api/rooms')
    if (!res.ok) { roomContainer.innerHTML = '<div class="empty">Ошибка загрузки</div>'; return }
    rooms = await res.json()
    renderRooms(rooms)
    } catch {
    roomContainer.innerHTML = '<div class="empty">Ошибка соединения</div>'
    }
}

function renderRooms(rooms) {
    if (rooms.length === 0) {
    roomContainer.innerHTML = '<div class="empty">Нет комнат. Создайте первую.</div>'
    return
    }

    roomContainer.innerHTML = rooms.map(function (r) {
    var source = r.current_source ? r.current_source.name : '\u2014'
    var members = r.members_online || 0
    var isOwner = r.owner_id === currentUserId
    return (
        '<div class="room-item">' +
        '<div class="room-info">' +
            '<h3>' + esc(r.name) + '</h3>' +
            '<div class="room-meta">' + esc(source) + ' \u00b7 ' + members + ' онлайн \u00b7 <span class="room-code" onclick="copyInvite(\'' + r.invite_code + '\')" title="Копировать ссылку">' + r.invite_code + '</span></div>' +
        '</div>' +
        '<div class="room-actions">' +
            '<a href="/room.html?invite=' + r.invite_code + '" class="btn btn-primary btn-sm">Войти</a>' +
            (isOwner ? '<button data-code="' + esc(r.invite_code) + '" class="btn btn-danger btn-sm btn-delete-room">\u2715</button>' : '') +
        '</div>' +
        '</div>'
    )
    }).join('')
}

function esc(s) {
    var d = document.createElement('div')
    d.textContent = s
    return d.innerHTML
    .replace(/\\/g, '\\\\')
    .replace(/'/g, '&#39;')
    .replace(/"/g, '&quot;');
}

function copyInvite(code) {
    var url = window.location.origin + '/room.html?invite=' + code
    navigator.clipboard.writeText(url).then(function () {
    showToast('Ссылка скопирована')
    })
}

// --- create room ---
function openCreateModal() {
    document.getElementById('create-modal').style.display = 'flex'
    document.getElementById('room-name').value = ''
    document.getElementById('room-name').focus()
}

function closeCreateModal() {
    document.getElementById('create-modal').style.display = 'none'
}

async function handleCreate(e) {
    e.preventDefault()
    var name = document.getElementById('room-name').value.trim()
    if (!name) return

    var btn = e.target.querySelector('button[type="submit"]')
    btn.disabled = true
    btn.textContent = 'Создание...'

    try {
    var res = await fetch('/api/rooms', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ name: name }),
    })
    var data = await res.json()
    if (res.ok) {
        closeCreateModal()
        window.location.href = '/room.html?invite=' + data.invite_code
    } else {
        alert(data.error || 'Ошибка создания')
    }
    } catch {
    alert('Ошибка соединения')
    } finally {
    btn.disabled = false
    btn.textContent = 'Создать'
    }
}

// --- join ---
async function joinByCode() {
    var code = document.getElementById('join-code').value.trim().toUpperCase()
    if (!code) return

    try {
        var res = await fetch('/api/rooms/' + code, { method: 'GET' })
        if (res.ok) {
            window.location.href = '/room.html?invite=' + code
        } else {
            showToast('Нет такой комнаты')
        }
    } catch {
        alert('Ошибка соединения')
    }
}

// --- delete room ---
async function deleteRoom(code) {
    var delRoom = rooms.find(function (r) {
    return r.invite_code === code
    })

    if (!confirm('Удалить комнату "' + delRoom.name + '"? Это действие нельзя отменить.')) return
    
    try {
        var res = await fetch('/api/rooms/' + delRoom.invite_code, { method: 'DELETE' })
        if (res.ok) {
            loadRooms()
            showToast('Комната удалена')
        } else {
            var data = await res.json()
            alert(data.error || 'Ошибка удаления')
        }
    } catch {
        alert('Ошибка соединения')
    }
}

// --- toast ---
function showToast(msg) {
    var el = document.createElement('div')
    el.textContent = msg
    el.className = 'toast'
    document.body.appendChild(el)
    setTimeout(function () { el.style.opacity = '0'; setTimeout(function () { el.remove() }, 300) }, 2000)
}

var sakuraPetals = [];

function createPetal() {
    var petal = document.createElement('div')
    petal.className = 'sakura'
    
    var size = 10 + Math.random() * 15
    petal.style.width = size + 'px'
    petal.style.height = size + 'px'
    
    var startX = Math.random() * window.innerWidth
    petal.style.left = startX + 'px'
    petal.style.top = '-30px' 
    
    petal.style.opacity = 0.6 + Math.random() * 0.4
    
    document.body.appendChild(petal)
    
    return {
        element: petal,
        x: startX,        
        y: -30,          
        speed: 1 + Math.random() * 2,  
        amplitude: 20 + Math.random() * 30,  
        phase: Math.random() * Math.PI * 2,  
        rotation: 0,      
        rotationSpeed: (Math.random() - 0.5) * 4
    }
}

function animateSakura() {
    if (Math.random() < 0.05) { 
        sakuraPetals.push(createPetal())
    }
    
    for (var i = sakuraPetals.length - 1; i >= 0; i--) {
        var p = sakuraPetals[i]
        
        p.y += p.speed
        
        p.x += Math.sin(p.y * 0.02 + p.phase) * 0.5
        
        p.rotation += p.rotationSpeed

        p.element.style.top = p.y + 'px'
        p.element.style.left = p.x + 'px'
        p.element.style.transform = 'rotate(' + p.rotation + 'deg)'
        
        if (p.y > window.innerHeight) {
            p.element.remove()       
            sakuraPetals.splice(i, 1) 
        }
    }

    requestAnimationFrame(animateSakura)
}

animateSakura()