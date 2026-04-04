// Novel Details Page - JavaScript
//handles all interactions for the novel details page

(function() {
    'use strict';

    // Get novel ID from URL
    const urlParams = new URLSearchParams(window.location.search);
    const novelID = urlParams.get('id');

    if (!novelID) {
        showToast('معرف الرواية مطلوب', 'error');
        setTimeout(() => window.location.href = 'novels.html', 2000);
        return;
    }

    // State
    let currentNovel = null;
    let userReaction = '';
    let userRating = 0;
    let isFavorite = false;

    // DOM Elements
    const detailsContainer = document.getElementById('novel-details');
    const chaptersContainer = document.getElementById('chapters-container');
    const resumeContainer = document.getElementById('resume-reading-container');
    const commentsList = document.getElementById('comments-list');
    const recommendationsGrid = document.getElementById('recommendations-grid');
    const commentForm = document.getElementById('add-comment-form');
    const commentContent = document.getElementById('comment-content');
    const submitCommentBtn = document.getElementById('submit-comment');
    const cancelCommentBtn = document.getElementById('cancel-comment');
    const showCommentFormBtn = document.getElementById('show-comment-form');
    const commentsSort = document.getElementById('comments-sort');

    // Initialize page
    document.addEventListener('DOMContentLoaded', () => {
        loadNovelDetails();
        loadNovelComments();
        loadRecommendations();
        setupEventListeners();
    });

    // Load novel details
    async function loadNovelDetails() {
        try {
            const novel = await API.getNovel(novelID);
            currentNovel = novel;
            renderNovelDetails(novel);
            await loadNovelChapters(novelID);
            checkUserEngagement();
            updateReadingProgress();
        } catch (error) {
            console.error('Error loading novel:', error);
            showMessage('فشل تحميل تفاصيل الرواية', 'error');
        }
    }

    // Render novel details
    function renderNovelDetails(novel) {
        document.title = `MS-AI | ${novel.title}`;
        document.getElementById('novel-title-header').textContent = novel.title;

        const stars = '★'.repeat(Math.round(novel.average_rating || 0)) + '☆'.repeat(10 - Math.round(novel.average_rating || 0));

        detailsContainer.innerHTML = `
            <div class="details-hero">
                <div class="details-cover">
                    <img src="${novel.cover_image || 'placeholder-manga.jpg'}" alt="${novel.title}" 
                         onerror="this.src='placeholder-manga.jpg'">
                </div>
                <div class="details-info">
                    <h2 class="details-title">${novel.title}</h2>
                    <div class="details-meta">
                        <span class="rating"><i class="fas fa-star"></i> ${novel.average_rating?.toFixed(1) || '0.0'} (${novel.rating_count || 0} تقييم)</span>
                        <span class="views"><i class="fas fa-eye"></i> ${(novel.views_count || 0).toLocaleString()}</span>
                        <span class="favorites"><i class="fas fa-heart"></i> ${(novel.favorites_count || 0).toLocaleString()}</span>
                    </div>
                    <div class="details-tags">
                        ${novel.tags?.map(tag => `<span class="tag">${tag}</span>`).join('') || ''}
                    </div>
                    <p class="details-description">${novel.description || 'لا يوجد وصف متاح'}</p>
                    <div class="details-actions">
                        <button id="btn-favorite" class="btn btn-primary" onclick="toggleFavorite()">
                            <i class="fas fa-heart"></i> ${isFavorite ? 'إزالة من المفضلة' : 'إضافة للمفضلة'}
                        </button>
                        <button id="btn-rate" class="btn btn-secondary" onclick="showRatingModal()">
                            <i class="fas fa-star"></i> قيم الرواية
                        </button>
                        <button id="btn-react" class="btn btn-secondary" onclick="showReactionModal()">
                            <i class="fas fa-thumbs-up"></i> تفاعل
                        </button>
                    </div>
                    <div class="details-stats">
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.upvote || 0}</span>
                            <span class="stat-label">👍</span>
                        </div>
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.funny || 0}</span>
                            <span class="stat-label">😂</span>
                        </div>
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.love || 0}</span>
                            <span class="stat-label">❤️</span>
                        </div>
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.surprised || 0}</span>
                            <span class="stat-label">😮</span>
                        </div>
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.angry || 0}</span>
                            <span class="stat-label">😠</span>
                        </div>
                        <div class="stat">
                            <span class="stat-value">${novel.reactions_count?.sad || 0}</span>
                            <span class="stat-label">😢</span>
                        </div>
                    </div>
                </div>
            </div>
        `;
    }

    // Load novel chapters
    async function loadNovelChapters(novelId) {
        try {
            // For now, show a message that chapters are coming soon
            chaptersContainer.innerHTML = `
                <div class="chapters-header">
                    <h3 class="section-title"><i class="fas fa-list"></i> الفصول</h3>
                </div>
                <div class="chapters-grid">
                    <div class="empty-state">
                        <i class="fas fa-book-open fa-3x"></i>
                        <p>الفصول قادمة قريباً</p>
                    </div>
                </div>
            `;
        } catch (error) {
            console.error('Error loading chapters:', error);
        }
    }

    // Check user engagement (favorite, rating, reaction)
    async function checkUserEngagement() {
        if (!Auth.isLoggedIn()) return;

        try {
            // Check if favorited
            const favStatus = await API.checkNovelFavorite(novelID);
            isFavorite = favStatus.is_favorite;

            // Get user reaction
            const reaction = await API.getNovelUserReaction(novelID);
            userReaction = reaction.reaction;

            // Update UI
            updateFavoriteButton();
        } catch (error) {
            console.error('Error checking engagement:', error);
        }
    }

    // Update reading progress
    async function updateReadingProgress() {
        try {
            const progress = await API.getNovelReadingProgress(novelID);
            if (progress && progress.last_read_chapter) {
                resumeContainer.innerHTML = `
                    <a href="novel-reader.html?id=${novelID}&chapter=${progress.last_read_chapter}" 
                       class="btn btn-primary btn-lg">
                        <i class="fas fa-bookmark"></i> متابعة القراءة
                    </a>
                `;
            }
        } catch (error) {
            // Silently fail if no reading progress
        }
    }

    // Load novel comments
    async function loadNovelComments() {
        try {
            const sort = commentsSort.value;
            const response = await API.getNovelComments(novelID, 1, 20, sort);

            if (response.items.length === 0) {
                document.getElementById('comments-empty').style.display = 'block';
                return;
            }

            commentsList.innerHTML = response.items.map(comment => `
                <div class="comment-item" data-id="${comment.id}">
                    <div class="comment-header">
                        <span class="comment-author">${comment.username || 'مستخدم'}</span>
                        <span class="comment-date">${new Date(comment.created_at).toLocaleDateString('ar-SA')}</span>
                        ${isOwnComment(comment) ? `
                            <button class="btn btn-sm btn-danger" onclick="deleteComment('${comment.id}')">
                                <i class="fas fa-trash"></i>
                            </button>
                        ` : ''}
                    </div>
                    <p class="comment-content">${escapeHtml(comment.content)}</p>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error loading comments:', error);
        }
    }

    // Load recommendations
    async function loadRecommendations() {
        try {
            const response = await API.getNovels(1, 8);
            const otherNovels = response.items.filter(n => n.id !== novelID).slice(0, 4);

            recommendationsGrid.innerHTML = otherNovels.map(novel => `
                <div class="manga-card">
                    <a href="novel-details.html?id=${novel.id}">
                        <img src="${novel.cover_image || 'placeholder-manga.jpg'}" 
                             alt="${novel.title}" 
                             onerror="this.src='placeholder-manga.jpg'">
                        <div class="manga-card-info">
                            <h3 class="manga-title">${novel.title}</h3>
                            <div class="manga-rating">
                                <i class="fas fa-star"></i> ${novel.average_rating?.toFixed(1) || '0.0'}
                            </div>
                        </div>
                    </a>
                </div>
            `).join('');
        } catch (error) {
            console.error('Error loading recommendations:', error);
        }
    }

    // Setup event listeners
    function setupEventListeners() {
        // Comment form toggle
        showCommentFormBtn.addEventListener('click', () => {
            if (!Auth.isLoggedIn()) {
                showToast('يجب تسجيل الدخول لإضافة تعليق', 'warning');
                return;
            }
            commentForm.style.display = commentForm.style.display === 'none' ? 'block' : 'none';
        });

        cancelCommentBtn.addEventListener('click', () => {
            commentForm.style.display = 'none';
            commentContent.value = '';
        });

        submitCommentBtn.addEventListener('click', submitComment);

        // Comments sort
        commentsSort.addEventListener('change', () => loadNovelComments());
    }

    // Submit comment
    async function submitComment() {
        const content = commentContent.value.trim();
        if (!content) {
            showToast('الرجاء كتابة تعليق', 'warning');
            return;
        }

        try {
            await API.addNovelComment(novelID, content);
            showToast('تم إضافة التعليق بنجاح', 'success');
            commentContent.value = '';
            commentForm.style.display = 'none';
            loadNovelComments();
        } catch (error) {
            showToast('فشل إضافة التعليق', 'error');
        }
    }

    // Delete comment
    window.deleteComment = async function(commentId) {
        if (!confirm('هل أنت متأكد من حذف هذا التعليق؟')) return;

        try {
            await API.deleteNovelComment(novelID, commentId);
            showToast('تم حذف التعليق', 'success');
            loadNovelComments();
        } catch (error) {
            showToast('فشل حذف التعليق', 'error');
        }
    };

    // Toggle favorite
    window.toggleFavorite = async function() {
        if (!Auth.isLoggedIn()) {
            showToast('يجب تسجيل الدخول', 'warning');
            return;
        }

        try {
            if (isFavorite) {
                await API.removeNovelFavorite(novelID);
                isFavorite = false;
                showToast('تمت الإزالة من المفضلة', 'success');
            } else {
                await API.addNovelFavorite(novelID);
                isFavorite = true;
                showToast('تمت الإضافة للمفضلة', 'success');
            }
            updateFavoriteButton();
            loadNovelDetails(); // Refresh to update count
        } catch (error) {
            showToast('فشل التحديث', 'error');
        }
    };

    // Update favorite button
    function updateFavoriteButton() {
        const btn = document.getElementById('btn-favorite');
        if (btn) {
            btn.innerHTML = isFavorite ? 
                '<i class="fas fa-heart"></i> إزالة من المفضلة' : 
                '<i class="fas fa-heart"></i> إضافة للمفضلة';
            btn.classList.toggle('btn-danger', isFavorite);
            btn.classList.toggle('btn-primary', !isFavorite);
        }
    }

    // Show rating modal
    window.showRatingModal = function() {
        if (!Auth.isLoggedIn()) {
            showToast('يجب تسجيل الدخول للتقييم', 'warning');
            return;
        }

        const rating = prompt('قيّم الرواية من 1 إلى 10:');
        if (rating) {
            const score = parseInt(rating);
            if (score >= 1 && score <= 10) {
                rateNovel(score);
            } else {
                showToast('التقييم يجب أن يكون بين 1 و 10', 'warning');
            }
        }
    };

    // Rate novel
    async function rateNovel(score) {
        try {
            await API.rateNovel(novelID, score);
            showToast('تم التقييم بنجاح', 'success');
            loadNovelDetails(); // Refresh
        } catch (error) {
            showToast('فشل التقييم', 'error');
        }
    }

    // Show reaction modal
    window.showReactionModal = function() {
        if (!Auth.isLoggedIn()) {
            showToast('يجب تسجيل الدخول للتفاعل', 'warning');
            return;
        }

        const reactions = [
            { type: 'upvote', emoji: '👍', label: 'إعجاب' },
            { type: 'funny', emoji: '😂', label: 'مضحك' },
            { type: 'love', emoji: '❤️', label: 'محب' },
            { type: 'surprised', emoji: '😮', label: 'مندهش' },
            { type: 'angry', emoji: '😠', label: 'غاضب' },
            { type: 'sad', emoji: '😢', label: 'حزين' }
        ];

        const choices = reactions.map(r => `${r.emoji} ${r.label}`).join('\n');
        const choice = prompt(`اختر رد الفعل:\n${choices}\n\nاكتب الرقم (1-6):`);

        if (choice) {
            const index = parseInt(choice) - 1;
            if (index >= 0 && index < reactions.length) {
                setReaction(reactions[index].type);
            }
        }
    };

    // Set reaction
    async function setReaction(type) {
        try {
            await API.setNovelReaction(novelID, type);
            showToast('تم التفاعل بنجاح', 'success');
            loadNovelDetails(); // Refresh
        } catch (error) {
            showToast('فشل التفاعل', 'error');
        }
    }

    // Helper functions
    function isOwnComment(comment) {
        // Check if current user is the comment author
        return Auth.getCurrentUser() && comment.user_id === Auth.getCurrentUser().id;
    }

    function escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

    function showMessage(text, type = 'info') {
        const msgDiv = document.getElementById('details-message');
        msgDiv.textContent = text;
        msgDiv.className = `message message-${type}`;
        msgDiv.style.display = 'block';
    }

})();