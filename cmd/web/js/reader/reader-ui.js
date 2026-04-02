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

// تحديث قائمة الصفحات المنسدلة
function updatePageSelector() {
    const selector = document.getElementById('page-selector');
    const totalBadge = document.getElementById('total-pages-badge');
    const pages = readerState.chapter?.pages || [];
    const total = pages.length;
    
    if (!selector) return;
    
    if (selector.options.length !== total) {
        selector.innerHTML = '';
        for (let i = 0; i < total; i++) {
            const option = document.createElement('option');
            option.value = i;
            option.textContent = `${i + 1}`;
            selector.appendChild(option);
        }
    }
    
    if (readerState.pageIndex >= 0 && readerState.pageIndex < total) {
        selector.value = readerState.pageIndex;
    }
    
    if (totalBadge) {
        totalBadge.textContent = `/ ${total}`;
    }
    
    readerState.totalPages = total;
}

// تحديث حالة أزرار التنقل
function updatePageButtons() {
    const prevPageBtn = document.getElementById('prev-page-btn');
    const nextPageBtn = document.getElementById('next-page-btn');
    const firstPageBtn = document.getElementById('first-page-btn');
    const lastPageBtn = document.getElementById('last-page-btn');
    const pageSelector = document.getElementById('page-selector');
    
    if (!prevPageBtn || !nextPageBtn) return;
    
    const pages = readerState.chapter?.pages || [];
    const isWebtoon = readerState.viewMode === 'webtoon';
    
    if (isWebtoon) {
        prevPageBtn.disabled = true;
        nextPageBtn.disabled = true;
        firstPageBtn.disabled = true;
        lastPageBtn.disabled = true;
        if (pageSelector) pageSelector.disabled = true;
    } else {
        prevPageBtn.disabled = readerState.pageIndex <= 0;
        nextPageBtn.disabled = readerState.pageIndex >= pages.length - 1;
        firstPageBtn.disabled = readerState.pageIndex <= 0;
        lastPageBtn.disabled = readerState.pageIndex >= pages.length - 1;
        if (pageSelector) pageSelector.disabled = false;
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
    const chapterTitle = document.getElementById('reader-chapter-title');
    
    if (mangaTitle) {
        mangaTitle.textContent = readerState.manga?.title || 'القراءة';
    }
    if (chapterTitle) {
        chapterTitle.textContent = readerState.chapter
            ? `📖 الفصل ${readerState.chapter.number} - ${readerState.chapter.title || ''}`
            : '';
    }
}

// تحديث زر تبديل وضع العرض
function updateViewModeButton() {
    const button = document.getElementById('toggle-view-mode-btn');
    if (!button) return;
    button.innerHTML = readerState.viewMode === 'webtoon' ? '📄 وضع الصفحات' : '📱 وضع الويب توون';
}

// عرض وضع الصفحات المنفردة
function renderPagedReader() {
    const pageImage = document.getElementById('reader-page-image');
    const pagedContainer = document.getElementById('paged-mode-container');
    const webtoonPages = document.getElementById('reader-webtoon-pages');
    const pages = readerState.chapter?.pages || [];

    if (!pageImage || !webtoonPages) return;

    if (pagedContainer) pagedContainer.style.display = 'flex';
    webtoonPages.hidden = true;
    webtoonPages.innerHTML = '';
    pageImage.hidden = false;

    if (!pages.length) {
        pageImage.src = '';
        pageImage.alt = 'لا توجد صفحات';
        document.getElementById('reader-error').textContent = '⚠️ لا توجد صفحات داخل هذا الفصل';
        updatePageButtons();
        updatePageSelector();
        return;
    }

    const safeIndex = Math.max(0, Math.min(readerState.pageIndex, pages.length - 1));
    if (safeIndex !== readerState.pageIndex) {
        readerState.pageIndex = safeIndex;
    }

    const originalPageUrl = pages[safeIndex];
    pageImage.dataset.originalUrl = originalPageUrl;
    
    pageImage.classList.add('loading');
    showProgressBar(true);
    updateProgressBar(((safeIndex + 1) / pages.length) * 100);
    
    const resolvedUrl = resolveReaderImageUrl(originalPageUrl);
    pageImage.src = resolvedUrl;
    retryImageLoad(pageImage, originalPageUrl, () => {
        pageImage.classList.remove('loading');
        showProgressBar(false);
    });

    pageImage.alt = `صفحة ${safeIndex + 1}`;
    document.getElementById('reader-error').textContent = '';
    updatePageButtons();
    updatePageSelector();
    
    // تحميل مسبق للصفحات التالية
    prefetchNextPages(safeIndex, pages, 2);
}

// عرض وضع الويب توون (تمرير طويل)
function renderWebtoonReader() {
    const pageImage = document.getElementById('reader-page-image');
    const pagedContainer = document.getElementById('paged-mode-container');
    const webtoonPages = document.getElementById('reader-webtoon-pages');
    const pages = readerState.chapter?.pages || [];

    if (!pageImage || !webtoonPages) return;

    if (pagedContainer) pagedContainer.style.display = 'none';
    pageImage.hidden = true;
    pageImage.src = '';
    webtoonPages.hidden = false;

    if (!pages.length) {
        webtoonPages.innerHTML = '';
        document.getElementById('reader-error').textContent = '⚠️ لا توجد صفحات داخل هذا الفصل';
        updatePageButtons();
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
    updatePageButtons();
    showProgressBar(false);
}

// الدالة الرئيسية لعرض القارئ
async function renderReaderPage() {
    updateReaderMeta();
    updateViewModeButton();

    if (readerState.viewMode === 'webtoon') {
        renderWebtoonReader();
    } else {
        renderPagedReader();
    }
    saveBookmark();
}