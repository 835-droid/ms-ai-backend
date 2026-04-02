// reader-events.js - الأحداث واختصارات لوحة المفاتيح

// تبديل وضع ملء الشاشة
function toggleFullscreen() {
    const container = document.querySelector('.reader-container');
    if (!document.fullscreenElement) {
        container.requestFullscreen().catch(err => {
            showToast(`❌ ${err.message}`, 'error');
        });
    } else {
        document.exitFullscreen();
    }
}

// العودة إلى الصفحة السابقة أو التفاصيل
function handleReaderBack() {
    if (window.history.length > 1) {
        window.history.back();
        return;
    }
    const mangaId = readerState.mangaId || getQueryParam('mangaId');
    if (mangaId) {
        window.location.href = `${webPagePath('manga-details.html')}?id=${encodeURIComponent(mangaId)}`;
    } else {
        window.location.href = webPagePath('dashboard.html');
    }
}

// تبديل السمة (ليلي/نهاري)
function initThemeToggle() {
    const themeBtn = document.getElementById('theme-toggle-btn');
    if (!themeBtn) return;
    
    const savedTheme = localStorage.getItem('readerTheme') || 'dark';
    applyReaderTheme(savedTheme);
    
    themeBtn.addEventListener('click', () => {
        const container = document.querySelector('.reader-container');
        const currentTheme = container.classList.contains('bg-white') ? 'dark' : 'light';
        const newTheme = currentTheme === 'dark' ? 'light' : 'dark';
        applyReaderTheme(newTheme);
        localStorage.setItem('readerTheme', newTheme);
    });
}

function applyReaderTheme(theme) {
    const container = document.querySelector('.reader-container');
    if (!container) return;
    container.classList.remove('bg-white', 'bg-gray');
    if (theme === 'light') {
        container.classList.add('bg-white');
    } else if (theme === 'gray') {
        container.classList.add('bg-gray');
    }
}

// اختصارات لوحة المفاتيح المتقدمة
function initKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // تجاهل إذا كان التركيز على حقل إدخال
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            return;
        }
        
        switch(e.key) {
            case 'ArrowLeft':
                if (!isWebtoonMode()) goPrevPage();
                break;
            case 'ArrowRight':
                if (!isWebtoonMode()) goNextPage();
                break;
            case 'ArrowUp':
                if (isWebtoonMode()) {
                    document.getElementById('reader-main')?.scrollBy({ top: -300, behavior: 'smooth' });
                }
                break;
            case 'ArrowDown':
                if (isWebtoonMode()) {
                    document.getElementById('reader-main')?.scrollBy({ top: 300, behavior: 'smooth' });
                }
                break;
            case 'Home':
                e.preventDefault();
                goFirstPage();
                break;
            case 'End':
                e.preventDefault();
                goLastPage();
                break;
            case 'f':
            case 'F':
                toggleFullscreen();
                break;
            case 'n':
            case 'N':
                goToNextChapter();
                break;
            case 'p':
            case 'P':
                goToPrevChapter();
                break;
            case 't':
            case 'T':
                toggleReaderViewMode();
                break;
            case 'Escape':
                if (document.fullscreenElement) document.exitFullscreen();
                break;
        }
    });
}

// تسجيل جميع الأحداث
function bindReaderEvents() {
    // أزرار التنقل بين الصفحات
    document.getElementById('prev-page-btn')?.addEventListener('click', goPrevPage);
    document.getElementById('next-page-btn')?.addEventListener('click', goNextPage);
    document.getElementById('first-page-btn')?.addEventListener('click', goFirstPage);
    document.getElementById('last-page-btn')?.addEventListener('click', goLastPage);
    
    // أزرار التنقل بين الفصول
    document.getElementById('prev-chapter-btn')?.addEventListener('click', goToPrevChapter);
    document.getElementById('next-chapter-btn')?.addEventListener('click', goToNextChapter);
    
    // عناصر التحكم الأخرى
    document.getElementById('chapter-select')?.addEventListener('change', changeChapterFromSelect);
    document.getElementById('page-selector')?.addEventListener('change', jumpToPage);
    document.getElementById('back-button')?.addEventListener('click', handleReaderBack);
    document.getElementById('toggle-view-mode-btn')?.addEventListener('click', toggleReaderViewMode);
    document.getElementById('fullscreen-btn')?.addEventListener('click', toggleFullscreen);
    
    // معالجة أخطاء الصور
    const pagedImage = document.getElementById('reader-page-image');
    if (pagedImage) {
        pagedImage.addEventListener('error', (e) => {
            const img = e.currentTarget;
            const originalUrl = img?.dataset?.originalUrl || '';
            if (!originalUrl) return;
            const proxyUrl = `${CONFIG.API_BASE}/assets/image-proxy?url=${encodeURIComponent(originalUrl)}`;
            if (img.src !== proxyUrl) {
                img.src = proxyUrl;
            }
        });
    }
}