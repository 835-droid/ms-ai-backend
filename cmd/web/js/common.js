function qs(selector, root = document) {
    return root.querySelector(selector);
}

function qsa(selector, root = document) {
    return Array.from(root.querySelectorAll(selector));
}

function escapeHtml(value) {
    return String(value ?? '')
        .replaceAll('&', '&amp;')
        .replaceAll('<', '&lt;')
        .replaceAll('>', '&gt;')
        .replaceAll('"', '&quot;')
        .replaceAll("'", '&#39;');
}

function formatDate(value) {
    if (!value) return '—';
    const date = new Date(value);
    if (Number.isNaN(date.getTime())) return '—';
    return date.toLocaleDateString('ar-EG');
}

function getQueryParam(name) {
    return new URLSearchParams(window.location.search).get(name);
}

function setLoading(container, text = 'جاري التحميل...') {
    if (!container) return;
    container.innerHTML = `
        <div class="loading">
            <div class="loading-spinner"></div>
            <p>${escapeHtml(text)}</p>
        </div>
    `;
}

function setError(container, text = 'حدث خطأ غير متوقع') {
    if (!container) return;
    container.innerHTML = `<div class="error">${escapeHtml(text)}</div>`;
}

function showMessage(type, text, elementId = 'message-box') {
    const box = document.getElementById(elementId);
    if (!box) return;

    box.textContent = text;
    box.className = `message ${type}`;
    box.style.display = 'block';

    clearTimeout(box._hideTimer);
    box._hideTimer = setTimeout(() => {
        box.style.display = 'none';
    }, 4000);
}

function splitLines(text) {
    return String(text || '')
        .split('\n')
        .map(line => line.trim())
        .filter(Boolean);
}

function toTagArray(text) {
    return String(text || '')
        .split(',')
        .map(t => t.trim())
        .filter(Boolean);
}

function requireAuth(redirectUrl = '/web/index.html') {
    if (!hasAuthToken()) {
        window.location.href = redirectUrl;
        return false;
    }
    return true;
}

async function ensureAdmin() {
    try {
        await apiFetch('/admin/check');
        return true;
    } catch {
        return false;
    }
}