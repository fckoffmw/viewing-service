(async function checkAlreadyLogged() {
    try {
    var res = await fetch('/auth/me')
    if (res.ok) window.location.href = '/index.html'
    } catch {}
})()

document.getElementById('login-form').addEventListener('submit', async function (e) {
    e.preventDefault()
    var username = document.getElementById('username').value.trim()
    var password = document.getElementById('password').value
    var errorEl = document.getElementById('error')
    var btn = this.querySelector('button')

    errorEl.style.display = 'none'
    btn.disabled = true
    btn.textContent = 'Вход...'

    try {
    var res = await fetch('/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username: username, password: password }),
    })
    var data = await res.json()

    if (res.ok) {
        window.location.href = '/index.html'
        return
    }
    errorEl.textContent = data.error || 'Ошибка входа'
    errorEl.style.display = 'block'
    } catch {
    errorEl.textContent = 'Ошибка соединения'
    errorEl.style.display = 'block'
    } finally {
    btn.disabled = false
    btn.textContent = 'Войти'
    }
})