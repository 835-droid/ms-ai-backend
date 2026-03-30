// app.js
// هذا الملف يحتوي على الدوال العامة والمساعدة والمنطق الخاص بصفحات المصادقة والإدارة.

const BASE_URL = 'http://localhost:8080/api/auth'; // لصفحات المصادقة (Auth)
const ADMIN_URL = 'http://localhost:8080/api/admin'; // لصفحات الإدارة (Admin)

// -----------------------------------------------------
// الدوال المساعدة للواجهة والتخزين (عامة لجميع الصفحات)
// -----------------------------------------------------

// عناصر نموذجية (قد لا تكون موجودة في كل صفحة)
const loginForm = document.getElementById('login-form');
const signupForm = document.getElementById('signup-form');
// للرسائل العامة في صفحات Auth/Admin
const messageElementId = document.getElementById('auth-message') ? 'auth-message' : 'admin-message';

/**
 * لعرض رسائل التنبيه والخطأ (تستخدم في كل الصفحات ما عدا الشات).
 */
function showMessage(type, text, elementId = messageElementId) {
    const messageElement = document.getElementById(elementId);
    if (!messageElement) return;

    messageElement.textContent = text;
    messageElement.className = `message ${type}`;
    messageElement.style.display = 'block';
    setTimeout(() => {
        messageElement.style.display = 'none';
    }, 5000);
}

/**
 * لعرض التنبيهات الصغيرة.
 */
function showNotification(message, type = 'info') {
    const notif = document.createElement('div');
    notif.className = `notification ${type}`;
    notif.setAttribute('role', 'status');
    notif.textContent = message;
    notif.style.cssText = 'position:fixed;top:20px;left:20px;padding:10px 14px;border-radius:6px;z-index:9999;';
    if (type === 'error') notif.style.background = '#e74c3c';
    else if (type === 'success') notif.style.background = '#27ae60';
    else notif.style.background = '#3498db';
    notif.style.color = 'white';
    document.body.appendChild(notif);
    setTimeout(() => notif.remove(), 4000);
}

function saveTokens(accessToken, refreshToken) {
    localStorage.setItem('accessToken', accessToken);
    localStorage.setItem('refreshToken', refreshToken);
}

function getAccessToken() {
    return localStorage.getItem('accessToken');
}

function getRefreshToken() {
    return localStorage.getItem('refreshToken');
}

function clearTokens() {
    localStorage.removeItem('accessToken');
    localStorage.removeItem('refreshToken');
}

/**
 * دالة التحقق من المصادقة عبر طلب للباك اند.
 */
async function checkAuthentication(redirectUrl = 'index.html') {
    const token = getAccessToken();
    if (!token) {
        window.location.href = redirectUrl;
        return;
    }

    try {
        const response = await fetch(`${BASE_URL}/verify`, {
            method: 'GET',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        if (!response.ok) {
            clearTokens();
            window.location.href = redirectUrl;
        }
    } catch (err) {
        console.warn('Authentication check failed:', err);
        clearTokens();
        window.location.href = redirectUrl;
    }
}

// -----------------------------------------------------
// منطق المصادقة (Authentication Logic)
// -----------------------------------------------------

async function handleLogin(event) {
    event.preventDefault();

    const usernameInput = document.getElementById('username');
    const passwordInput = document.getElementById('password');

    if (!usernameInput || !passwordInput) return;

    const username = usernameInput.value.trim();
    const password = passwordInput.value;

    if (!username || !password) {
        showMessage('error', 'يرجى إدخال اسم المستخدم وكلمة المرور');
        return;
    }

    try {
        const loginBtn = event.target.querySelector('button');
        if (loginBtn) {
            loginBtn.disabled = true;
            loginBtn.textContent = 'جاري التحقق...';
        }

        const response = await fetch(`${BASE_URL}/login`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ username, password })
        });

        const responseData = await response.json();

        if (response.ok && responseData.access_token) {
            saveTokens(responseData.access_token, responseData.refresh_token);
            showMessage('success', 'تم تسجيل الدخول بنجاح! جاري التحويل...');
            setTimeout(() => {
                window.location.href = 'dashboard.html';
            }, 1500);
        } else {
            throw new Error(responseData.message || 'اسم المستخدم أو كلمة المرور غير صحيحة');
        }

    } catch (error) {
        console.error('Login Error:', error);
        showMessage('error', error.message);
        
        const loginBtn = event.target.querySelector('button');
        if (loginBtn) {
            loginBtn.disabled = false;
            loginBtn.textContent = 'تسجيل الدخول';
        }
    }
}

