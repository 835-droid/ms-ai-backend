let allMangas = [];

function formatCompactNumber(value) {
    const number = Number(value || 0);
    if (number >= 1000000) return `${(number / 1000000).toFixed(1).replace(/\.0$/, '')}M`;
    if (number >= 1000) return `${(number / 1000).toFixed(1).replace(/\.0$/, '')}K`;
    return String(number);
}

function formatRating(value) {
    const number = Number(value || 0);
    return number > 0 ? number.toFixed(1) : '0.0';
}

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
                    ${manga.average_rating > 0 ? `<div class="manga-rating-badge">★ ${escapeHtml(formatRating(manga.average_rating))}</div>` : ''}
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">${escapeHtml(manga.title)}</h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-meta">
                            <span>${Array.isArray(manga.tags) ? manga.tags.length : 0} tags</span>
                        </div>
                        
                        ${manga.reactions_count ? `
                            <div class="manga-reactions-preview">
                                ${Object.entries(manga.reactions_count)
                                    .filter(([_, count]) => count > 0)
                                    .sort((a, b) => b[1] - a[1])
                                    .slice(0, 3)
                                    .map(([type, count]) => {
                                        const emojis = {
                                            upvote: '👍',
                                            funny: '😂',
                                            love: '❤️',
                                            surprised: '😮',
                                            angry: '😡',
                                            sad: '😢'
                                        };
                                        return `<div class="manga-reaction-badge"><span class="manga-reaction-emoji">${emojis[type] || '👍'}</span> ${formatCompactNumber(count)}</div>`;
                                    }).join('')}
                            </div>
                        ` : ''}
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

async function loadLibrary() {
    console.log('loadLibrary: checking auth');
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

async function handleLibraryLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal(true);
}

document.addEventListener('DOMContentLoaded', () => {
    loadLibrary();
    document.getElementById('search-input')?.addEventListener('input', applySearchFilter);
    document.getElementById('logout-button')?.addEventListener('click', handleLibraryLogout);
});