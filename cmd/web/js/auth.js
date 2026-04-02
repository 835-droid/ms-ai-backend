function getAccessToken() {
    return localStorage.getItem(CONFIG.TOKEN_KEYS.ACCESS);
}

function getRefreshToken() {
    return localStorage.getItem(CONFIG.TOKEN_KEYS.REFRESH);
}

function getStoredUser() {
    const raw = localStorage.getItem(CONFIG.TOKEN_KEYS.USER);
    if (!raw) return null;
    try {
        return JSON.parse(raw);
    } catch {
        localStorage.removeItem(CONFIG.TOKEN_KEYS.USER);
        return null;
    }
}

function saveTokens(accessToken, refreshToken) {
    localStorage.setItem(CONFIG.TOKEN_KEYS.ACCESS, accessToken);
    localStorage.setItem(CONFIG.TOKEN_KEYS.REFRESH, refreshToken);
}

function saveSession(accessToken, refreshToken, user = null) {
    saveTokens(accessToken, refreshToken);
    if (user) {
        localStorage.setItem(CONFIG.TOKEN_KEYS.USER, JSON.stringify(user));
    }
}

function clearTokens() {
    localStorage.removeItem(CONFIG.TOKEN_KEYS.ACCESS);
    localStorage.removeItem(CONFIG.TOKEN_KEYS.REFRESH);
    localStorage.removeItem(CONFIG.TOKEN_KEYS.USER);
}

function hasAuthToken() {
    const token = getAccessToken();
    return Boolean(token);
}

function getTokenPayload(token) {
    try {
        const parts = token.split('.');
        if (parts.length !== 3) return null;
        const base64 = parts[1].replace(/-/g, '+').replace(/_/g, '/');
        const json = atob(base64);
        return JSON.parse(decodeURIComponent(Array.prototype.map.call(json, c =>
            '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2)
        ).join('')));
    } catch {
        return null;
    }
}

function getUserFromToken() {
    const storedUser = getStoredUser();
    if (storedUser) {
        return storedUser;
    }

    const token = getAccessToken();
    if (!token) return null;

    const payload = getTokenPayload(token);
    if (!payload) return null;

    return {
        id: payload.user_id || null,
        roles: Array.isArray(payload.roles) ? payload.roles : []
    };
}

function logoutLocal(redirect = true) {
    clearTokens();
    if (redirect) {
        window.location.href = webPagePath('index.html');
    }
}