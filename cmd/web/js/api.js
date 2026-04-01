async function parseApiResponse(response) {
    const text = await response.text();
    if (!text) return null;
    try {
        return JSON.parse(text);
    } catch {
        return { raw: text };
    }
}

async function refreshSession() {
    const refreshToken = getRefreshToken();
    if (!refreshToken) return false;

    const response = await fetch(`${CONFIG.API_BASE}${CONFIG.ROUTES.AUTH}/refresh`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refresh_token: refreshToken })
    });

    const data = await parseApiResponse(response);

    if (!response.ok) {
        clearTokens();
        return false;
    }

    const payload = data?.data || data || {};
    if (payload.access_token && payload.refresh_token) {
        saveTokens(payload.access_token, payload.refresh_token);
        return true;
    }

    clearTokens();
    return false;
}

async function apiFetch(path, options = {}, meta = {}) {
    const { skipAuth = false, retryOnAuth = true } = meta;

    const headers = new Headers(options.headers || {});
    const isFormData = options.body instanceof FormData;

    if (!isFormData && !headers.has('Content-Type') && options.body !== undefined) {
        headers.set('Content-Type', 'application/json');
    }

    if (!skipAuth) {
        const token = getAccessToken();
        if (token) {
            headers.set('Authorization', `Bearer ${token}`);
        }
    }

    const response = await fetch(`${CONFIG.API_BASE}${path}`, {
        ...options,
        headers
    });

    const data = await parseApiResponse(response);

    if (response.status === 401 && retryOnAuth && !skipAuth) {
        const refreshed = await refreshSession();
        if (refreshed) {
            return apiFetch(path, options, { skipAuth, retryOnAuth: false });
        }
        clearTokens();
        window.location.href = CONFIG.DEFAULT_REDIRECT;
        throw new Error('Authentication expired');
    }

    if (!response.ok) {
        const message = data?.error || data?.message || 'Request failed';
        throw new Error(message);
    }

    return data?.data !== undefined ? data.data : data;
}