// العنوان الثابت للخادم الخلفي (يجب أن يكون هو نفسه المنفذ الذي يعمل عليه Go)
const BACKEND_URL = 'http://localhost:8080';

// المسار الأساسي لملفات الويب – دائماً /web لأن Gin يخدمها عبر r.Static("/web", "./cmd/web")
const WEB_BASE = '/web';

// دالة مساعدة لتكوين مسار كامل لصفحة ويب - تستخدم مسارات نسبية للتوافق
function webPagePath(page) {
    const normalized = String(page || '').replace(/^\/+/, '');
    if (window.location.protocol === 'file:') {
        console.warn('تحذير: يتم تشغيل التطبيق محلياً. يُفضل تشغيل الخادم لتجنب مشاكل التحميل.');
        return normalized; // نسبي، لكن قد لا يعمل في file://
    }
    return normalized; // نسبي، يعمل في HTTP
}

const CONFIG = {
    API_BASE: `${BACKEND_URL}/api`,
    BACKEND_URL,
    WEB_BASE,
    ROUTES: {
        AUTH: '/auth',
        MANGA: '/mangas',
        ADMIN: '/admin'
    },
    TOKEN_KEYS: {
        ACCESS: 'accessToken',
        REFRESH: 'refreshToken',
        USER: 'currentUser'
    },
    DEFAULT_REDIRECT: webPagePath('index.html')
};