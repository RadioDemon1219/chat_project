const statusEl = document.getElementById('status');
const messagesEl = document.getElementById('messages');
const form = document.getElementById('send-form');
const nameInput = document.getElementById('name');
const textInput = document.getElementById('text');

let ws = null;

function connect() {
    const proto = location.protocol === 'https:' ? 'wss' : 'ws';
    ws = new WebSocket(`${proto}://${location.host}/ws`);

    ws.onopen = () => setStatus('подключён', 'online');
    ws.onclose = () => {
        setStatus('переподключение...', 'offline');
        setTimeout(connect, 1500);
    };
    ws.onmessage = (e) => {
        let msg;
        try { msg = JSON.parse(e.data); } catch { return; }
        appendMessage(msg);
    };
}

function setStatus(text, cls) {
    statusEl.textContent = text;
    statusEl.className = 'status ' + cls;
}

function appendMessage(msg) {
    const li = document.createElement('li');
    li.className = 'message';

    const meta = document.createElement('div');
    meta.className = 'meta';
    meta.textContent = msg.user + (msg.time ? ' · ' + msg.time : '');

    const body = document.createElement('div');
    body.textContent = msg.text;

    li.appendChild(meta);
    li.appendChild(body);
    messagesEl.appendChild(li);
    messagesEl.scrollTop = messagesEl.scrollHeight;
}

form.addEventListener('submit', (e) => {
    e.preventDefault();
    const user = (nameInput.value.trim() || 'аноним').slice(0, 24);
    const text = textInput.value.trim();
    if (!text || !ws || ws.readyState !== WebSocket.OPEN) return;

    ws.send(JSON.stringify({ user, text }));
    textInput.value = '';
    textInput.focus();
});

connect();