// manga-details-enhanced.js - نسخة متكاملة مع إعجاب ومفضلة وتعليقات
let currentManga = null;
let currentChapters = [];
let likeInFlight = false;
let ratingInFlight = false;
let favoriteInFlight = false;
let reactionInFlight = false;
let currentRating = 0;
let galleryImages = [];
let currentGalleryIndex = 0;
let comments = [];
let currentUserReaction = null;
let currentReactionType = null;

// ========== دوال مساعدة ==========
function formatCompactNumber(value) {
    const number = Number(value || 0);
    if (number >= 1000000) return `${(number / 1000000).toFixed(1)}M`;
    if (number >= 1000) return `${(number / 1000).toFixed(1)}K`;
    return String(number);
}

function formatRating(value) {
    return Number(value || 0).toFixed(1);
}

const userChapterRatingCache = {};
let lastRatingRequestTimestamp = 0;
const myRatingRequestIntervalMs = 150;

async function fetchUserChapterRating(chapterId) {
    if (!chapterId) return 0;
    if (userChapterRatingCache[chapterId] !== undefined) {
        return userChapterRatingCache[chapterId];
    }

    const wait = myRatingRequestIntervalMs - (Date.now() - lastRatingRequestTimestamp);
    if (wait > 0) {
        await new Promise(resolve => setTimeout(resolve, wait));
    }

    lastRatingRequestTimestamp = Date.now();

    try {
        const mangaId = getCurrentMangaId();
        if (!mangaId) return 0;
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}/my-rating`);
        const rating = data.has_rated ? Number(data.score || 0) : 0;
        userChapterRatingCache[chapterId] = rating;
        return rating;
    } catch (error) {
        console.debug('Failed to load user chapter rating:', error);
        return 0;
    }
}

function getCurrentMangaId() {
    return currentManga?.id || currentManga?._id || getQueryParam('id');
}

// Debounce utility function
function debounce(func, wait) {
    let timeout;
    return function executedFunction(...args) {
        const later = () => {
            clearTimeout(timeout);
            func(...args);
        };
        clearTimeout(timeout);
        timeout = setTimeout(later, wait);
    };
}

// ========== إدارة المفضلة (LocalStorage + API إذا وُجد) ==========
async function loadFavorites() {
    try {
        const data = await apiFetch('/mangas/favorites');
        return data.items || [];
    } catch (error) {
        console.error('Failed to load favorites from API:', error);
        // Fallback to localStorage if API fails
        return JSON.parse(localStorage.getItem('favoriteMangas') || '[]');
    }
}

async function isFavorite(mangaId) {
    try {
        const response = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/favorite`);
        return response.is_favorite || false;
    } catch (error) {
        console.error('Failed to check favorite status:', error);
        // Fallback to localStorage if API fails
        const favs = JSON.parse(localStorage.getItem('favoriteMangas') || '[]');
        return favs.includes(mangaId);
    }
}

async function toggleFavorite() {
    if (favoriteInFlight) return;
    const mangaId = getCurrentMangaId();
    if (!mangaId) return;

    favoriteInFlight = true;

    try {
        const isFav = await isFavorite(mangaId);

        if (isFav) {
            await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/favorite`, { method: 'DELETE' });
            currentManga.favorites_count = Math.max((currentManga.favorites_count || 0) - 1, 0);
            showToast('تمت إزالة من المفضلة', 'info');
        } else {
            await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/favorite`, { method: 'POST' });
            currentManga.favorites_count = (currentManga.favorites_count || 0) + 1;
            showToast('أضيفت إلى المفضلة', 'success');
        }

        // Update favorites count in UI
        const favCountEl = document.getElementById('favorites-count');
        if (favCountEl) {
            favCountEl.textContent = formatCompactNumber(currentManga.favorites_count);
        }

        updateFavoriteButton(!isFav);
    } catch (error) {
        console.error('Failed to toggle favorite:', error);
        showToast('فشل في تحديث المفضلة', 'error');
    } finally {
        favoriteInFlight = false;
    }
}

function updateFavoriteButton(isFav) {
    const btn = document.getElementById('favorite-btn');
    if (!btn) return;
    if (isFav) {
        btn.innerHTML = '<i class="fas fa-heart"></i> في المفضلة';
        btn.classList.add('btn-primary');
        btn.classList.remove('btn-secondary');
    } else {
        btn.innerHTML = '<i class="far fa-heart"></i> أضف للمفضلة';
        btn.classList.remove('btn-primary');
        btn.classList.add('btn-secondary');
    }
}

// ========== زر الإعجاب المحسن مع تأثير ==========
function animateLikeButton() {
    const btn = document.getElementById('like-button-enhanced');
    if (!btn) return;
    btn.classList.add('like-animation');
    setTimeout(() => btn.classList.remove('like-animation'), 300);
    
    // تأثير حبيبات (confetti بسيط)
    const rect = btn.getBoundingClientRect();
    for (let i = 0; i < 10; i++) {
        const particle = document.createElement('div');
        particle.className = 'like-particle';
        particle.style.left = rect.left + rect.width/2 + 'px';
        particle.style.top = rect.top + rect.height/2 + 'px';
        particle.style.setProperty('--dx', (Math.random() - 0.5) * 100 + 'px');
        particle.style.setProperty('--dy', (Math.random() - 0.5) * 100 - 50 + 'px');
        document.body.appendChild(particle);
        setTimeout(() => particle.remove(), 500);
    }
}

