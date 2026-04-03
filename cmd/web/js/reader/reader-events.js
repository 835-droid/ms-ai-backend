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
        window.location.href = webPagePath('library.html');
    }
}

// تبديل السمة (ليلي/نهاري)
function initThemeToggle() {
    // Removed: theme toggle no longer available
}

function applyReaderTheme(theme) {
    // Removed: theme toggle no longer available
}

// اختصارات لوحة المفاتيح المتقدمة
function initKeyboardShortcuts() {
    document.addEventListener('keydown', (e) => {
        // تجاهل إذا كان التركيز على حقل إدخال
        if (e.target.tagName === 'INPUT' || e.target.tagName === 'TEXTAREA' || e.target.tagName === 'SELECT') {
            return;
        }
        
        switch(e.key) {
            case 'ArrowUp':
                document.getElementById('reader-main')?.scrollBy({ top: -300, behavior: 'smooth' });
                break;
            case 'ArrowDown':
                document.getElementById('reader-main')?.scrollBy({ top: 300, behavior: 'smooth' });
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
            case 'Escape':
                if (document.fullscreenElement) document.exitFullscreen();
                break;
        }
    });
}

// تسجيل جميع الأحداث
function bindReaderEvents() {
    // أزرار التنقل بين الفصول
    document.getElementById('prev-chapter-btn')?.addEventListener('click', goToPrevChapter);
    document.getElementById('next-chapter-btn')?.addEventListener('click', goToNextChapter);
    
    // عناصر التحكم الأخرى
    document.getElementById('chapter-select')?.addEventListener('change', changeChapterFromSelect);
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