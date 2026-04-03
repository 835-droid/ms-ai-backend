// reader-navigation.js - التنقل بين الصفحات والفصول

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
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
        
        // Track view for the new chapter
        await trackChapterView(readerState.chapterId);
        
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
        readerState.pageIndex = 0; // Start from beginning in webtoon-only mode
        await renderReaderPage();
        updateChapterSelect();
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
        
        // Track view for the new chapter
        await trackChapterView(readerState.chapterId);
        
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
        updateChapterSelect();
        
        const url = new URL(window.location.href);
        url.searchParams.set('chapterId', readerState.chapterId);
        history.pushState({}, '', url);
        
        // Track view for the new chapter
        await trackChapterView(readerState.chapterId);
    }
}