async function handleLikeToggleEnhanced() {
    const mangaId = getCurrentMangaId();
    if (!mangaId || likeInFlight) return;
    likeInFlight = true;
    animateLikeButton();
    try {
        // Send reaction with type 'upvote' (default reaction type)
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/react`, { 
            method: 'POST',
            body: JSON.stringify({ type: 'upvote' })
        });
        currentManga = data.manga || { ...currentManga, ...data };
        currentReactionType = data.removed ? null : (data.reaction_type || 'upvote');
        updateStatsUi();
        const message = data.removed ? 'تم إلغاء ردة الفعل' : 'تم تسجيل ردة فعلك! ♥';
        showToast(message, data.removed ? 'info' : 'success');
        // تحديث عداد الإعجابات في الزر
        const likeCountSpan = document.getElementById('like-count');
        if (likeCountSpan) likeCountSpan.textContent = formatCompactNumber(currentManga.likes_count);
    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        likeInFlight = false;
    }
}

// ========== التقييم بالنجوم (مثل السابق) ==========
function renderStars(ratingValue, interactive = false) {
    // Keep read-only rating display; chapter-level ratings are 1-10.
    const container = document.getElementById('stars-container');
    if (!container) return;
    const maxStars = 10;
    const fullStars = Math.floor(Math.min(ratingValue, maxStars));
    const hasHalf = ratingValue % 1 >= 0.5;
    container.innerHTML = '';
    for (let i = 1; i <= maxStars; i++) {
        const star = document.createElement('i');
        star.className = `fas fa-star star ${i <= fullStars ? 'active' : ''}`;
        if (i === fullStars + 1 && hasHalf && !(i <= fullStars)) {
            star.className = 'fas fa-star-half-alt star active';
        }
        star.dataset.value = i;
        // interactivity disabled for manga-level rating (deprecated)
        container.appendChild(star);
    }
    const avgSpan = document.getElementById('rating-average');
    if (avgSpan) avgSpan.textContent = `(${formatRating(currentManga?.average_rating)} من 10)`;
}

function highlightStars(value) {
    document.querySelectorAll('#stars-container .star').forEach((star, idx) => {
        if (idx + 1 <= value) star.classList.add('hover');
        else star.classList.remove('hover');
    });
}
function resetStars() {
    document.querySelectorAll('#stars-container .star').forEach(star => star.classList.remove('hover'));
}
// submitRating function removed - rating moved to chapter level

// ========== تقييم الفصول ==========
async function renderChapterStars(chapterId) {
    const container = document.getElementById(`chapter-stars-${chapterId}`);
    if (!container) return;

    const mangaId = getCurrentMangaId();
    if (!mangaId) return;

    try {
        // Fetch user's existing rating for this chapter with caching and throttling
        const userRating = await fetchUserChapterRating(chapterId);

        const maxStars = 10;
        container.innerHTML = '';
        for (let i = 1; i <= maxStars; i++) {
            const star = document.createElement('i');
            star.className = `fas fa-star chapter-star ${i <= userRating ? 'active' : ''}`;
            star.dataset.value = i;
            star.dataset.chapterId = chapterId;

            // Always allow rating (including re-rating)
            star.addEventListener('click', handleChapterRating);
            star.addEventListener('mouseenter', (e) => highlightChapterStars(e.target.dataset.chapterId, e.target.dataset.value));
            star.addEventListener('mouseleave', (e) => resetChapterStars(e.target.dataset.chapterId));

            container.appendChild(star);
        }
    } catch (error) {
        console.error('Failed to load user rating:', error);
        // Fallback to empty stars if API fails
        renderEmptyChapterStars(chapterId);
    }
}

function renderEmptyChapterStars(chapterId) {
    const container = document.getElementById(`chapter-stars-${chapterId}`);
    if (!container) return;

    const maxStars = 10;
    container.innerHTML = '';
    for (let i = 1; i <= maxStars; i++) {
        const star = document.createElement('i');
        star.className = `fas fa-star chapter-star`;
        star.dataset.value = i;
        star.dataset.chapterId = chapterId;
        star.addEventListener('click', handleChapterRating);
        star.addEventListener('mouseenter', (e) => highlightChapterStars(e.target.dataset.chapterId, e.target.dataset.value));
        star.addEventListener('mouseleave', (e) => resetChapterStars(e.target.dataset.chapterId));
        container.appendChild(star);
    }
}

function highlightChapterStars(chapterId, value) {
    document.querySelectorAll(`#chapter-stars-${chapterId} .chapter-star`).forEach((star, idx) => {
        if (idx + 1 <= value) star.classList.add('hover');
        else star.classList.remove('hover');
    });
}

function resetChapterStars(chapterId) {
    document.querySelectorAll(`#chapter-stars-${chapterId} .chapter-star`).forEach(star => star.classList.remove('hover'));
}

async function handleChapterRating(e) {
    const chapterId = e.target.dataset.chapterId;
    const score = parseInt(e.target.dataset.value);
    const mangaId = getCurrentMangaId();

    if (!mangaId || !chapterId) return;

    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}/rate`, {
            method: 'POST',
            body: JSON.stringify({ score })
        });

        showToast(`قيمت الفصل بـ ${score} نجوم`, 'success');

        // Cache the new rating for this chapter and re-render
        userChapterRatingCache[chapterId] = data.user_score || score;

        // Re-render the stars to show the updated rating
        await renderChapterStars(chapterId);

        // Update the chapter data and re-render
        const chapterIndex = currentChapters.findIndex(ch => (ch.id || ch._id) === chapterId);
        if (chapterIndex !== -1) {
            // Refresh chapter data from server
            const chapterResp = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
            currentChapters = chapterResp.chapters || [];
            renderChapters();
            // Refresh manga data to update stats
            currentManga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
            updateStatsUi();
        }
    } catch (error) {
        showToast(error.message || 'فشل في تقييم الفصل', 'error');
    }
}

// ========== معرض الصور ==========
function initGallery(images) {
    galleryImages = images || [];
    const container = document.getElementById('manga-gallery');
    if (!container || galleryImages.length <= 1) return;
    
    container.innerHTML = `
        <div class="gallery-container">
            <img id="gallery-inline-img" src="${escapeHtml(galleryImages[0])}" alt="Gallery Image">
            <div class="gallery-controls">
                <button id="gallery-inline-prev" class="btn btn-sm btn-secondary"><i class="fas fa-chevron-left"></i></button>
                <span class="gallery-indicator"><span id="gallery-inline-current">1</span> / ${galleryImages.length}</span>
                <button id="gallery-inline-next" class="btn btn-sm btn-secondary"><i class="fas fa-chevron-right"></i></button>
            </div>
        </div>
    `;
    document.getElementById('gallery-inline-prev')?.addEventListener('click', prevGallery);
    document.getElementById('gallery-inline-next')?.addEventListener('click', nextGallery);
    currentGalleryIndex = 0;
}

function openGallery(index) {
    if (index < 0 || index >= galleryImages.length) return;
    currentGalleryIndex = index;
    const img = document.getElementById('gallery-inline-img');
    if (img) {
        img.src = galleryImages[index];
        document.getElementById('gallery-inline-current').textContent = index + 1;
    }
}

function nextGallery() {
    currentGalleryIndex = (currentGalleryIndex + 1) % galleryImages.length;
    openGallery(currentGalleryIndex);
}

function prevGallery() {
    currentGalleryIndex = (currentGalleryIndex - 1 + galleryImages.length) % galleryImages.length;
    openGallery(currentGalleryIndex);
}

// ========== مشاركة المانجا ==========
function shareManga(platform) {
    const title = currentManga?.title || 'مانجا';
    const url = window.location.href;
    const text = `اقرأ ${title} على منصة MS-AI`;
    const shareUrls = {
        twitter: `https://twitter.com/intent/tweet?text=${encodeURIComponent(text)}&url=${encodeURIComponent(url)}`,
        facebook: `https://www.facebook.com/sharer/sharer.php?u=${encodeURIComponent(url)}`,
        whatsapp: `https://wa.me/?text=${encodeURIComponent(text + ' ' + url)}`
    };
    if (platform === 'copy') {
        navigator.clipboard.writeText(url);
        showToast('تم نسخ الرابط', 'success');
        return;
    }
    if (shareUrls[platform]) window.open(shareUrls[platform], '_blank', 'width=600,height=400');
}

