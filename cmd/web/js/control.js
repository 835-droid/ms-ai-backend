async function loadControlPage() {
    console.log('loadControlPage: checking auth');
    if (!requireAuth()) return;

    const welcomeBox = document.getElementById('welcome-box');
    const user = getUserFromToken();

    if (welcomeBox) {
        welcomeBox.textContent = user?.username
            ? `مرحباً ${user.username}`
            : 'مرحباً بك';
    }

    const adminLink = document.getElementById('admin-link');
    try {
        const isAdmin = await ensureAdmin();
        console.log('isAdmin:', isAdmin);
        if (adminLink) {
            adminLink.style.display = isAdmin ? 'flex' : 'none';
        }
    } catch (e) {
        console.warn('Error checking admin:', e);
        if (adminLink) adminLink.style.display = 'none';
    }
}

async function handleLogoutClick() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal(true);
}

document.addEventListener('DOMContentLoaded', () => {
    loadControlPage();
    document.getElementById('logout-button')?.addEventListener('click', handleLogoutClick);
});