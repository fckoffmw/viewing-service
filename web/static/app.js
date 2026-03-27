(function () {
  const messagesEl = document.getElementById("messages");
  const formEl = document.getElementById("chat-form");
  const inputEl = document.getElementById("message-input");
  const statusEl = document.getElementById("status");

  const clientId = self.crypto && crypto.randomUUID ? crypto.randomUUID() : String(Date.now());
  const maxLen = 500;
  let socket = null;
  let reconnectDelay = 1000;
  let reconnectTimer = null;

  function wsUrl() {
    const scheme = window.location.protocol === "https:" ? "wss:" : "ws:";
    return `${scheme}//${window.location.host}/ws`;
  }

  function setStatus(isOnline) {
    statusEl.textContent = isOnline ? "online" : "offline";
    statusEl.classList.toggle("online", isOnline);
    statusEl.classList.toggle("offline", !isOnline);
    formEl.querySelector("button").disabled = !isOnline;
  }

  function addMessage(sender, text, isMe) {
    const row = document.createElement("div");
    row.className = `message ${isMe ? "me" : "not-me"}`;

    const name = document.createElement("div");
    name.className = "name";
    name.textContent = sender;

    const body = document.createElement("div");
    body.className = "text";
    body.textContent = text;

    row.appendChild(name);
    row.appendChild(body);
    messagesEl.appendChild(row);
    messagesEl.scrollTop = messagesEl.scrollHeight;
  }

  function connect() {
    socket = new WebSocket(wsUrl());
    setStatus(false);

    socket.addEventListener("open", function () {
      reconnectDelay = 1000;
      setStatus(true);
    });

    socket.addEventListener("message", function (event) {
      let payload;
      try {
        payload = JSON.parse(event.data);
      } catch (e) {
        return;
      }
      if (!payload || typeof payload.text !== "string") {
        return;
      }
      const isMe = payload.clientId === clientId;
      addMessage(isMe ? "Me" : "NotMe", payload.text, isMe);
    });

    socket.addEventListener("close", function () {
      setStatus(false);
      if (reconnectTimer) {
        clearTimeout(reconnectTimer);
      }
      reconnectTimer = setTimeout(connect, reconnectDelay);
      reconnectDelay = Math.min(reconnectDelay * 2, 10000);
    });

    socket.addEventListener("error", function () {
      socket.close();
    });
  }

  formEl.addEventListener("submit", function (event) {
    event.preventDefault();
    if (!socket || socket.readyState !== WebSocket.OPEN) {
      return;
    }

    const text = inputEl.value.trim();
    if (!text || text.length > maxLen) {
      return;
    }

    socket.send(JSON.stringify({ clientId, text }));
    inputEl.value = "";
    inputEl.focus();
  });

  connect();
})();
