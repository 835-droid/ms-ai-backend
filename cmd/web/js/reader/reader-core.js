// reader-core.js - الحالة الأساسية والوظائف الرئيسية للقارئ

const READER_VIEW_MODE_KEY = 'readerViewMode';
const READER_PREFETCH_KEY = 'prefetchEnabled';

// الحالة العامة للقارئ
let readerState = {
    mangaId: null,
    chapterId: null,
    manga: null,
    chapters: [],
    chapter: null,
    pageIndex: 0,
    viewTracked: false,
    viewMode: localStorage.getItem(READER_VIEW_MODE_KEY) === 'webtoon' ? 'webtoon' : 'paged',
    autoNextChapter: true,
    isFullscreen: false,
    prefetchEnabled: localStorage.getItem(READER_PREFETCH_KEY) !== 'false',
    isLoading: false,
    totalPages: 0
};

// دوال للوصول إلى الحالة (getters)
function getReaderState() {
    return readerState;
}

function getCurrentMangaId() {
    return readerState.mangaId;
}

function getCurrentChapterId() {
    return readerState.chapterId;
}

function getCurrentPageIndex() {
    return readerState.pageIndex;
}

function getCurrentChapter() {
    return readerState.chapter;
}

function getCurrentManga() {
    return readerState.manga;
}

function getChaptersList() {
    return readerState.chapters;
}

function getViewMode() {
    return readerState.viewMode;
}

function isAutoNextChapterEnabled() {
    return readerState.autoNextChapter;
}

function isPrefetchEnabled() {
    return readerState.prefetchEnabled;
}

// دوال لتعديل الحالة (setters)
function setReaderState(newState) {
    readerState = { ...readerState, ...newState };
}

function setPageIndex(index) {
    readerState.pageIndex = Math.max(0, index);
}

function setChapter(chapter) {
    readerState.chapter = chapter;
    readerState.totalPages = chapter?.pages?.length || 0;
}

function setChapters(chapters) {
    readerState.chapters = chapters;
}

function setManga(manga) {
    readerState.manga = manga;
}

function setViewMode(mode) {
    readerState.viewMode = mode;
    localStorage.setItem(READER_VIEW_MODE_KEY, mode);
}

function setViewTracked(tracked) {
    readerState.viewTracked = tracked;
}

function setLoading(loading) {
    readerState.isLoading = loading;
}

// دوال مساعدة للتحقق
function isWebtoonMode() {
    return readerState.viewMode === 'webtoon';
}

function hasPages() {
    return readerState.chapter?.pages?.length > 0;
}

function getTotalPages() {
    return readerState.chapter?.pages?.length || 0;
}

function isValidPageIndex(index) {
    const total = getTotalPages();
    return index >= 0 && index < total;
}