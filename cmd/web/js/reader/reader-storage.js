// reader-storage.js - إدارة التخزين المحلي (bookmarks, settings)

const READER_BOOKMARKS_KEY = 'readerBookmarks';

// حفظ إشارة مرجعية لآخر صفحة مقروءة
function saveBookmark() {
    const bookmark = {
        mangaId: readerState.mangaId,
        chapterId: readerState.chapterId,
        pageIndex: readerState.pageIndex,
        viewMode: readerState.viewMode,
        timestamp: Date.now()
    };
    
    let bookmarks = JSON.parse(localStorage.getItem(READER_BOOKMARKS_KEY) || '{}');
    bookmarks[`${readerState.mangaId}_${readerState.chapterId}`] = bookmark;
    localStorage.setItem(READER_BOOKMARKS_KEY, JSON.stringify(bookmarks));
    
    // تحديث شريط التقدم
    const total = readerState.chapter?.pages?.length || 1;
    const percent = ((readerState.pageIndex + 1) / total) * 100;
    updateProgressBar(percent);
}

// تحميل إشارة مرجعية للفصل الحالي
function loadBookmark() {
    const bookmarks = JSON.parse(localStorage.getItem(READER_BOOKMARKS_KEY) || '{}');
    const key = `${readerState.mangaId}_${readerState.chapterId}`;
    const bookmark = bookmarks[key];
    
    if (bookmark && bookmark.pageIndex > 0) {
        readerState.pageIndex = bookmark.pageIndex;
        showToast(`📖 استكمال من الصفحة ${bookmark.pageIndex + 1}`, 'info', 2000);
        return true;
    }
    return false;
}

// حذف إشارة مرجعية لفصل معين
function deleteBookmark(mangaId, chapterId) {
    let bookmarks = JSON.parse(localStorage.getItem(READER_BOOKMARKS_KEY) || '{}');
    delete bookmarks[`${mangaId}_${chapterId}`];
    localStorage.setItem(READER_BOOKMARKS_KEY, JSON.stringify(bookmarks));
}

// مسح جميع الإشارات المرجعية لمستخدم معين
function clearAllBookmarks() {
    localStorage.removeItem(READER_BOOKMARKS_KEY);
}

// حفظ إعدادات القارئ
function saveReaderSettings() {
    const settings = {
        viewMode: readerState.viewMode,
        prefetchEnabled: readerState.prefetchEnabled,
        autoNextChapter: readerState.autoNextChapter
    };
    localStorage.setItem('readerSettings', JSON.stringify(settings));
}

// تحميل إعدادات القارئ
function loadReaderSettings() {
    const settings = JSON.parse(localStorage.getItem('readerSettings') || '{}');
    if (typeof settings.prefetchEnabled === 'boolean') readerState.prefetchEnabled = settings.prefetchEnabled;
    if (typeof settings.autoNextChapter === 'boolean') readerState.autoNextChapter = settings.autoNextChapter;
}

// الحصول على آخر فصل مقروء للمانجا الحالية
function getLastReadChapter() {
    const bookmarks = JSON.parse(localStorage.getItem(READER_BOOKMARKS_KEY) || '{}');
    let lastChapterId = null;
    let lastChapterNumber = -1;
    
    for (const key in bookmarks) {
        if (key.startsWith(readerState.mangaId + '_')) {
            const chapterId = key.split('_')[1];
            const chapter = readerState.chapters.find(ch => (ch.id || ch._id) === chapterId);
            if (chapter && chapter.number > lastChapterNumber) {
                lastChapterNumber = chapter.number;
                lastChapterId = chapterId;
            }
        }
    }
    return lastChapterId;
}