// ========== آخر فصل مقروء ومتابعة القراءة ==========
function getLastReadChapter() {
    const bookmarks = JSON.parse(localStorage.getItem('readerBookmarks') || '{}');
    const mangaId = getCurrentMangaId();
    let lastChapterId = null, lastChapterNumber = 0;
    for (const key in bookmarks) {
        if (key.startsWith(mangaId + '_')) {
            const chapterId = key.split('_')[1];
            const chapter = currentChapters.find(ch => (ch.id || ch._id) === chapterId);
            if (chapter && chapter.number > lastChapterNumber) {
                lastChapterNumber = chapter.number;
                lastChapterId = chapterId;
            }
        }
    }
    return lastChapterId;
}
function addResumeButton() {
    const lastChapterId = getLastReadChapter();
    const container = document.getElementById('resume-reading-container');
    if (!container) return;
    if (lastChapterId) {
        container.innerHTML = `<a href="${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(getCurrentMangaId())}&chapterId=${encodeURIComponent(lastChapterId)}" class="resume-reading-btn"><i class="fas fa-play"></i> متابعة القراءة</a>`;
    } else {
        container.innerHTML = '';
    }
}

// ========== إحصائيات متقدمة ==========
function updateStatsUi() {
    if (!currentManga) return;
    // Update favorites count if the element exists
    const favCountEl = document.getElementById('favorites-count');
    if (favCountEl) {
        favCountEl.textContent = formatCompactNumber(currentManga.favorites_count);
    }
    // Update manga rating display
    renderStars(currentManga.average_rating);
    
    // Update all rating displays across the page
    updateAllRatingDisplays();
}

// Update all rating displays on the page (for real-time updates)
function updateAllRatingDisplays() {
    if (!currentManga) return;
    const avgRating = currentManga.average_rating || 0;
    const ratingCount = currentManga.rating_count || 0;
    
    // Update rating in header/details section
    document.querySelectorAll('.manga-average-rating').forEach(el => {
        el.textContent = formatRating(avgRating);
    });
    document.querySelectorAll('.manga-rating-count').forEach(el => {
        el.textContent = `${ratingCount} تقييم`;
    });
    
    // Update any stat displays
    document.querySelectorAll('.stat-rating-value').forEach(el => {
        el.textContent = formatRating(avgRating);
    });
}

// ========== عرض تفاصيل المانجا مع أزرار الإعجاب والمفضلة ==========
async function renderMangaDetails() {
    const container = document.getElementById('manga-details');
    if (!container || !currentManga) return;
    const cover = currentManga.cover_image || '';
    galleryImages = [cover, ...(currentManga.gallery || [])].filter(Boolean);
    
    // استخدم حالة محايدة أولاً حتى لا يكون هناك Promise غير محلول في الواجهة
    let initialFavState = false;
    const mangaId = getCurrentMangaId();
    try {
        initialFavState = await isFavorite(mangaId);
    } catch (error) {
        console.debug('Could not resolve initial favorite state:', error);
        initialFavState = false;
    }

    // استخراج أول وأحدث فصل للربط بالأزرار
    const firstChapterId = currentChapters.length > 0 ? (currentChapters[currentChapters.length - 1].id || currentChapters[currentChapters.length - 1]._id) : null;
    const latestChapterId = currentChapters.length > 0 ? (currentChapters[0].id || currentChapters[0]._id) : null;

    container.innerHTML = `
        <div class="manga-hero-section">
            <div class="manga-hero-sidebar">
                <div class="manga-cover-wrapper">
                    <img src="${escapeHtml(cover)}" alt="${escapeHtml(currentManga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                </div>
                <div class="manga-stats-box">
                    <div class="stat-item">
                        <i class="fas fa-star" style="color: #fbbf24;"></i>
                        <strong>${formatRating(currentManga.average_rating)}</strong>
                        <small>التقييم</small>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-list" style="color: #10b981;"></i>
                        <strong>${currentChapters.length}</strong>
                        <small>فصل</small>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-bookmark" style="color: #8b5cf6;"></i>
                        <strong id="favorites-count">${formatCompactNumber(currentManga.favorites_count)}</strong>
                        <small>حفظ</small>
                    </div>
                    <div class="stat-item">
                        <i class="fas fa-eye" style="color: #06b6d4;"></i>
                        <strong>${formatCompactNumber(currentManga.views_count)}</strong>
                        <small>مشاهدة</small>
                    </div>
                </div>
            </div>

            <div class="manga-hero-content">
                <h1>${escapeHtml(currentManga.title)}</h1>
                
                <div class="manga-tags">
                    ${(currentManga.tags || []).map(tag => `<span class="tag-item">${escapeHtml(tag)}</span>`).join('')}
                </div>
                
                <div class="manga-synopsis">
                    <p>${escapeHtml(currentManga.description || 'لا يوجد وصف متاح لهذه المانجا.')}</p>
                </div>
                
                <div class="manga-action-buttons">
                    <button id="favorite-btn" class="btn ${initialFavState ? 'btn-primary' : 'btn-secondary'}">
                        <i class="fas fa-bookmark"></i> ${initialFavState ? 'مضافة للمفضلة' : 'إضافة للمفضلة'}
                    </button>
                    ${firstChapterId ? `<a href="${webPagePath('manga-reader.html')}?mangaId=${mangaId}&chapterId=${firstChapterId}" class="btn btn-secondary"><i class="fas fa-eye"></i> أول فصل</a>` : ''}
                    ${latestChapterId ? `<a href="${webPagePath('manga-reader.html')}?mangaId=${mangaId}&chapterId=${latestChapterId}" class="btn btn-secondary"><i class="fas fa-clock"></i> أحدث فصل</a>` : ''}
                </div>

                <div class="reactions-container-new">
                    <div id="reaction-counts-row" class="reaction-counts-row"></div>
                </div>
            </div>
        </div>
    `;
    
    // ربط الأحداث
    document.getElementById('favorite-btn')?.addEventListener('click', toggleFavorite);
    if (galleryImages.length > 1) initGallery(galleryImages);

    // تابع حالة المفضلة بعد التحميل وقم بالتحديث إذا تغيرت
    try {
        const resolvedFav = await isFavorite(mangaId);
        updateFavoriteButton(resolvedFav);
    } catch (error) {
        console.debug('Failed to refresh favorite state after render:', error);
    }
}

