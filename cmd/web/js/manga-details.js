let currentManga = null;
let currentChapters = [];

async function loadMangaDetailsPage() {
    if (!requireAuth()) return;

    const mangaId = getQueryParam('id');
    const detailsContainer = document.getElementById('manga-details');
    const chaptersContainer = document.getElementById('chapters-container');

    if (!mangaId) {
        setError(detailsContainer, 'معرّف المانجا مفقود');
        return;
    }

    setLoading(detailsContainer, 'جاري تحميل تفاصيل المانجا...');
    setLoading(chaptersContainer, 'جاري تحميل الفصول...');

    try {
        const manga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
        currentManga = manga;

        detailsContainer.innerHTML = `
            <div class="details-layout">
                <div class="details-cover">
                    <img src="${escapeHtml(manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga')}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                </div>
                <div class="details-body">
                    <h1>${escapeHtml(manga.title)}</h1>
                    <p class="details-description">${escapeHtml(manga.description || 'لا يوجد وصف')}</p>
                    <div class="details-tags">
                        ${(Array.isArray(manga.tags) ? manga.tags : []).map(tag => `<span class="tag-item">${escapeHtml(tag)}</span>`).join('')}
                    </div>
                    <div class="details-meta">
                        <span>تاريخ الإضافة: ${formatDate(manga.created_at)}</span>
                    </div>
                </div>
            </div>
        `;

        const chapterResp = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        currentChapters = chapterResp.chapters || [];

        if (!currentChapters.length) {
            chaptersContainer.innerHTML = `<div class="empty-state">لا توجد فصول بعد</div>`;
            return;
        }

        chaptersContainer.innerHTML = `
            <div class="chapters-header">
                <h2>الفصول</h2>
                <span>${currentChapters.length} فصل</span>
            </div>
            <div class="chapters-list">
                ${currentChapters.map(chapter => {
                    const chapterId = chapter.id || chapter._id;
                    return `
                        <div class="chapter-row">
                            <div class="chapter-row-main">
                                <strong>الفصل ${escapeHtml(chapter.number)}</strong>
                                <span>${escapeHtml(chapter.title || '')}</span>
                                <small>${Array.isArray(chapter.pages) ? chapter.pages.length : 0} صفحة</small>
                            </div>
                            <a class="btn btn-secondary" href="manga-reader.html?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}">اقرأ</a>
                        </div>
                    `;
                }).join('')}
            </div>
        `;
    } catch (error) {
        setError(detailsContainer, error.message);
    } finally {
        if (chaptersContainer && !currentChapters.length) {
            chaptersContainer.innerHTML = '';
        }
    }
}

async function handleDetailsLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal();
}

document.addEventListener('DOMContentLoaded', () => {
    loadMangaDetailsPage();
    document.getElementById('logout-button')?.addEventListener('click', handleDetailsLogout);
});