let readerState = {
    mangaId: null,
    chapterId: null,
    manga: null,
    chapters: [],
    chapter: null,
    pageIndex: 0
};

function renderReaderPage() {
    const pageImage = document.getElementById('reader-page-image');
    const pageCounter = document.getElementById('page-counter');
    const mangaTitle = document.getElementById('reader-manga-title');
    const chapterTitle = document.getElementById('reader-chapter-title');
    const pages = readerState.chapter?.pages || [];

    if (mangaTitle) mangaTitle.textContent = readerState.manga?.title || 'القراءة';
    if (chapterTitle) chapterTitle.textContent = readerState.chapter ? `الفصل ${readerState.chapter.number} - ${readerState.chapter.title}` : '';

    if (!pages.length) {
        pageImage.src = '';
        pageImage.alt = 'لا توجد صفحات';
        pageCounter.textContent = '0 / 0';
        document.getElementById('reader-error').textContent = 'لا توجد صفحات داخل هذا الفصل';
        return;
    }

    const safeIndex = Math.max(0, Math.min(readerState.pageIndex, pages.length - 1));
    readerState.pageIndex = safeIndex;

    pageImage.src = pages[safeIndex];
    pageImage.alt = `صفحة ${safeIndex + 1}`;
    pageCounter.textContent = `${safeIndex + 1} / ${pages.length}`;

    document.getElementById('prev-page-btn').disabled = safeIndex <= 0;
    document.getElementById('next-page-btn').disabled = safeIndex >= pages.length - 1;
    document.getElementById('reader-error').textContent = '';
}

function goToPage(index) {
    readerState.pageIndex = index;
    renderReaderPage();
}

function goPrevPage() {
    if (readerState.pageIndex > 0) goToPage(readerState.pageIndex - 1);
}

function goNextPage() {
    const pages = readerState.chapter?.pages || [];
    if (readerState.pageIndex < pages.length - 1) goToPage(readerState.pageIndex + 1);
}

function findChapterIndex(chapterId) {
    return readerState.chapters.findIndex(ch => (ch.id || ch._id) === chapterId);
}

function loadChapterByIndex(index) {
    const chapter = readerState.chapters[index];
    if (!chapter) return;

    readerState.chapter = chapter;
    readerState.chapterId = chapter.id || chapter._id;
    readerState.pageIndex = 0;
    renderReaderPage();

    const chapterSelect = document.getElementById('chapter-select');
    if (chapterSelect) chapterSelect.value = readerState.chapterId;

    const url = new URL(window.location.href);
    url.searchParams.set('chapterId', readerState.chapterId);
    url.searchParams.delete('page');
    history.replaceState({}, '', url);
}

async function loadReaderData() {
    if (!requireAuth()) return;

    const mangaId = getQueryParam('mangaId');
    const chapterId = getQueryParam('chapterId');
    const pageParam = Number(getQueryParam('page') || 1);

    if (!mangaId || !chapterId) {
        document.getElementById('reader-error').textContent = 'الرابط غير مكتمل';
        return;
    }

    readerState.mangaId = mangaId;
    readerState.chapterId = chapterId;
    readerState.pageIndex = Number.isFinite(pageParam) && pageParam > 0 ? pageParam - 1 : 0;

    try {
        readerState.manga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
        const chapterResp = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        readerState.chapters = chapterResp.chapters || [];

        const currentIndex = findChapterIndex(chapterId);
        if (currentIndex === -1) {
            document.getElementById('reader-error').textContent = 'الفصل غير موجود';
            return;
        }

        readerState.chapter = readerState.chapters[currentIndex];

        const chapterSelect = document.getElementById('chapter-select');
        if (chapterSelect) {
            chapterSelect.innerHTML = readerState.chapters.map(ch => {
                const id = ch.id || ch._id;
                return `<option value="${escapeHtml(id)}">الفصل ${escapeHtml(ch.number)} - ${escapeHtml(ch.title || '')}</option>`;
            }).join('');
            chapterSelect.value = chapterId;
        }

        renderReaderPage();
    } catch (error) {
        document.getElementById('reader-error').textContent = error.message;
    }
}

async function changeChapterFromSelect() {
    const selectedId = document.getElementById('chapter-select')?.value;
    const index = findChapterIndex(selectedId);
    if (index >= 0) loadChapterByIndex(index);
}

async function handleReaderLogout() {
    try {
        await apiFetch('/auth/logout', { method: 'POST' });
    } catch {
        // ignore
    }
    logoutLocal();
}

document.addEventListener('DOMContentLoaded', () => {
    loadReaderData();
    document.getElementById('prev-page-btn')?.addEventListener('click', goPrevPage);
    document.getElementById('next-page-btn')?.addEventListener('click', goNextPage);
    document.getElementById('chapter-select')?.addEventListener('change', changeChapterFromSelect);
    document.getElementById('logout-button')?.addEventListener('click', handleReaderLogout);

    document.addEventListener('keydown', (event) => {
        if (event.key === 'ArrowLeft' || event.key === 'PageDown') goNextPage();
        if (event.key === 'ArrowRight' || event.key === 'PageUp') goPrevPage();
    });
});