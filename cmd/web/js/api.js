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

// ==================== NOVEL API FUNCTIONS ====================

const API = {
    // Manga APIs
    getMangas: (page = 1, limit = 20) => apiFetch(`/mangas?page=${page}&limit=${limit}`),
    getManga: (id) => apiFetch(`/mangas/${id}`),
    getMostViewedMangas: (period = 'day', limit = 10) => apiFetch(`/mangas/most-viewed?period=${period}&limit=${limit}`),
    getRecentlyUpdatedMangas: (limit = 10) => apiFetch(`/mangas/recently-updated?limit=${limit}`),
    getMostFollowedMangas: (limit = 10) => apiFetch(`/mangas/most-followed?limit=${limit}`),
    getTopRatedMangas: (limit = 10) => apiFetch(`/mangas/top-rated?limit=${limit}`),
    
    // Manga engagement
    incrementMangaViews: (id) => apiFetch(`/mangas/${id}/view`, { method: 'POST' }),
    setMangaReaction: (id, type) => apiFetch(`/mangas/${id}/react`, { method: 'POST', body: JSON.stringify({ type }) }),
    getMangaUserReaction: (id) => apiFetch(`/mangas/${id}/my-reaction`),
    rateManga: (id, score) => apiFetch(`/mangas/${id}/rate`, { method: 'POST', body: JSON.stringify({ score }) }),
    addMangaFavorite: (id) => apiFetch(`/mangas/${id}/favorite`, { method: 'POST' }),
    removeMangaFavorite: (id) => apiFetch(`/mangas/${id}/favorite`, { method: 'DELETE' }),
    checkMangaFavorite: (id) => apiFetch(`/mangas/${id}/favorite`),
    listMangaFavorites: (page = 1, limit = 20) => apiFetch(`/mangas/favorites?page=${page}&limit=${limit}`),
    addMangaComment: (id, content) => apiFetch(`/mangas/${id}/comments`, { method: 'POST', body: JSON.stringify({ content }) }),
    getMangaComments: (id, page = 1, limit = 20, sort = 'newest') => apiFetch(`/mangas/${id}/comments?page=${page}&limit=${limit}&sort=${sort}`),
    deleteMangaComment: (id, commentId) => apiFetch(`/mangas/${id}/comments/${commentId}`, { method: 'DELETE' }),

    // Novel APIs
    getNovels: (page = 1, limit = 20) => apiFetch(`/novels?page=${page}&limit=${limit}`),
    getNovel: (id) => apiFetch(`/novels/${id}`),
    getMostViewedNovels: (period = 'day', limit = 10) => apiFetch(`/novels/most-viewed?period=${period}&limit=${limit}`),
    getRecentlyUpdatedNovels: (limit = 10) => apiFetch(`/novels/recently-updated?limit=${limit}`),
    getMostFollowedNovels: (limit = 10) => apiFetch(`/novels/most-followed?limit=${limit}`),
    getTopRatedNovels: (limit = 10) => apiFetch(`/novels/top-rated?limit=${limit}`),
    
    // Novel engagement
    incrementNovelViews: (id) => apiFetch(`/novels/${id}/view`, { method: 'POST' }),
    setNovelReaction: (id, type) => apiFetch(`/novels/${id}/react`, { method: 'POST', body: JSON.stringify({ type }) }),
    getNovelUserReaction: (id) => apiFetch(`/novels/${id}/my-reaction`),
    rateNovel: (id, score) => apiFetch(`/novels/${id}/rate`, { method: 'POST', body: JSON.stringify({ score }) }),
    addNovelFavorite: (id) => apiFetch(`/novels/${id}/favorite`, { method: 'POST' }),
    removeNovelFavorite: (id) => apiFetch(`/novels/${id}/favorite`, { method: 'DELETE' }),
    checkNovelFavorite: (id) => apiFetch(`/novels/${id}/favorite`),
    listNovelFavorites: (page = 1, limit = 20) => apiFetch(`/novels/favorites?page=${page}&limit=${limit}`),
    addNovelComment: (id, content) => apiFetch(`/novels/${id}/comments`, { method: 'POST', body: JSON.stringify({ content }) }),
    getNovelComments: (id, page = 1, limit = 20, sort = 'newest') => apiFetch(`/novels/${id}/comments?page=${page}&limit=${limit}&sort=${sort}`),
    deleteNovelComment: (id, commentId) => apiFetch(`/novels/${id}/comments/${commentId}`, { method: 'DELETE' }),
    
    // Novel chapters
    getNovelChapters: (novelId, page = 1, limit = 100) => apiFetch(`/novels/${novelId}/chapters?page=${page}&limit=${limit}`),
    getNovelChapter: (novelId, chapterId) => apiFetch(`/novels/${novelId}/chapters/${chapterId}`),
    
    // Reading progress
    getNovelReadingProgress: (novelId) => apiFetch(`/novels/${novelId}/progress`, { skipAuth: true }),
    setNovelReadingProgress: (novelId, chapterId) => apiFetch(`/novels/${novelId}/progress`, { method: 'POST', body: JSON.stringify({ chapter_id: chapterId }) }),
};
