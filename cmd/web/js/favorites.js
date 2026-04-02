let favoritesList = [];

function renderFavorites(mangas) {
    const container = document.getElementById('favorites-grid');
    const emptyState = document.getElementById('favorites-empty');
    if (!container) return;

    if (!Array.isArray(mangas) || mangas.length === 0) {
        container.innerHTML = '';
        if (emptyState) emptyState.style.display = 'block';
        return;
    }

    if (emptyState) emptyState.style.display = 'none';

    container.innerHTML = mangas.map(manga => {
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';
        return `
            <article class="manga-card" data-id="${escapeHtml(id)}">
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(id)}">
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">${escapeHtml(manga.title)}</h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-stats">
                            <span class="stat-pill">👁 ${escapeHtml(String(manga.views_count || 0))}</span>
                            <span class="stat-pill">♥ ${escapeHtml(String(manga.likes_count || 0))}</span>
                            <span class="stat-pill">★ ${(Number(manga.average_rating || 0)).toFixed(1)}</span>
                        </div>
                    </div>
                </a>
            </article>
        `;
    }).join('');
}

function applyFavoritesFilter() {
    const query = document.getElementById('favorites-search')?.value.trim().toLowerCase() || '';
    if (!query) {
        renderFavorites(favoritesList);
        return;
    }

    const filtered = favoritesList.filter(m => {
        const title = String(m.title || '').toLowerCase();
        const desc = String(m.description || '').toLowerCase();
        return title.includes(query) || desc.includes(query);
    });
    renderFavorites(filtered);
}

async function loadFavorites() {
    if (!requireAuth()) return;

    setLoading(document.getElementById('favorites-grid'), 'تحميل المفضلة...');

    try {
        const data = await apiFetch('/mangas/favorites/list');
        favoritesList = data.items || [];
        renderFavorites(favoritesList);
    } catch (error) {
        console.error('favorites load error', error);
        setError(document.getElementById('favorites-grid'), error.message);
    }
}

async function handleLogoutFromFavorites() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
    }
    logoutLocal(true);
}

function initFavoritesPage() {
    if (!requireAuth()) return;
    document.getElementById('logout-button')?.addEventListener('click', handleLogoutFromFavorites);
    document.getElementById('favorites-search')?.addEventListener('input', applyFavoritesFilter);
    loadFavorites();
}

document.addEventListener('DOMContentLoaded', initFavoritesPage);