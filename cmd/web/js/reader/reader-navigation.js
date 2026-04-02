// reader-navigation.js - التنقل بين الصفحات والفصول

// الانتقال إلى صفحة محددة
function jumpToPage() {
    const selector = document.getElementById('page-selector');
    if (!selector || isWebtoonMode()) return;
    
    const newIndex = parseInt(selector.value, 10);
    if (!isNaN(newIndex) && newIndex !== readerState.pageIndex && isValidPageIndex(newIndex)) {
        readerState.pageIndex = newIndex;
        renderReaderPage();
        saveBookmark();
    }
}

// الانتقال إلى أول صفحة
function goFirstPage() {
    if (isWebtoonMode()) return;
    if (readerState.pageIndex !== 0 && getTotalPages() > 0) {
        readerState.pageIndex = 0;
        renderReaderPage();
        saveBookmark();
    }
}

// الانتقال إلى آخر صفحة
function goLastPage() {
    if (isWebtoonMode()) return;
    const lastIndex = getTotalPages() - 1;
    if (readerState.pageIndex !== lastIndex && lastIndex >= 0) {
        readerState.pageIndex = lastIndex;
        renderReaderPage();
        saveBookmark();
    }
}

// الانتقال إلى الصفحة السابقة
function goPrevPage() {
    if (isWebtoonMode()) return;
    if (readerState.pageIndex > 0) {
        readerState.pageIndex--;
        renderReaderPage();
        saveBookmark();
    }
}

// الانتقال إلى الصفحة التالية
function goNextPage() {
    if (isWebtoonMode()) return;
    const pages = readerState.chapter?.pages || [];
    if (readerState.pageIndex < pages.length - 1) {
        readerState.pageIndex++;
        renderReaderPage();
        saveBookmark();
        
        // تحميل مسبق للصفحة التالية
        if (readerState.prefetchEnabled && readerState.pageIndex + 1 < pages.length) {
            prefetchImage(pages[readerState.pageIndex + 1]);
        }
        
        // الانتقال التلقائي للفصل التالي
        if (readerState.autoNextChapter && readerState.pageIndex === pages.length - 1 && pages.length > 0) {
            showToast('⏩ الانتقال إلى الفصل التالي...', 'info', 1500);
            setTimeout(() => goToNextChapter(), 1000);
        }
    }
}

// الانتقال إلى الفصل التالي
async function goToNextChapter() {
    const currentIndex = readerState.chapters.findIndex(ch => (ch.id || ch._id) === readerState.chapterId);
    if (currentIndex >= 0 && currentIndex < readerState.chapters.length - 1) {
        const nextChapter = readerState.chapters[currentIndex + 1];
        readerState.chapterId = nextChapter.id || nextChapter._id;
        readerState.chapter = nextChapter;
        readerState.pageIndex = 0;
        await renderReaderPage();
        updateChapterSelect();
        updatePageSelector();
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
        
        showToast(`📚 الفصل التالي: ${nextChapter.title || nextChapter.number}`, 'success', 2000);
    } else {
        showToast('✨ أنت في آخر فصل', 'info', 1500);
    }
}

// الانتقال إلى الفصل السابق
async function goToPrevChapter() {
    const currentIndex = readerState.chapters.findIndex(ch => (ch.id || ch._id) === readerState.chapterId);
    if (currentIndex > 0) {
        const prevChapter = readerState.chapters[currentIndex - 1];
        readerState.chapterId = prevChapter.id || prevChapter._id;
        readerState.chapter = prevChapter;
        readerState.pageIndex = (prevChapter.pages?.length || 1) - 1;
        await renderReaderPage();
        updateChapterSelect();
        updatePageSelector();
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
        
        showToast(`📚 الفصل السابق: ${prevChapter.title || prevChapter.number}`, 'success', 2000);
    } else {
        showToast('✨ أنت في أول فصل', 'info', 1500);
    }
}

// تغيير الفصل من القائمة المنسدلة
async function changeChapterFromSelect() {
    const selectedId = document.getElementById('chapter-select')?.value;
    const index = readerState.chapters.findIndex(ch => (ch.id || ch._id) === selectedId);
    if (index >= 0) {
        readerState.chapterId = selectedId;
        readerState.chapter = readerState.chapters[index];
        readerState.pageIndex = 0;
        await renderReaderPage();
        updatePageSelector();
        updateChapterSelect();
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
    }
}

// تبديل وضع العرض (صفحات / ويب توون)
function toggleReaderViewMode() {
    const newMode = readerState.viewMode === 'webtoon' ? 'paged' : 'webtoon';
    setViewMode(newMode);
    renderReaderPage();
}