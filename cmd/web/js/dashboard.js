let allMangas = [];

function renderMangas(mangas) {
    const container = document.getElementById('manga-grid');
    const emptyState = document.getElementById('empty-state');

    if (!container) return;

    if (!mangas.length) {
        container.innerHTML = '';
        if (emptyState) emptyState.style.display = 'block';
        return;
    }

    if (emptyState) emptyState.style.display = 'none';

    container.innerHTML = mangas.map(manga => {
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';

        return `
            <article class="manga-card" data-title="${escapeHtml(manga.title)}" data-id="${escapeHtml(id)}">
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(id)}">
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">${escapeHtml(manga.title)}</h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-meta">
                            <span>${Array.isArray(manga.tags) ? manga.tags.length : 0} tags</span>
                        </div>
                    </div>
                </a>
            </article>
        `;
    }).join('');
}

function applySearchFilter() {
    const query = document.getElementById('search-input')?.value.trim().toLowerCase() || '';
    const filtered = !query
        ? allMangas
        : allMangas.filter(m => {
            const title = String(m.title || '').toLowerCase();
            const desc = String(m.description || '').toLowerCase();
            const tags = Array.isArray(m.tags) ? m.tags.join(' ').toLowerCase() : '';
            return title.includes(query) || desc.includes(query) || tags.includes(query);
        });

    renderMangas(filtered);
}

async function loadDashboard() {
    if (!requireAuth()) return;

    const container = document.getElementById('manga-grid');
    setLoading(container, 'جاري تحميل المانجا...');

    try {
        const data = await apiFetch('/mangas?limit=100');
        allMangas = data.items || [];
        renderMangas(allMangas);
    } catch (error) {
        setError(container, error.message);
    }
}

async function handleDashboardLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal();
}

document.addEventListener('DOMContentLoaded', () => {
    loadDashboard();
    document.getElementById('search-input')?.addEventListener('input', applySearchFilter);
    document.getElementById('logout-button')?.addEventListener('click', handleDashboardLogout);
});