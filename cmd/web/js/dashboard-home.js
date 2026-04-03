// Dashboard Home Page Logic

let currentPeriod = 'day';

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

function renderRecentlyUpdated(mangas) {
    const container = document.getElementById('recently-updated');
    if (!container) return;

    if (!mangas.length) {
        container.innerHTML = '<p>لا توجد مانجا محدثة</p>';
        return;
    }

    container.innerHTML = mangas.map(manga => {
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';

        return `
            <article class="manga-card horizontal-card">
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

function renderTrending(rankedMangas) {
    const container = document.getElementById('trending');
    if (!container) return;

    if (!rankedMangas.length) {
        container.innerHTML = '<p>لا توجد مانجا رواجة</p>';
        return;
    }

    container.innerHTML = rankedMangas.map((item, index) => {
        const manga = item.manga || item;
        const viewCount = item.view_count || manga.views_count || 0;
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';

        return `
            <article class="manga-card trending-card" data-title="${escapeHtml(manga.title)}" data-id="${escapeHtml(id)}">
                <div class="trending-rank">#${index + 1}</div>
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(id)}">
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">
                            <i class="fas fa-fire trending-fire"></i>
                            ${escapeHtml(manga.title)}
                        </h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-meta">
                            <span>${Array.isArray(manga.tags) ? manga.tags.length : 0} tags</span>
                            <span class="trending-views">👁 ${escapeHtml(formatCompactNumber(viewCount))}</span>
                        </div>
                    </div>
                </a>
            </article>
        `;
    }).join('');
}

function renderMostFollowed(mangas) {
    const container = document.getElementById('most-followed');
    if (!container) return;

    if (!mangas.length) {
        container.innerHTML = '<p>لا توجد مانجا متابعة</p>';
        return;
    }

    container.innerHTML = mangas.map(manga => {
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';
        const favoritesCount = manga.favorites_count || 0;

        return `
            <article class="manga-card" data-title="${escapeHtml(manga.title)}" data-id="${escapeHtml(id)}">
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(id)}">
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">${escapeHtml(manga.title)}</h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-meta">
                            <span>${Array.isArray(manga.tags) ? manga.tags.length : 0} tags</span>
                            <span class="favorites-count">❤️ ${escapeHtml(formatCompactNumber(favoritesCount))}</span>
                        </div>
                    </div>
                </a>
            </article>
        `;
    }).join('');
}

function renderTopRated(mangas) {
    const container = document.getElementById('top-rated');
    if (!container) return;

    if (!mangas.length) {
        container.innerHTML = '<p>لا توجد مانجا مقيمة</p>';
        return;
    }

    container.innerHTML = mangas.map(manga => {
        const id = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';
        const averageRating = manga.average_rating || 0;
        const ratingCount = manga.rating_count || 0;

        return `
            <article class="manga-card" data-title="${escapeHtml(manga.title)}" data-id="${escapeHtml(id)}">
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(id)}">
                    <img class="manga-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-info">
                        <h3 class="manga-title">${escapeHtml(manga.title)}</h3>
                        <p class="manga-description">${escapeHtml(manga.description || '')}</p>
                        <div class="manga-meta">
                            <span>${Array.isArray(manga.tags) ? manga.tags.length : 0} tags</span>
                            <span class="rating-score">⭐ ${escapeHtml(formatRating(averageRating))} (${escapeHtml(formatCompactNumber(ratingCount))})</span>
                        </div>
                    </div>
                </a>
            </article>
        `;
    }).join('');
}

async function loadRecentlyUpdated() {
    try {
        const data = await apiFetch('/mangas/recently-updated?limit=10');
        renderRecentlyUpdated(data);
    } catch (error) {
        console.error('Failed to load recently updated:', error);
        document.getElementById('recently-updated').innerHTML = '<p>فشل في تحميل آخر التحديثات</p>';
    }
}

async function loadTrending(period) {
    try {
        const data = await apiFetch(`/mangas/most-viewed?period=${period}&limit=10`);
        renderTrending(data);
    } catch (error) {
        console.error('Failed to load trending:', error);
        document.getElementById('trending').innerHTML = '<p>فشل في تحميل الأكثر رواجاً</p>';
    }
}

async function loadMostFollowed() {
    try {
        // Load most followed from the dedicated endpoint that returns mangas ordered by favorites_count
        // Use CONFIG.ROUTES.MANGA to compose the path and avoid double /api prefix
        const data = await apiFetch(`${CONFIG.ROUTES.MANGA}/most-followed?limit=10`);
        if (data && Array.isArray(data)) {
            renderMostFollowed(data);
        }
    } catch (error) {
        console.error('Failed to load most followed:', error);
        document.getElementById('most-followed').innerHTML = '<p>فشل في تحميل الأكثر متابعة</p>';
    }
}

async function loadTopRated() {
    try {
        // Load top rated mangas ordered by average_rating
        const data = await apiFetch(`${CONFIG.ROUTES.MANGA}/top-rated?limit=10`);
        if (data && Array.isArray(data)) {
            renderTopRated(data);
        }
    } catch (error) {
        console.error('Failed to load top rated:', error);
        document.getElementById('top-rated').innerHTML = '<p>فشل في تحميل الأعلى تقييماً</p>';
    }
}

async function loadDashboard() {
    console.log('loadDashboard: checking auth');
    if (!requireAuth()) return;

    // Load all sections
    await Promise.all([
        loadRecentlyUpdated(),
        loadTrending(currentPeriod),
        loadMostFollowed(),
        loadTopRated()
    ]);
}

async function handleDashboardLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal(true);
}

// Tab switching for trending
document.addEventListener('DOMContentLoaded', () => {
    loadDashboard();

    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const period = e.target.dataset.period;
            if (!period) return;

            // Update active tab
            document.querySelectorAll('.tab-btn').forEach(b => b.classList.remove('active'));
            e.target.classList.add('active');

            currentPeriod = period;
            loadTrending(period);
        });
    });

    document.getElementById('logout-button')?.addEventListener('click', handleDashboardLogout);
});