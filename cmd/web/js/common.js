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

function normalizeImageUrl(value) {
    return String(value || '')
        .trim()
        .replace(/^["']+|["']+$/g, '')
        .trim();
}

function toTagArray(text) {
    return String(text || '')
        .split(',')
        .map(t => t.trim())
        .filter(Boolean);
}

function requireAuth(redirectUrl = webPagePath('index.html')) {
    if (!hasAuthToken()) {
        window.location.href = redirectUrl;
        return false;
    }
    const token = getAccessToken();
    if (token) {
        try {
            const payload = getTokenPayload(token);
            if (payload && payload.exp) {
                const expiry = payload.exp * 1000;
                if (Date.now() >= expiry) {
                    clearTokens();
                    window.location.href = redirectUrl;
                    return false;
                }
            }
        } catch {
            // تجاهل أخطاء فك التشفير هنا؛ API سيتعامل مع فشل المصادقة.
        }
    }
    return true;
}

function getUserRolesFromToken() {
    const user = getUserFromToken();
    return user?.roles || [];
}

function isAdminFromToken() {
    const roles = getUserRolesFromToken();
    return roles.includes('admin');
}

async function refreshAndCheckAdmin() {
    try {
        const refreshed = await refreshSession();
        if (refreshed) {
            return isAdminFromToken();
        }
        return false;
    } catch (e) {
        console.error('Refresh failed in ensureAdmin', e);
        return false;
    }
}

// ensureAdmin محسنة
async function ensureAdmin() {
    if (isAdminFromToken()) {
        console.log('ensureAdmin: admin from token');
        return true;
    }

    console.log('ensureAdmin: token does not have admin role, attempting refresh');
    const refreshed = await refreshAndCheckAdmin();
    if (refreshed) {
        console.log('ensureAdmin: after refresh, admin role found');
        return true;
    }

    try {
        console.log('ensureAdmin: calling /admin/check');
        await apiFetch('/admin/check');
        console.log('ensureAdmin: server check passed');
        return true;
    } catch (err) {
        console.error('ensureAdmin: server check failed', err);
        // Distinguish between 403 (not admin) and other errors (network/500)
        if (err.status === 403) {
            // User is not admin
            return false;
        } else {
            // Network or server error - throw error instead of returning false
            throw new Error('admin_check_network_error');
        }
    }
}

// زر العودة للأعلى العالمي
function initGlobalScrollToTop() {
    const btn = document.createElement('button');
    btn.id = 'global-scroll-top';
    btn.innerHTML = '↑';
    btn.style.cssText = `
        position: fixed;
        bottom: 20px;
        right: 20px;
        width: 45px;
        height: 45px;
        border-radius: 50%;
        background: var(--primary);
        color: white;
        border: none;
        cursor: pointer;
        opacity: 0;
        transition: opacity 0.3s;
        z-index: 1000;
        font-size: 1.5rem;
        display: flex;
        align-items: center;
        justify-content: center;
        box-shadow: 0 4px 12px rgba(0,0,0,0.2);
    `;
    document.body.appendChild(btn);
    
    window.addEventListener('scroll', () => {
        btn.style.opacity = window.scrollY > 300 ? '1' : '0';
    });
    
    btn.addEventListener('click', () => {
        window.scrollTo({ top: 0, behavior: 'smooth' });
    });
}

document.addEventListener('DOMContentLoaded', initGlobalScrollToTop);

// ========== Dark Mode Toggle ==========
function initDarkModeToggle() {
    if (window._darkModeToggleInitialized) return;
    window._darkModeToggleInitialized = true;
    
    // Apply saved/default dark mode to body FIRST, regardless of toggle button existence
    const savedMode = localStorage.getItem('darkMode');
    const isDarkMode = savedMode === null ? true : savedMode === 'true';
    
    if (isDarkMode) {
        document.body.classList.add('dark-mode');
    }
    
    // Only register click listener if toggle button exists
    const toggle = document.querySelector('.dark-mode-toggle');
    if (!toggle) return;
    
    toggle.addEventListener('click', () => {
        document.body.classList.toggle('dark-mode');
        const now = document.body.classList.contains('dark-mode');
        localStorage.setItem('darkMode', now ? 'true' : 'false');
    });
}

// ========== Navbar Setup ==========
function initNavbar() {
    if (window._navbarInitialized) return;
    window._navbarInitialized = true;
    
    // Get user info from token
    const user = getUserFromToken();
    const initials = user?.username ? user.username.substring(0, 2).toUpperCase() : 'US';
    
    // Setup navbar user badge
    const userBadge = document.querySelector('.navbar-user-badge');
    if (userBadge) {
        const displayName = user?.username || 'مستخدم';
        const avatar = userBadge.querySelector('.navbar-user-avatar');
        if (avatar) {
            avatar.textContent = initials;
        }
    }
    
    // Mobile menu toggle (if implemented)
    const toggle = document.querySelector('.navbar-toggle');
    const menu = document.querySelector('.navbar-menu');
    if (toggle && menu) {
        toggle.addEventListener('click', () => {
            menu.classList.toggle('active');
        });
    }
}

document.addEventListener('DOMContentLoaded', () => {
    initDarkModeToggle();
    initNavbar();
});