// ========== تحميل ردة فعل المستخدم الحالية ==========
async function loadUserReaction() {
    const mangaId = getCurrentMangaId();
    if (!mangaId) return;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/my-reaction`);
        currentReactionType = data.reaction_type || null;
        // Update UI to show reaction status - handled by updateReactionCounts()
    } catch (error) {
        console.debug('Failed to load user reaction:', error.message);
        currentReactionType = null;
    }
}

// ========== عرض الفصول المحسنة مع شريط تقدم ==========
async function renderChapters() {
    const container = document.getElementById('chapters-container');
    if (!container) return;
    if (!currentChapters.length) {
        container.innerHTML = `<div class="empty-state">لا توجد فصول بعد</div>`;
        return;
    }
    const bookmarks = JSON.parse(localStorage.getItem('readerBookmarks') || '{}');
    const mangaId = getCurrentMangaId();
    container.innerHTML = `<div class="chapters-header"><h2 class="section-title"><i class="fas fa-list"></i> الفصول</h2><span>${currentChapters.length} فصل</span></div><div class="chapters-list-enhanced" id="chapters-list"></div>`;
    const listContainer = document.getElementById('chapters-list');
    listContainer.innerHTML = currentChapters.map(chapter => {
        const chapterId = chapter.id || chapter._id;
        const pagesCount = chapter.pages?.length || 0;
        const bookmarkKey = `${mangaId}_${chapterId}`;
        const bookmark = bookmarks[bookmarkKey];
        const progress = bookmark ? ((bookmark.pageIndex + 1) / pagesCount) * 100 : 0;
        const hasViewed = chapter.has_user_viewed || false;
        return `
            <div class="chapter-item-enhanced">
                <div class="chapter-info">
                    <div class="chapter-number">الفصل ${chapter.number}</div>
                    <div class="chapter-title">${escapeHtml(chapter.title || '')}</div>
                    <div class="chapter-meta">
                        <span><i class="far fa-image"></i> ${pagesCount} صفحة</span>
                        <span><i class="fas fa-eye"></i> ${formatCompactNumber(chapter.views_count || 0)} مشاهدة</span>
                        <span><i class="fas fa-star"></i> ${formatRating(chapter.average_rating || 0)} (${chapter.rating_count || 0} تقييم)</span>
                        ${progress > 0 ? `<span><i class="fas fa-chart-line"></i> ${Math.round(progress)}% مكتمل</span>` : ''}
                    </div>
                    <div class="read-progress"><div class="read-progress-fill" style="width: ${progress}%;"></div></div>
                    ${hasViewed ? `
                        <div class="chapter-rating" data-chapter-id="${chapterId}">
                            <div class="rating-stars" id="chapter-stars-${chapterId}"></div>
                            <span class="rating-text">قيم هذا الفصل (1-10)</span>
                        </div>
                    ` : `
                        <div class="chapter-rating-placeholder">
                            <span class="rating-placeholder-text">اقرأ الفصل أولاً لتقييمه</span>
                        </div>
                    `}
                </div>
                <div class="chapter-actions">
                    <a href="${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}" class="btn btn-primary btn-sm"><i class="fas fa-book-open"></i> اقرأ</a>
                    <button class="btn btn-secondary btn-sm chapter-comments-btn" data-chapter-id="${chapterId}"><i class="fas fa-comments"></i> تعليقات</button>
                </div>
            </div>
        `;
    }).join('');

    // Initialize chapter ratings in limited concurrency to avoid too many concurrent my-rating calls
    const chaptersToRender = currentChapters
        .map(chapter => ({ chapter, chapterId: chapter.id || chapter._id }))
        .filter(({ chapter }) => chapter.has_user_viewed);

    const maxConcurrent = 4;
    let index = 0;

    async function renderNextChapter() {
        while (true) {
            let nextIndex;
            // lock-free retrieval; JS is single-threaded so this is safe here
            if (index >= chaptersToRender.length) return;
            nextIndex = index;
            index += 1;

            const chapterId = chaptersToRender[nextIndex].chapterId;
            try {
                await renderChapterStars(chapterId);
            } catch (error) {
                console.debug('Failed to render chapter stars for', chapterId, error);
            }
        }
    }

    await Promise.all(Array.from({ length: Math.min(maxConcurrent, chaptersToRender.length) }, () => renderNextChapter()));

    document.querySelectorAll('.chapter-comments-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const chapterId = btn.dataset.chapterId;
            openChapterCommentsModal(chapterId);
        });
    });
}

// ========== توصيات (مانجا مشابهة) ==========
async function loadRecommendations() {
    const container = document.getElementById('recommendations-grid');
    if (!container) return;
    
    try {
        container.innerHTML = '<div class="loading"><div class="loading-spinner"></div></div>';
        const data = await apiFetch('/mangas?limit=6');
        const mangas = data.items || data || [];
        
        if (!mangas.length) {
            container.innerHTML = '<div class="empty-state">لا توجد توصيات</div>';
            return;
        }
        
        container.innerHTML = mangas.slice(0, 6).map(manga => `
            <a href="${webPagePath('manga-details.html')}?id=${encodeURIComponent(manga.id || manga._id)}" class="manga-card-link">
                <div class="manga-card">
                    <img src="${escapeHtml(manga.cover_image)}" alt="${escapeHtml(manga.title)}" class="manga-thumbnail" onerror="this.src='https://via.placeholder.com/200x280?text=Manga'">
                    <div class="manga-info">
                        <h3>${escapeHtml(manga.title)}</h3>
                        <p class="manga-meta">${manga.chapters_count || 0} فصول</p>
                    </div>
                </div>
            </a>
        `).join('');
    } catch (error) {
        console.error('Error loading recommendations:', error);
        container.innerHTML = '<div class="error">فشل تحميل التوصيات</div>';
    }
}

// Track which mangas have had their view counted in this session to prevent double-counting
const viewedMangaSession = new Set();

// Record a view for the current manga (with per-manga session guard)
async function recordMangaView(mangaId) {
    if (!mangaId) return;
    // Per-manga guard: only record once per page session
    if (viewedMangaSession.has(mangaId)) return;
    viewedMangaSession.add(mangaId);

    try {
        // Use optional auth endpoint - works for both authenticated and anonymous users
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/view`, { method: 'POST' });
        // Update the in-memory view count after successful response
        if (currentManga) {
            currentManga.views_count = (currentManga.views_count || 0) + 1;
            // Update the views display in the stats box
            const viewsEl = document.querySelector('.stat-item[style*="fa-eye"] strong');
            if (viewsEl) {
                viewsEl.textContent = formatCompactNumber(currentManga.views_count);
            }
        }
    } catch (error) {
        // Silently fail - view tracking should not disrupt the user experience
        console.debug('Failed to record manga view:', error);
    }
}

