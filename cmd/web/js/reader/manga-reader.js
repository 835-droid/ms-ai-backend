// manga-reader.js - الملف الرئيسي للقارئ (مدخل التطبيق)
// هذا الملف يقوم بتحميل جميع الوحدات وتهيئة القارئ

// تحميل البيانات من الخادم
async function loadReaderData() {
    if (!requireAuth()) return;

    const mangaId = getQueryParam('mangaId');
    const chapterId = getQueryParam('chapterId');

    if (!mangaId || !chapterId) {
        document.getElementById('reader-error').textContent = '❌ الرابط غير مكتمل';
        return;
    }

    readerState.mangaId = mangaId;
    readerState.chapterId = chapterId;
    
    showProgressBar(true);
    updateProgressBar(20);

    try {
        // تحميل بيانات المانجا
        const manga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
        setManga(manga);
        updateProgressBar(50);
        
        // تحميل قائمة الفصول
        const chapterResp = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        const chapters = (chapterResp.chapters || []).map(normalizeChapterPages);
        setChapters(chapters);
        updateProgressBar(70);

        // العثور على الفصل الحالي
        const currentIndex = chapters.findIndex(ch => (ch.id || ch._id) === chapterId);
        if (currentIndex === -1) {
            document.getElementById('reader-error').textContent = '❌ الفصل غير موجود';
            showProgressBar(false);
            return;
        }

        const currentChapter = normalizeChapterPages(chapters[currentIndex]);
        setChapter(currentChapter);
        
        // تحميل الإشارة المرجعية
        loadBookmark();

        // بناء قائمة الفصول المنسدلة
        const chapterSelect = document.getElementById('chapter-select');
        if (chapterSelect) {
            chapterSelect.innerHTML = chapters.map(ch => {
                const id = ch.id || ch._id;
                return `<option value="${escapeHtml(id)}">📖 الفصل ${escapeHtml(ch.number)} - ${escapeHtml(ch.title || '')}</option>`;
            }).join('');
            chapterSelect.value = chapterId;
        }

        updateProgressBar(90);
        
        // عرض القارئ
        await renderReaderPage();
        
        // تسجيل المشاهدة
        if (!readerState.viewTracked) {
            try {
                await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}/view`, { method: 'POST' });
                setViewTracked(true);
            } catch { /* ignore */ }
        }
        
        updateProgressBar(100);
        setTimeout(() => showProgressBar(false), 500);
        
    } catch (error) {
        document.getElementById('reader-error').textContent = `❌ ${error.message}`;
        showProgressBar(false);
    }
}

// إضافة أزرار التنقل السريع في وضع الويب توون
function addWebtoonScrollButtons() {
    if (document.getElementById('webtoon-scroll-top')) return;
    
    const scrollTopBtn = document.createElement('button');
    scrollTopBtn.id = 'webtoon-scroll-top';
    scrollTopBtn.innerHTML = '↑';
    scrollTopBtn.className = 'btn btn-secondary scroll-btn';
    scrollTopBtn.style.cssText = `
        position: fixed;
        bottom: 80px;
        right: 20px;
        z-index: 100;
        border-radius: 50%;
        width: 45px;
        height: 45px;
        padding: 0;
        font-size: 1.5rem;
        opacity: 0;
        transition: opacity 0.3s;
    `;
    
    const scrollBottomBtn = document.createElement('button');
    scrollBottomBtn.id = 'webtoon-scroll-bottom';
    scrollBottomBtn.innerHTML = '↓';
    scrollBottomBtn.className = 'btn btn-secondary scroll-btn';
    scrollBottomBtn.style.cssText = `
        position: fixed;
        bottom: 140px;
        right: 20px;
        z-index: 100;
        border-radius: 50%;
        width: 45px;
        height: 45px;
        padding: 0;
        font-size: 1.5rem;
        opacity: 0;
        transition: opacity 0.3s;
    `;
    
    document.body.appendChild(scrollTopBtn);
    document.body.appendChild(scrollBottomBtn);
    
    const mainEl = document.getElementById('reader-main');
    if (mainEl) {
        mainEl.addEventListener('scroll', () => {
            const opacity = mainEl.scrollTop > 200 ? 1 : 0;
            scrollTopBtn.style.opacity = opacity;
            scrollBottomBtn.style.opacity = opacity;
        });
    }
    
    scrollTopBtn.addEventListener('click', () => {
        document.getElementById('reader-main')?.scrollTo({ top: 0, behavior: 'smooth' });
    });
    scrollBottomBtn.addEventListener('click', () => {
        const mainEl = document.getElementById('reader-main');
        if (mainEl) {
            mainEl.scrollTo({ top: mainEl.scrollHeight, behavior: 'smooth' });
        }
    });
}

// مراقبة تغيير وضع العرض لإضافة أزرار التمرير
function watchViewModeChange() {
    // إذا كان الوضع ويب توون، أضف الأزرار
    if (readerState.viewMode === 'webtoon') {
        addWebtoonScrollButtons();
    }
}

// تهيئة القارئ
function initReader() {
    loadReaderSettings();
    bindReaderEvents();
    initKeyboardShortcuts();
    initThemeToggle();
    loadReaderData();
    watchViewModeChange();
}

// بدء التشغيل عند تحميل الصفحة
document.addEventListener('DOMContentLoaded', initReader);