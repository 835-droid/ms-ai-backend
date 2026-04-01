function getAccessToken() {
    return localStorage.getItem(CONFIG.TOKEN_KEYS.ACCESS);
}

function getRefreshToken() {
    return localStorage.getItem(CONFIG.TOKEN_KEYS.REFRESH);
}

function saveTokens(accessToken, refreshToken) {
    localStorage.setItem(CONFIG.TOKEN_KEYS.ACCESS, accessToken);
    localStorage.setItem(CONFIG.TOKEN_KEYS.REFRESH, refreshToken);
}

function clearTokens() {
    localStorage.removeItem(CONFIG.TOKEN_KEYS.ACCESS);
    localStorage.removeItem(CONFIG.TOKEN_KEYS.REFRESH);
}

function hasAuthToken() {
    return Boolean(getAccessToken());
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
    const token = getAccessToken();
    if (!token) return null;
    return getTokenPayload(token);
}

function logoutLocal() {
    clearTokens();
    window.location.href = '/web/index.html';
}