// ========== تحميل الصفحة الرئيسية مع Skeleton ==========
async function loadMangaDetailsPage() {
    if (!requireAuth()) return;
    const mangaId = getQueryParam('id');
    if (!mangaId) { setError(document.getElementById('manga-details'), 'معرّف مفقود'); return; }
    // عرض Skeleton
    document.getElementById('manga-details').innerHTML = `<div class="skeleton-details"><div class="skeleton-cover"></div><div class="skeleton-info"><div class="skeleton-line"></div><div class="skeleton-line"></div><div class="skeleton-line short"></div></div></div>`;
    try {
        currentManga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
        const chapterResp = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        currentChapters = chapterResp.chapters || [];
        renderMangaDetails();
        renderChapters();
        updateStatsUi();
        addResumeButton();
        await loadRecommendations();
        await loadUserReaction();
        await loadComments();
        updateReactionCounts();
        document.getElementById('manga-title-header').textContent = currentManga.title;
        document.getElementById('manga-subtitle').textContent = `${currentChapters.length} فصل | ${currentManga.tags?.length || 0} تصنيف`;
        // Record manga view for analytics and most-viewed rankings
        const viewMangaId = getCurrentMangaId();
        if (viewMangaId) {
            recordMangaView(viewMangaId);
        }
    } catch (error) {
        setError(document.getElementById('manga-details'), error.message);
    }
}

// ========== تسجيل الخروج ==========
async function handleLogout() { try { await apiFetch('/auth/logout', { method: 'POST' }); } catch {} logoutLocal(); }

// ========== تهيئة المودال والأحداث ==========
function initModals() {
    const modal = document.getElementById('gallery-modal');
    if (!modal) return;
    
    // إغلاق المودال عند النقر على زر الإغلاق
    const closeBtn = modal.querySelector('.close-modal');
    if (closeBtn) {
        closeBtn.addEventListener('click', () => {
            modal.style.display = 'none';
        });
    }
    
    // إغلاق المودال عند النقر خارجه
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });
    
    // إغلاق جميع المودالات عند الضغط على ESC
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            document.querySelectorAll('.modal').forEach(m => {
                m.style.display = 'none';
            });
        }
    });
}

// ========== تعليقات المانجا ==========
let mangaComments = [];
let commentsPage = 1;
let commentsTotal = 0;
let commentsLoading = false;