async function handleSignup(e) {
    e.preventDefault();
    const username = document.getElementById('signup-username').value;
    const password = document.getElementById('signup-password').value;
    const inviteCode = document.getElementById('signup-invite-code').value;
    const button = document.getElementById('signup-button');

    button.disabled = true;
    button.textContent = 'جاري التسجيل...';

    try {
        const response = await fetch(`${BASE_URL}/signup`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json'
            },
            body: JSON.stringify({ username, password, invite_code: inviteCode })
        });

        const responseData = await response.json();

        if (response.ok && responseData.success) {
            saveTokens(responseData.access_token, responseData.refresh_token);
            showMessage('success', 'تم التسجيل بنجاح! جاري التوجيه...');
            setTimeout(() => {
                window.location.href = 'dashboard.html';
            }, 1000);
        } else {
            const errorMessage = responseData.error || 'فشل التسجيل. تحقق من رمز الدعوة.';
            showMessage('error', errorMessage);
        }
    } catch (error) {
        showMessage('error', 'فشل الاتصال بخادم المصادقة.');
        console.error('Signup error:', error);
    } finally {
        button.disabled = false;
        button.textContent = 'تسجيل';
    }
}

async function handleLogout() {
    const token = getAccessToken();
    if (token) {
        try {
            await fetch(`${BASE_URL}/logout`, {
                method: 'POST',
                headers: {
                    'Authorization': `Bearer ${token}`
                }
            });
        } catch (err) {
            console.warn('Logout request failed:', err);
        }
    }

    clearTokens();
    window.location.href = 'index.html';
}

function setupAuthListeners() {
    if (loginForm) {
        loginForm.addEventListener('submit', handleLogin);
    }
    if (signupForm) {
        signupForm.addEventListener('submit', handleSignup);
    }
    
    const showSignupLink = document.getElementById('show-signup');
    const showLoginLink = document.getElementById('show-login');
    
    if (showSignupLink && showLoginLink && loginForm && signupForm) {
        showSignupLink.addEventListener('click', (e) => {
            e.preventDefault();
            loginForm.style.display = 'none';
            signupForm.style.display = 'block';
        });
        showLoginLink.addEventListener('click', (e) => {
            e.preventDefault();
            signupForm.style.display = 'none';
            loginForm.style.display = 'block';
        });
    }

    const logoutButton = document.getElementById('logout-button');
    if (logoutButton) {
        logoutButton.addEventListener('click', handleLogout);
    }
}

// -----------------------------------------------------
// منطق الإدارة (Admin Logic)
// -----------------------------------------------------

async function addContent(type, title, description) {
    const button = document.getElementById('add-content-button');
    const token = getAccessToken();

    if (!token) {
        showMessage('error', 'يجب تسجيل الدخول أولاً.');
        return;
    }

    button.disabled = true;
    button.textContent = 'جاري الإضافة...';

    try {
        const response = await fetch(`${ADMIN_URL}/content`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({ type, title, description })
        });

        const responseData = await response.json();

        if (response.ok) {
            showMessage('success', `${title} (نوع: ${type}) تمت إضافته بنجاح!`);
            document.getElementById('add-content-form').reset();
        } else {
            const errorMessage = responseData.error || 'خطأ أثناء إضافة المحتوى.';
            showMessage('error', errorMessage);
        }
    } catch (error) {
        showMessage('error', 'فشل الاتصال بخادم الإدارة.');
        console.error('Fetch error:', error);
    } finally {
        if (button) { button.disabled = false; button.textContent = 'إضافة المحتوى'; }
    }
}

function setupAdminListeners() {
    const addContentForm = document.getElementById('add-content-form');
    if (addContentForm) {
        addContentForm.addEventListener('submit', function(e) {
            e.preventDefault();
            const type = document.getElementById('content-type').value;
            const title = document.getElementById('content-title').value;
            const description = document.getElementById('content-description').value;

            if (type && title && description) {
                addContent(type, title, description);
            } else {
                showMessage('error', 'الرجاء ملء جميع الحقول واختيار نوع المحتوى.');
            }
        });
    }
}

// -----------------------------------------------------
// إعداد مستمعات الأحداث الرئيسية عند تحميل الصفحة
// -----------------------------------------------------
window.onload = function() {
    if (loginForm || signupForm || document.getElementById('logout-button')) {
        setupAuthListeners();
    }
    
    if (document.getElementById('add-content-form')) {
        setupAdminListeners();
    }
};