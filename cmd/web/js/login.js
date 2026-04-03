async function handleLoginSubmit(event) {
    event.preventDefault();

    const username = document.getElementById('login-username')?.value.trim();
    const password = document.getElementById('login-password')?.value;

    if (!username || !password) {
        showMessage('error', 'أدخل اسم المستخدم وكلمة المرور', 'auth-message');
        return;
    }

    const button = document.getElementById('login-button');
    if (button) {
        button.disabled = true;
        button.textContent = 'جاري تسجيل الدخول...';
    }

    try {
        const result = await apiFetch('/auth/login', {
            method: 'POST',
            body: JSON.stringify({ username, password })
        }, { skipAuth: true });

        saveSession(result.access_token, result.refresh_token, result.user || null);
        showMessage('success', 'تم تسجيل الدخول بنجاح', 'auth-message');

        setTimeout(() => {
            window.location.href = webPagePath('dashboard.html');
        }, 700);
    } catch (error) {
        showMessage('error', error.message, 'auth-message');
    } finally {
        if (button) {
            button.disabled = false;
            button.textContent = 'دخول';
        }
    }
}

async function handleSignupSubmit(event) {
    event.preventDefault();

    const username = document.getElementById('signup-username')?.value.trim();
    const password = document.getElementById('signup-password')?.value;
    const inviteCode = document.getElementById('signup-invite-code')?.value.trim();

    if (!username || !password || !inviteCode) {
        showMessage('error', 'أكمل جميع الحقول', 'auth-message');
        return;
    }

    const button = document.getElementById('signup-button');
    if (button) {
        button.disabled = true;
        button.textContent = 'جاري التسجيل...';
    }

    try {
        const result = await apiFetch('/auth/signup', {
            method: 'POST',
            body: JSON.stringify({
                username,
                password,
                invite_code: inviteCode
            })
        }, { skipAuth: true });

        saveSession(result.access_token, result.refresh_token, result.user || null);
        showMessage('success', 'تم إنشاء الحساب بنجاح', 'auth-message');

        setTimeout(() => {
            window.location.href = webPagePath('dashboard.html');
        }, 700);
    } catch (error) {
        showMessage('error', error.message, 'auth-message');
    } finally {
        if (button) {
            button.disabled = false;
            button.textContent = 'تسجيل';
        }
    }
}

function switchAuthTab(tab) {
    const loginPanel = document.getElementById('login-panel');
    const signupPanel = document.getElementById('signup-panel');
    const loginTab = document.getElementById('tab-login');
    const signupTab = document.getElementById('tab-signup');

    if (tab === 'signup') {
        loginPanel?.classList.remove('active');
        signupPanel?.classList.add('active');
        loginTab?.classList.remove('active');
        signupTab?.classList.add('active');
    } else {
        signupPanel?.classList.remove('active');
        loginPanel?.classList.add('active');
        signupTab?.classList.remove('active');
        loginTab?.classList.add('active');
    }
}

document.addEventListener('DOMContentLoaded', () => {
    document.getElementById('login-form')?.addEventListener('submit', handleLoginSubmit);
    document.getElementById('signup-form')?.addEventListener('submit', handleSignupSubmit);

    document.getElementById('tab-login')?.addEventListener('click', () => switchAuthTab('login'));
    document.getElementById('tab-signup')?.addEventListener('click', () => switchAuthTab('signup'));

    if (hasAuthToken()) {
        window.location.href = webPagePath('dashboard.html');
    }
});