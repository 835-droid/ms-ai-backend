// manga-details-enhanced.js - نسخة متكاملة مع إعجاب ومفضلة وتعليقات
let currentManga = null;
let currentChapters = [];
let likeInFlight = false;
let ratingInFlight = false;
let favoriteInFlight = false;
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

function getCurrentMangaId() {
    return currentManga?.id || currentManga?._id || getQueryParam('id');
}

// ========== إدارة المفضلة (LocalStorage + API إذا وُجد) ==========
async function loadFavorites() {
    try {
        const data = await apiFetch('/mangas/favorites/list');
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
            showToast('تمت إزالة من المفضلة', 'info');
        } else {
            await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/favorite`, { method: 'POST' });
            showToast('أضيفت إلى المفضلة', 'success');
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
        currentReactionType = data.reaction_type || 'upvote';
        updateStatsUi();
        showToast('تم تسجيل ردة فعلك! ♥', 'success');
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
    const container = document.getElementById('stars-container');
    if (!container) return;
    const fullStars = Math.floor(ratingValue);
    const hasHalf = ratingValue % 1 >= 0.5;
    container.innerHTML = '';
    for (let i = 1; i <= 5; i++) {
        const star = document.createElement('i');
        star.className = `fas fa-star star ${i <= fullStars ? 'active' : ''}`;
        if (i === fullStars + 1 && hasHalf && !(i <= fullStars)) {
            star.className = 'fas fa-star-half-alt star active';
        }
        star.dataset.value = i;
        if (interactive) {
            star.addEventListener('mouseenter', () => highlightStars(i));
            star.addEventListener('mouseleave', () => resetStars());
            star.addEventListener('click', () => submitRating(i));
        }
        container.appendChild(star);
    }
    const avgSpan = document.getElementById('rating-average');
    if (avgSpan) avgSpan.textContent = `(${formatRating(currentManga?.average_rating)} من 5)`;
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
/*
async function submitRating(score) {
    if (ratingInFlight) return;
    const mangaId = getCurrentMangaId();
    if (!mangaId) return;
    ratingInFlight = true;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/rate`, {
            method: 'POST',
            body: JSON.stringify({ score })
        });
        currentManga = data.manga || { ...currentManga, ...data };
        updateStatsUi();
        renderStars(currentManga.average_rating || 0, true);
        showToast(`قيمت بـ ${score} نجوم`, 'success');
    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        ratingInFlight = false;
    }
}
*/

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

// ========== تحميل الفصل كملف PDF (معاينة) ==========
function downloadChapterAsPDF(chapter) {
    const pages = chapter.pages || [];
    if (!pages.length) { showToast('لا توجد صفحات', 'error'); return; }
    const htmlContent = `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>${chapter.title} - الفصل ${chapter.number}</title><style>img{max-width:100%;margin:20px 0;}</style></head><body dir="rtl"><h1>${chapter.title}</h1>${pages.map(url => `<img src="${url}">`).join('')}</body></html>`;
    const win = window.open();
    win.document.write(htmlContent);
    win.print();
    showToast('تم فتح معاينة للطباعة', 'info');
}

// ========== إحصائيات متقدمة ==========
function updateStatsUi() {
    if (!currentManga) return;
    document.getElementById('stat-views').innerHTML = `<div class="stat-number">${formatCompactNumber(currentManga.views_count)}</div><div class="stat-label">مشاهدة</div>`;
    document.getElementById('stat-likes').innerHTML = `<div class="stat-number">${formatCompactNumber(currentManga.likes_count)}</div><div class="stat-label">ردود فعل</div>`;
    document.getElementById('stat-rating').innerHTML = `<div class="stat-number">${formatRating(currentManga.average_rating)}</div><div class="stat-label">التقييم</div>`;
    document.getElementById('stat-chapters').innerHTML = `<div class="stat-number">${currentChapters.length}</div><div class="stat-label">فصل</div>`;
}

// ========== عرض تفاصيل المانجا مع أزرار الإعجاب والمفضلة ==========
function renderMangaDetails() {
    const container = document.getElementById('manga-details');
    if (!container || !currentManga) return;
    const cover = currentManga.cover_image || '';
    galleryImages = [cover, ...(currentManga.gallery || [])].filter(Boolean);
    const isFav = isFavorite(getCurrentMangaId());
    
    // استخراج أول وأحدث فصل للربط بالأزرار
    const mangaId = getCurrentMangaId();
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
                        <strong>${formatCompactNumber(currentManga.likes_count)}</strong>
                        <small>حفظ</small>
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
                    <button id="favorite-btn" class="btn ${isFav ? 'btn-primary' : 'btn-secondary'}">
                        <i class="fas fa-bookmark"></i> ${isFav ? 'مضافة للمفضلة' : 'إضافة للمفضلة'}
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
}