async function loadComments(page = 1, sortOrder = null) {
    if (commentsLoading) return;
    commentsLoading = true;

    const container = document.getElementById('comments-list');
    const loading = document.getElementById('comments-loading');
    const empty = document.getElementById('comments-empty');

    if (page === 1) {
        container.innerHTML = '';
        loading.style.display = 'block';
        empty.style.display = 'none';
    }

    try {
        const mangaId = getCurrentMangaId();
        const sort = sortOrder || document.getElementById('comments-sort')?.value || 'newest';
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/comments?page=${page}&limit=10&sort=${sort}`);
        const newComments = data.data || [];

        if (page === 1) {
            mangaComments = newComments;
        } else {
            mangaComments = [...mangaComments, ...newComments];
        }

        commentsTotal = data.total || 0;
        commentsPage = page;

        renderComments();

        if (mangaComments.length === 0) {
            empty.style.display = 'block';
        }
    } catch (error) {
        console.error('Failed to load comments:', error);
        if (page === 1) {
            container.innerHTML = '<div class="error-state">فشل في تحميل التعليقات</div>';
        }
    } finally {
        loading.style.display = 'none';
        commentsLoading = false;
    }
}

function renderComments() {
    const container = document.getElementById('comments-list');
    if (!mangaComments.length) return;

    container.innerHTML = mangaComments.map(comment => `
        <div class="comment-item" data-comment-id="${comment.id}">
            <div class="comment-header">
                <div class="comment-author">
                    <div class="comment-avatar">${(comment.author_name || 'مستخدم').substring(0, 2)}</div>
                    <span class="comment-author-name">${escapeHtml(comment.author_name || 'مستخدم')}</span>
                    <span class="comment-date">${formatCommentDate(comment.created_at)}</span>
                </div>
                ${comment.can_delete ? `<button class="comment-delete-btn" data-comment-id="${comment.id}"><i class="fas fa-trash"></i></button>` : ''}
            </div>
            <div class="comment-content">${escapeHtml(comment.content)}</div>
        </div>
    `).join('');

    // Add delete event listeners
    container.querySelectorAll('.comment-delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const commentId = e.currentTarget.dataset.commentId;
            deleteComment(commentId);
        });
    });

    // Add load more button if there are more comments
    if (mangaComments.length < commentsTotal) {
        const loadMoreBtn = document.createElement('button');
        loadMoreBtn.id = 'load-more-comments';
        loadMoreBtn.className = 'btn btn-secondary';
        loadMoreBtn.innerHTML = '<i class="fas fa-plus"></i> تحميل المزيد';
        loadMoreBtn.addEventListener('click', () => loadComments(commentsPage + 1));
        container.appendChild(loadMoreBtn);
    }
}

async function submitComment() {
    const content = document.getElementById('comment-content').value.trim();
    if (!content) {
        showToast('يرجى كتابة تعليق', 'warning');
        return;
    }

    const submitBtn = document.getElementById('submit-comment');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = 'جاري الإرسال...';

    try {
        const mangaId = getCurrentMangaId();
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/comments`, {
            method: 'POST',
            body: JSON.stringify({ content })
        });

        document.getElementById('comment-content').value = '';
        hideCommentForm();
        showToast('تم إرسال التعليق بنجاح', 'success');
        await loadComments(1); // Reload comments
    } catch (error) {
        showToast(error.message || 'فشل في إرسال التعليق', 'error');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

async function deleteComment(commentId) {
    if (!confirm('هل أنت متأكد من حذف هذا التعليق؟')) return;

    try {
        const mangaId = getCurrentMangaId();
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/comments/${encodeURIComponent(commentId)}`, {
            method: 'DELETE'
        });

        showToast('تم حذف التعليق', 'success');
        await loadComments(1); // Reload comments
    } catch (error) {
        showToast(error.message || 'فشل في حذف التعليق', 'error');
    }
}

// ========== تعليقات الفصول ==========
let currentChapterComments = [];
let currentChapterId = null;
let chapterCommentsPage = 1;
let chapterCommentsTotal = 0;
let chapterCommentsLoading = false;

async function openChapterCommentsModal(chapterId) {
    currentChapterId = chapterId;
    currentChapterComments = [];
    chapterCommentsPage = 1;
    chapterCommentsTotal = 0;

    const modal = document.getElementById('chapter-comments-modal');
    if (!modal) {
        // Create modal if it doesn't exist
        createChapterCommentsModal();
    }

    document.getElementById('chapter-comments-modal').style.display = 'block';
    const sortOrder = document.getElementById('chapter-comments-sort')?.value || 'newest';
    await loadChapterComments(1, sortOrder);
}

function createChapterCommentsModal() {
    const modal = document.createElement('div');
    modal.id = 'chapter-comments-modal';
    modal.className = 'modal';
    modal.innerHTML = `
        <div class="modal-content chapter-comments-modal-content">
            <div class="modal-header">
                <h3>تعليقات الفصل</h3>
                <span class="close-modal" id="close-chapter-comments">&times;</span>
            </div>
            <div class="comments-controls" style="margin-bottom: 1rem; display: flex; justify-content: space-between; align-items: center;">
                <div class="sort-controls">
                    <label for="chapter-comments-sort" style="margin-right: 0.5rem;">ترتيب:</label>
                    <select id="chapter-comments-sort" class="form-control" style="padding: 0.25rem 0.5rem; border: 1px solid #ddd; border-radius: 4px;">
                        <option value="newest">الأحدث أولاً</option>
                        <option value="oldest">الأقدم أولاً</option>
                    </select>
                </div>
            </div>
            <div id="chapter-comments-container">
                <div id="add-chapter-comment-form" class="comment-form" style="display:none;">
                    <textarea id="chapter-comment-content" placeholder="اكتب تعليقك على هذا الفصل..." rows="3" maxlength="1000"></textarea>
                    <div class="comment-actions">
                        <button id="submit-chapter-comment" class="btn btn-primary">إرسال</button>
                        <button id="cancel-chapter-comment" class="btn btn-secondary">إلغاء</button>
                    </div>
                </div>
                <button id="show-chapter-comment-form" class="btn btn-secondary" style="margin-bottom: 1rem;"><i class="fas fa-plus"></i> أضف تعليق</button>
                <div id="chapter-comments-list" class="comments-list"></div>
                <div id="chapter-comments-loading" class="loading-state" style="display:none;">جاري تحميل التعليقات...</div>
                <div id="chapter-comments-empty" class="empty-state" style="display:none;">لا توجد تعليقات على هذا الفصل بعد.</div>
            </div>
        </div>
    `;
    document.body.appendChild(modal);

    // Add event listeners
    document.getElementById('close-chapter-comments').addEventListener('click', () => {
        document.getElementById('chapter-comments-modal').style.display = 'none';
    });

    // Add sort change listener
    document.getElementById('chapter-comments-sort').addEventListener('change', (e) => {
        loadChapterComments(1, e.target.value);
    });

    document.getElementById('show-chapter-comment-form').addEventListener('click', () => {
        document.getElementById('add-chapter-comment-form').style.display = 'block';
        document.getElementById('show-chapter-comment-form').style.display = 'none';
    });

    document.getElementById('cancel-chapter-comment').addEventListener('click', () => {
        document.getElementById('add-chapter-comment-form').style.display = 'none';
        document.getElementById('show-chapter-comment-form').style.display = 'block';
        document.getElementById('chapter-comment-content').value = '';
    });

    document.getElementById('submit-chapter-comment').addEventListener('click', submitChapterComment);

    // Close modal when clicking outside
    modal.addEventListener('click', (e) => {
        if (e.target === modal) {
            modal.style.display = 'none';
        }
    });
}

async function loadChapterComments(page = 1, sortOrder = 'newest') {
    if (chapterCommentsLoading || !currentChapterId) return;

    chapterCommentsLoading = true;
    const loading = document.getElementById('chapter-comments-loading');
    const empty = document.getElementById('chapter-comments-empty');
    const list = document.getElementById('chapter-comments-list');

    if (page === 1) {
        loading.style.display = 'block';
        empty.style.display = 'none';
        list.innerHTML = '';
    }

    try {
        const mangaId = getCurrentMangaId();
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(currentChapterId)}/comments?page=${page}&limit=20&sort=${sortOrder}`);

        if (page === 1) {
            currentChapterComments = data.data || [];
        } else {
            currentChapterComments = [...currentChapterComments, ...(data.data || [])];
        }

        chapterCommentsTotal = data.total || 0;
        chapterCommentsPage = page;

        renderChapterComments();

        if (currentChapterComments.length === 0) {
            empty.style.display = 'block';
        }
    } catch (error) {
        console.error('Failed to load chapter comments:', error);
        if (page === 1) {
            list.innerHTML = '<div class="error-state">فشل في تحميل التعليقات</div>';
        }
    } finally {
        loading.style.display = 'none';
        chapterCommentsLoading = false;
    }
}

