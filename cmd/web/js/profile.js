async function loadProfile() {
    if (!requireAuth()) return;

    const user = getUserFromToken();
    if (!user) {
        setError(document.getElementById('profile-info'), 'تعذر الحصول على بيانات المستخدم');
        return;
    }

    document.getElementById('profile-username').textContent = user.username || user.id || 'غير معروف';
    document.getElementById('profile-roles').textContent = (Array.isArray(user.roles) ? user.roles.join(', ') : 'مستخدم') || 'مستخدم';
    document.getElementById('profile-userid').textContent = user.id || '-';
}

async function handleProfileLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal(true);
}

async function handleChangePassword(event) {
    event.preventDefault();

    const currentPassword = document.getElementById('current-password').value;
    const newPassword = document.getElementById('new-password').value;
    const messageBox = document.getElementById('password-message');

    if (!currentPassword || !newPassword) {
        showMessage('error', 'يرجى إدخال كلمة المرور الحالية والجديدة', 'password-message');
        return;
    }

    try {
        await apiFetch('/auth/password', {
            method: 'PUT',
            body: JSON.stringify({ current_password: currentPassword, new_password: newPassword })
        });
        showMessage('success', 'تم تغيير كلمة المرور بنجاح', 'password-message');
        document.getElementById('current-password').value = '';
        document.getElementById('new-password').value = '';
    } catch (error) {
        showMessage('error', error.message, 'password-message');
    }
}

function initProfilePage() {
    loadProfile();
    document.getElementById('logout-button')?.addEventListener('click', handleProfileLogout);
    document.getElementById('change-password-form')?.addEventListener('submit', handleChangePassword);
}

document.addEventListener('DOMContentLoaded', initProfilePage);