// ========== تحميل ردة فعل المستخدم الحالية ==========
async function loadUserReaction() {
    const mangaId = getCurrentMangaId();
    if (!mangaId) return;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/my-reaction`);
        currentReactionType = data.reaction_type || null;
        // Update UI to show reaction status
        updateReactionUI();
    } catch (error) {
        console.debug('Failed to load user reaction:', error.message);
        currentReactionType = null;
    }
}

// تحديث واجهة المستخدم لعرض حالة التفاعل
function updateReactionUI() {
    const btn = document.getElementById('like-button-enhanced');
    if (!btn) return;
    if (currentReactionType) {
        btn.classList.add('has-reaction');
        btn.dataset.reaction = currentReactionType;
    } else {
        btn.classList.remove('has-reaction');
        btn.removeAttribute('data-reaction');
    }
}

// ========== عرض الفصول المحسنة مع شريط تقدم ==========
function renderChapters() {
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
                </div>
                <div class="chapter-actions">
                    <a href="${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}" class="btn btn-primary btn-sm"><i class="fas fa-book-open"></i> اقرأ</a>
                    <button class="btn btn-secondary btn-sm download-chapter" data-chapter-id="${chapterId}"><i class="fas fa-download"></i></button>
                </div>
            </div>
        `;
    }).join('');
    document.querySelectorAll('.download-chapter').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const chapterId = btn.dataset.chapterId;
            const chapter = currentChapters.find(ch => (ch.id || ch._id) === chapterId);
            if (chapter) downloadChapterAsPDF(chapter);
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

async function loadComments(page = 1) {
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
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/comments?page=${page}&limit=10`);
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
                    <div class="comment-avatar">${(comment.author_name || 'مستخدم')[0].toUpperCase()}</div>
                    <span class="comment-author-name">${escapeHtml(comment.author_name || 'مستخدم')}</span>
                    <span class="comment-date">${formatCommentDate(comment.created_at)}</span>
                </div>
                ${comment.can_delete ? `<button class="comment-delete-btn" data-comment-id="${comment.id}"><i class="fas fa-trash"></i></button>` : ''}
            </div>
            <div class="comment-content">${escapeHtml(comment.content)}</div>
        </div>
    `).join('');

    // Add delete event listeners
    document.querySelectorAll('.comment-delete-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            const commentId = e.currentTarget.dataset.commentId;
            deleteComment(commentId);
        });
    });
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

async function handleReaction(reactionType) {
    const mangaId = getCurrentMangaId();
    if (!mangaId || likeInFlight) return;
    
    likeInFlight = true;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/react`, { 
            method: 'POST',
            body: JSON.stringify({ type: reactionType })
        });
        currentManga = data.manga || { ...currentManga, ...data };
        currentReactionType = data.reaction_type || reactionType;
        
        updateStatsUi();
        updateReactionUI();
        updateReactionCounts();
        animateReactionConfetti(reactionType);
        
        const label = REACTION_TYPES[reactionType]?.label || reactionType;
        showToast(`ردة فعلك: ${label} ${REACTION_TYPES[reactionType]?.emoji}`, 'success');
        
        const picker = document.getElementById('reaction-picker');
        if (picker) picker.classList.remove('active');
    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        likeInFlight = false;
    }
}

function animateReactionConfetti(reactionType) {
    const emoji = REACTION_TYPES[reactionType]?.emoji || '👍';
    const btn = document.querySelector('.reaction-bar-btn');
    if (!btn) return;
    
    const rect = btn.getBoundingClientRect();
    for (let i = 0; i < 15; i++) {
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
        
        // Animate particle
        let y = 0;
        const interval = setInterval(() => {
            y -= 2;
            particle.style.transform = `translateY(${y}px)`;
            particle.style.opacity = Math.max(0, (500 - y) / 500);
            if (y <= -100) {
                clearInterval(interval);
                particle.remove();
            }
        }, 10);
    }
}

function updateReactionUI() {
    const btn = document.querySelector('.reaction-bar-btn');
    if (!btn) return;
    
    if (currentReactionType) {
        btn.classList.add('has-reaction');
        const reaction = REACTION_TYPES[currentReactionType];
        if (reaction) {
            btn.innerHTML = `${reaction.emoji} ${reaction.label} <i class="fas fa-chevron-down"></i>`;
        }
    } else {
        btn.classList.remove('has-reaction');
        btn.innerHTML = '<i class="fas fa-smile"></i> تفاعل <i class="fas fa-chevron-down"></i>';
    }
}

function updateReactionCounts() {
    const countsContainer = document.getElementById('reaction-counts-row');
    if (!countsContainer) return;
    
    const reactionsCount = currentManga.reactions_count || {};
    const totalReactions = Object.values(reactionsCount).reduce((a, b) => a + b, 0);
    
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
                    <div class="reaction-item-new ${isUserReacted ? 'active' : ''}" onclick="handleReaction('${type}')">
                        <div class="reaction-emoji-new">${reaction.emoji}</div>
                        <div class="reaction-count-new">${formatCompactNumber(count)}</div>
                        <div class="reaction-label-new">${reaction.label}</div>
                    </div>
                `;
            }).join('')}
        </div>
    `;
}