function renderChapterComments() {
    const container = document.getElementById('chapter-comments-list');
    if (!currentChapterComments.length) return;

    container.innerHTML = currentChapterComments.map(comment => `
        <div class="comment-item" data-comment-id="${comment.id}">
            <div class="comment-header">
                <div class="comment-author">
                    <div class="comment-avatar">${(comment.author_name || 'مستخدم').substring(0, 2)}</div>
                    <span class="comment-author-name">${escapeHtml(comment.author_name || 'مستخدم')}</span>
                    <span class="comment-date">${formatCommentDate(comment.created_at)}</span>
                </div>
                ${comment.can_delete ? `<button class="comment-delete-btn" data-comment-id="${comment.id}"><i class="fas fa-trash"></i></button>` : ''}
            </div>
            <div class="comment-content">${escapeHtml(comment.content)}</div>
        </div>
    `).join('');

    // Add delete event listeners
    container.querySelectorAll('.comment-delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const commentId = e.currentTarget.dataset.commentId;
            deleteChapterComment(commentId);
        });
    });

    // Add load more button if there are more comments
    if (currentChapterComments.length < chapterCommentsTotal) {
        const loadMoreBtn = document.createElement('button');
        loadMoreBtn.id = 'load-more-chapter-comments';
        loadMoreBtn.className = 'btn btn-secondary';
        loadMoreBtn.innerHTML = '<i class="fas fa-plus"></i> تحميل المزيد';
        loadMoreBtn.addEventListener('click', () => {
            const sortOrder = document.getElementById('chapter-comments-sort').value;
            loadChapterComments(chapterCommentsPage + 1, sortOrder);
        });
        container.appendChild(loadMoreBtn);
    }
}

async function submitChapterComment() {
    const content = document.getElementById('chapter-comment-content').value.trim();
    if (!content) {
        showToast('يرجى كتابة تعليق', 'warning');
        return;
    }

    const submitBtn = document.getElementById('submit-chapter-comment');
    const originalText = submitBtn.textContent;
    submitBtn.disabled = true;
    submitBtn.textContent = 'جاري الإرسال...';

    try {
        const mangaId = getCurrentMangaId();
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(currentChapterId)}/comments`, {
            method: 'POST',
            body: JSON.stringify({ content })
        });

        document.getElementById('chapter-comment-content').value = '';
        document.getElementById('add-chapter-comment-form').style.display = 'none';
        document.getElementById('show-chapter-comment-form').style.display = 'block';
        showToast('تم إرسال التعليق بنجاح', 'success');
        const sortOrder = document.getElementById('chapter-comments-sort').value;
        await loadChapterComments(1, sortOrder); // Reload comments
    } catch (error) {
        showToast(error.message || 'فشل في إرسال التعليق', 'error');
    } finally {
        submitBtn.disabled = false;
        submitBtn.textContent = originalText;
    }
}

async function deleteChapterComment(commentId) {
    if (!confirm('هل أنت متأكد من حذف هذا التعليق؟')) return;

    try {
        const mangaId = getCurrentMangaId();
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(currentChapterId)}/comments/${encodeURIComponent(commentId)}`, {
            method: 'DELETE'
        });

        showToast('تم حذف التعليق', 'success');
        const sortOrder = document.getElementById('chapter-comments-sort').value;
        await loadChapterComments(1, sortOrder); // Reload comments
    } catch (error) {
        showToast(error.message || 'فشل في حذف التعليق', 'error');
    }
}

function showCommentForm() {
    document.getElementById('add-comment-form').style.display = 'block';
    document.getElementById('show-comment-form').style.display = 'none';
    document.getElementById('comment-content').focus();
}

function hideCommentForm() {
    document.getElementById('add-comment-form').style.display = 'none';
    document.getElementById('show-comment-form').style.display = 'block';
    document.getElementById('comment-content').value = '';
}

function formatCommentDate(dateStr) {
    const date = new Date(dateStr);
    const now = new Date();
    const diffMs = now - date;
    const diffHours = diffMs / (1000 * 60 * 60);
    const diffDays = diffMs / (1000 * 60 * 60 * 24);

    if (diffHours < 1) return 'الآن';
    if (diffHours < 24) return `منذ ${Math.floor(diffHours)} ساعة`;
    if (diffDays < 7) return `منذ ${Math.floor(diffDays)} يوم`;
    return date.toLocaleDateString('ar-SA');
}

document.addEventListener('DOMContentLoaded', () => {
    loadMangaDetailsPage();
    document.getElementById('logout-button')?.addEventListener('click', handleLogout);
    initModals();
    window.shareManga = shareManga;

    // Comments event listeners
    document.getElementById('show-comment-form')?.addEventListener('click', showCommentForm);
    document.getElementById('cancel-comment')?.addEventListener('click', hideCommentForm);
    document.getElementById('submit-comment')?.addEventListener('click', submitComment);
    document.getElementById('comments-sort')?.addEventListener('change', (e) => {
        loadComments(1, e.target.value);
    });
});
// ========== Reaction Picker System ==========
const REACTION_TYPES = {
    upvote: { emoji: '👍', label: 'إيجابي', color: 'var(--reaction-upvote)' },
    funny: { emoji: '😂', label: 'مضحك', color: 'var(--reaction-funny)' },
    love: { emoji: '❤️', label: 'حب', color: 'var(--reaction-love)' },
    surprised: { emoji: '😮', label: 'مندهش', color: 'var(--reaction-surprised)' },
    angry: { emoji: '😡', label: 'غضب', color: 'var(--reaction-angry)' },
    sad: { emoji: '😢', label: 'حزن', color: 'var(--reaction-sad)' }
};

