async function loadControlPage() {
    if (!requireAuth()) return;

    const welcomeBox = document.getElementById('welcome-box');
    const user = getUserFromToken();

    if (welcomeBox) {
        welcomeBox.textContent = user?.username
            ? `مرحباً ${user.username}`
            : 'مرحباً بك';
    }

    const adminLink = document.getElementById('admin-link');
    const isAdmin = await ensureAdmin().catch(() => false);
    if (adminLink) {
        adminLink.style.display = isAdmin ? 'flex' : 'none';
    }
}

async function handleLogoutClick() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // حتى لو فشل النداء، سننهي الجلسة محلياً
    }
    logoutLocal();
}

document.addEventListener('DOMContentLoaded', () => {
    loadControlPage();
    document.getElementById('logout-button')?.addEventListener('click', handleLogoutClick);
});