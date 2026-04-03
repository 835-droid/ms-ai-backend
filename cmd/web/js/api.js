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
    if (!refreshToken) {
        console.log('No refresh token');
        return false;
    }

    console.log('Attempting to refresh session');
    try {
        const response = await fetch(`${CONFIG.API_BASE}${CONFIG.ROUTES.AUTH}/refresh`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ refresh_token: refreshToken })
        });

        const data = await parseApiResponse(response);

        if (!response.ok) {
            console.log('Refresh failed with status', response.status);
            // إذا كان الخادم يرفض التوكن، نمسحه
            clearTokens();
            return false;
        }

        const payload = data?.data || data || {};
        if (payload.access_token && payload.refresh_token) {
            console.log('Refresh successful');
            saveSession(payload.access_token, payload.refresh_token, getStoredUser());
            return true;
        }

        clearTokens();
        return false;
    } catch (err) {
        console.error('Refresh network error', err);
        clearTokens();
        return false;
    }
}

async function apiFetch(path, options = {}, meta = {}) {
    const { skipAuth = false, retryOnAuth = true, noRedirect = false } = meta;

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

    console.log(`apiFetch: ${path}`, { method: options.method, skipAuth, noRedirect });

    let response = await fetch(`${CONFIG.API_BASE}${path}`, {
        ...options,
        headers
    });

    // معالجة 401 مع إعادة محاولة التحديث مرة واحدة
    if (response.status === 401 && retryOnAuth && !skipAuth) {
        console.log('Got 401, attempting refresh');
        const refreshed = await refreshSession();
        if (refreshed) {
            // إعادة المحاولة مع التوكن الجديد
            const newToken = getAccessToken();
            headers.set('Authorization', `Bearer ${newToken}`);
            response = await fetch(`${CONFIG.API_BASE}${path}`, {
                ...options,
                headers
            });
        } else {
            // فشل التحديث – لا نعيد التوجيه تلقائياً، بل نرمي خطأ واضح
            if (!noRedirect) {
                // نسمح للكود المتصل باتخاذ القرار، ولكننا نظهر رسالة خطأ
                console.error('Session expired and refresh failed');
                throw new Error('انتهت صلاحية الجلسة. يرجى تسجيل الدخول مرة أخرى.');
            } else {
                throw new Error('Authentication required');
            }
        }
    }

    const data = await parseApiResponse(response);

    if (!response.ok) {
        const message = data?.error || data?.message || 'Request failed';
        console.error('apiFetch error', response.status, message);
        const error = new Error(message);
        error.status = response.status;
        error.responseBody = data;
        throw error;
    }

    return data?.data !== undefined ? data.data : data;
}