// reader-ui.js - تحديث واجهة المستخدم والعناصر

// إظهار/إخفاء شريط التقدم
function showProgressBar(show) {
    const container = document.getElementById('progress-container');
    if (container) {
        container.style.display = show ? 'block' : 'none';
    }
}

// تحديث شريط التقدم
function updateProgressBar(percent) {
    const bar = document.getElementById('progress-bar-reader');
    if (bar) {
        bar.style.width = percent + '%';
    }
}

// تحديث قائمة الفصول المنسدلة
function updateChapterSelect() {
    const select = document.getElementById('chapter-select');
    if (!select) return;
    
    for (let i = 0; i < select.options.length; i++) {
        if (select.options[i].value === readerState.chapterId) {
            select.selectedIndex = i;
            break;
        }
    }
}

// تحديث معلومات المانجا والفصل في الرأس
function updateReaderMeta() {
    const mangaTitle = document.getElementById('reader-manga-title');
    const mangaLink = document.getElementById('reader-manga-title-link');
    const chapterTitle = document.getElementById('reader-chapter-title');
    
    if (mangaTitle && readerState.manga) {
        mangaTitle.textContent = readerState.manga.title || 'القراءة';
        if (mangaLink) {
            mangaLink.href = `${webPagePath('manga-details.html')}?id=${encodeURIComponent(readerState.mangaId)}`;
        }
    }
    if (chapterTitle) {
        chapterTitle.textContent = readerState.chapter
            ? `📖 الفصل ${readerState.chapter.number} - ${readerState.chapter.title || ''}`
            : '';
    }
}

// عرض وضع الويب توون (تمرير طويل)
function renderWebtoonReader() {
    const pagedContainer = document.getElementById('paged-mode-container');
    const webtoonPages = document.getElementById('reader-webtoon-pages');
    const pages = readerState.chapter?.pages || [];

    if (!webtoonPages) return;

    if (pagedContainer) pagedContainer.style.display = 'none';
    webtoonPages.hidden = false;

    if (!pages.length) {
        webtoonPages.innerHTML = '';
        document.getElementById('reader-error').textContent = '⚠️ لا توجد صفحات داخل هذا الفصل';
        return;
    }

    showProgressBar(true);
    updateProgressBar(100);
    
    webtoonPages.innerHTML = pages.map((page, index) => {
        const safeOriginalUrl = normalizeImageUrl(page);
        const resolvedUrl = resolveReaderImageUrl(safeOriginalUrl);
        return `
            <div class="reader-webtoon-frame" data-page-index="${index}">
                <div class="reader-webtoon-label">📄 صفحة ${index + 1}</div>
                <img
                    class="reader-webtoon-page"
                    src="${escapeHtml(resolvedUrl)}"
                    alt="صفحة ${index + 1}"
                    data-original-url="${escapeHtml(safeOriginalUrl)}"
                    loading="lazy"
                >
            </div>
        `;
    }).join('');

    const webtoonImages = webtoonPages.querySelectorAll('.reader-webtoon-page');
    webtoonImages.forEach((img, idx) => {
        const originalUrl = img.dataset.originalUrl;
        retryImageLoad(img, originalUrl);
    });

    document.getElementById('reader-error').textContent = '';
    showProgressBar(false);
}

// الدالة الرئيسية لعرض القارئ
async function renderReaderPage() {
    updateReaderMeta();

    renderWebtoonReader();
    saveBookmark();
}