function toggleReactionPicker() {
    const picker = document.getElementById('reaction-picker');
    if (!picker) return;
    picker.classList.toggle('active');
}

async function handleReaction(reactionType, originElement) {
    const mangaId = getCurrentMangaId();
    if (!mangaId || reactionInFlight) return;
    
    // Store original state for rollback
    const previousReaction = currentReactionType;
    const originalCounts = { ...currentManga.reactions_count };
    
    // Optimistic UI update - update both state and counts
    currentReactionType = reactionType;
    
    // Perform optimistic count mutation
    if (!currentManga.reactions_count) currentManga.reactions_count = {};
    
    // If there was a previous reaction, decrement it
    if (previousReaction) {
        currentManga.reactions_count[previousReaction] = Math.max(0, (currentManga.reactions_count[previousReaction] || 0) - 1);
    }
    
    // If new reaction is different from previous, increment it
    // If same reaction, it means we're toggling off (already decremented above)
    if (reactionType !== previousReaction) {
        currentManga.reactions_count[reactionType] = (currentManga.reactions_count[reactionType] || 0) + 1;
    } else {
        // Same reaction clicked - we're removing it, so currentReactionType should be null
        currentReactionType = null;
    }
    
    updateReactionCounts();
    
    reactionInFlight = true;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/react`, { 
            method: 'POST',
            body: JSON.stringify({ type: reactionType })
        });
        currentManga = data.manga || { ...currentManga, ...data };
        currentReactionType = data.removed ? null : (data.reaction_type || reactionType);
        
        updateStatsUi();
        updateReactionCounts();
        animateReactionConfetti(reactionType, originElement);
        
        const label = REACTION_TYPES[reactionType]?.label || reactionType;
        showToast(`ردة فعلك: ${label} ${REACTION_TYPES[reactionType]?.emoji}`, 'success');
        
        const picker = document.getElementById('reaction-picker');
        if (picker) picker.classList.remove('active');
    } catch (error) {
        // Revert optimistic update on error
        currentReactionType = previousReaction;
        currentManga.reactions_count = originalCounts;
        updateReactionCounts();
        showToast(error.message, 'error');
    } finally {
        reactionInFlight = false;
    }
}

// Debounced reaction functions
const debouncedHandleReaction = debounce((reactionType, originElement) => handleReaction(reactionType, originElement), 300);
const debouncedHandleLikeToggle = debounce(handleLikeToggleEnhanced, 500);

function animateReactionConfetti(reactionType, originElement) {
    const emoji = REACTION_TYPES[reactionType]?.emoji || '👍';
    const btn = originElement || document.querySelector('.reaction-bar-btn');
    if (!btn) return;
    
    const rect = btn.getBoundingClientRect();
    for (let i = 0; i < 8; i++) {
        const particle = document.createElement('div');
        particle.style.cssText = `
            position: fixed;
            left: ${rect.left + rect.width/2}px;
            top: ${rect.top + rect.height/2}px;
            font-size: 1.5rem;
            pointer-events: none;
            z-index: 1000;
        `;
        particle.textContent = emoji;
        particle.style.opacity = '1';
        document.body.appendChild(particle);
        
        // Random directions
        const angle = (Math.PI / 4) * (Math.random() - 0.5); // -22.5 to 22.5 degrees
        const speed = 2 + Math.random() * 2; // 2-4 px per frame
        
        let x = 0;
        let y = 0;
        let opacity = 1;
        
        function animate() {
            x += Math.sin(angle) * speed;
            y -= Math.cos(angle) * speed;
            opacity = Math.max(0, 1 - Math.abs(y) / 100);
            
            particle.style.transform = `translate(${x}px, ${y}px)`;
            particle.style.opacity = opacity;
            
            if (opacity > 0) {
                requestAnimationFrame(animate);
            } else {
                particle.remove();
            }
        }
        
        requestAnimationFrame(animate);
    }
}

function updateReactionCounts() {
    const countsContainer = document.getElementById('reaction-counts-row');
    if (!countsContainer) return;
    
    const reactionsCount = currentManga.reactions_count || {};
    const totalReactions = Object.values(reactionsCount).reduce((a, b) => a + b, 0);
    
    // Preserve scroll position
    const scrollTop = countsContainer.scrollTop;
    
    // الترتيب المطلوب ظهوره دائماً (مثل الصورة الثانية)
    const order = ['upvote', 'funny', 'love', 'surprised', 'angry', 'sad'];
    
    countsContainer.innerHTML = `
        <div class="reactions-header">
            <h3>ما رأيك في هذا العمل؟</h3>
            <span class="reactions-total">${totalReactions} تفاعل</span>
        </div>
        <div class="reactions-grid-new">
            ${order.map(type => {
                const reaction = REACTION_TYPES[type];
                const count = reactionsCount[type] || 0;
                const isUserReacted = currentReactionType === type;
                return `
                    <div class="reaction-item-new ${isUserReacted ? 'active' : ''}" data-reaction-type="${type}">
                        <div class="reaction-emoji-new">${reaction.emoji}</div>
                        <div class="reaction-count-new">${formatCompactNumber(count)}</div>
                        <div class="reaction-label-new">${reaction.label}</div>
                    </div>
                `;
            }).join('')}
        </div>
    `;
    
    // Restore scroll position
    countsContainer.scrollTop = scrollTop;
    
    // Bind event listeners programmatically
    document.querySelectorAll('.reaction-item-new').forEach(item => {
        item.addEventListener('click', () => {
            const reactionType = item.dataset.reactionType;
            if (reactionType) {
                // Add animating class temporarily
                item.classList.add('animating');
                setTimeout(() => item.classList.remove('animating'), 350);
                debouncedHandleReaction(reactionType, item);
            }
        });
    });
}











