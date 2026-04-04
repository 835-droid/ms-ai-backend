// Novels List Page - JavaScript
// Handles all interactions for the novels listing page

(function() {
    'use strict';

    // State
    let currentPage = 1;
    let currentTab = 'all';
    let currentSort = 'newest';
    let currentGenre = '';
    let searchQuery = '';
    let isLoading = false;

    // DOM Elements
    const novelsGrid = document.getElementById('novels-grid');
    const loadingState = document.getElementById('loading-state');
    const emptyState = document.getElementById('empty-state');
    const pagination = document.getElementById('pagination');
    const searchInput = document.getElementById('search-input');
    const searchBtn = document.getElementById('search-btn');
    const sortSelect = document.getElementById('sort-select');
    const genreSelect = document.getElementById('genre-select');
    const tabs = document.querySelectorAll('.tab');

    // Initialize page
    document.addEventListener('DOMContentLoaded', () => {
        loadNovels();
        setupEventListeners();
    });

    // Setup event listeners
    function setupEventListeners() {
        // Tab clicks
        tabs.forEach(tab => {
            tab.addEventListener('click', () => {
                tabs.forEach(t => t.classList.remove('active'));
                tab.classList.add('active');
                currentTab = tab.dataset.tab;
                currentPage = 1;
                loadNovels();
            });
        });

        // Search
        searchBtn.addEventListener('click', () => {
            searchQuery = searchInput.value.trim();
            currentPage = 1;
            loadNovels();
        });

        searchInput.addEventListener('keypress', (e) => {
            if (e.key === 'Enter') {
                searchQuery = searchInput.value.trim();
                currentPage = 1;
                loadNovels();
            }
        });

        // Sort change
        sortSelect.addEventListener('change', () => {
            currentSort = sortSelect.value;
            currentPage = 1;
            loadNovels();
        });

        // Genre change
        genreSelect.addEventListener('change', () => {
            currentGenre = genreSelect.value;
            currentPage = 1;
            loadNovels();
        });
    }

    // Load novels based on current state
    async function loadNovels() {
        if (isLoading) return;
        isLoading = true;

        showLoading();

        try {
            let novels;

            switch (currentTab) {
                case 'popular':
                    novels = await API.getMostViewedNovels('all', 20);
                    break;
                case 'recent':
                    novels = await API.getRecentlyUpdatedNovels(20);
                    break;
                case 'top-rated':
                    novels = await API.getTopRatedNovels(20);
                    break;
                default:
                    novels = await API.getNovels(currentPage, 20);
                    break;
            }

            // Apply search filter if needed
            let items = novels.items || novels;
            if (searchQuery) {
                items = items.filter(novel => 
                    novel.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
                    (novel.description && novel.description.toLowerCase().includes(searchQuery.toLowerCase())) ||
                    (novel.tags && novel.tags.some(tag => tag.toLowerCase().includes(searchQuery.toLowerCase())))
                );
            }

            // Apply genre filter if needed
            if (currentGenre) {
                items = items.filter(novel => 
                    novel.tags && novel.tags.includes(currentGenre)
                );
            }

            // Apply sorting
            items = sortNovels(items, currentSort);

            renderNovels(items);
            renderPagination(novels.total || items.length, 20);

            if (items.length === 0) {
                showEmpty();
            } else {
                hideEmpty();
            }
        } catch (error) {
            console.error('Error loading novels:', error);
            showToast('فشل تحميل الروايات', 'error');
        } finally {
            isLoading = false;
            hideLoading();
        }
    }

    // Sort novels array
    function sortNovels(novels, sortBy) {
        const sorted = [...novels];
        switch (sortBy) {
            case 'newest':
                return sorted.sort((a, b) => new Date(b.created_at) - new Date(a.created_at));
            case 'popular':
            case 'views':
                return sorted.sort((a, b) => (b.views_count || 0) - (a.views_count || 0));
            case 'rated':
                return sorted.sort((a, b) => (b.average_rating || 0) - (a.average_rating || 0));
            default:
                return sorted;
        }
    }

    // Render novels grid
    function renderNovels(novels) {
        novelsGrid.innerHTML = novels.map(novel => `
            <div class="manga-card">
                <a href="novel-details.html?id=${novel.id}">
                    <div class="manga-cover-wrapper">
                        <img src="${novel.cover_image || 'placeholder-manga.jpg'}" 
                             alt="${novel.title}" 
                             onerror="this.src='placeholder-manga.jpg'">
                        ${novel.average_rating > 0 ? `
                            <div class="manga-rating-badge">
                                <i class="fas fa-star"></i> ${novel.average_rating.toFixed(1)}
                            </div>
                        ` : ''}
                    </div>
                    <div class="manga-card-info">
                        <h3 class="manga-title">${truncate(novel.title, 40)}</h3>
                        <div class="manga-meta">
                            <span><i class="fas fa-eye"></i> ${(novel.views_count || 0).toLocaleString()}</span>
                            <span><i class="fas fa-heart"></i> ${(novel.favorites_count || 0).toLocaleString()}</span>
                        </div>
                        ${novel.tags && novel.tags.length > 0 ? `
                            <div class="manga-tags">
                                ${novel.tags.slice(0, 3).map(tag => `<span class="tag tag-sm">${tag}</span>`).join('')}
                            </div>
                        ` : ''}
                    </div>
                </a>
            </div>
        `).join('');
    }

    // Render pagination
    function renderPagination(total, perPage) {
        const totalPages = Math.ceil(total / perPage);
        
        if (totalPages <= 1) {
            pagination.innerHTML = '';
            return;
        }

        let html = '';

        // Previous button
        if (currentPage > 1) {
            html += `<button class="pagination-btn" onclick="goToPage(${currentPage - 1})">
                <i class="fas fa-chevron-right"></i> السابق
            </button>`;
        }

        // Page numbers
        for (let i = 1; i <= Math.min(totalPages, 5); i++) {
            if (i === currentPage) {
                html += `<button class="pagination-btn active">${i}</button>`;
            } else {
                html += `<button class="pagination-btn" onclick="goToPage(${i})">${i}</button>`;
            }
        }

        // Next button
        if (currentPage < totalPages) {
            html += `<button class="pagination-btn" onclick="goToPage(${currentPage + 1})">
                التالي <i class="fas fa-chevron-left"></i>
            </button>`;
        }

        pagination.innerHTML = html;
    }

    // Go to page
    window.goToPage = function(page) {
        currentPage = page;
        loadNovels();
        window.scrollTo({ top: 0, behavior: 'smooth' });
    };

    // Show/hide loading
    function showLoading() {
        loadingState.style.display = 'block';
        novelsGrid.style.display = 'none';
    }

    function hideLoading() {
        loadingState.style.display = 'none';
        novelsGrid.style.display = 'grid';
    }

    // Show/hide empty state
    function showEmpty() {
        emptyState.style.display = 'block';
    }

    function hideEmpty() {
        emptyState.style.display = 'none';
    }

    // Truncate text
    function truncate(text, maxLength) {
        if (!text) return '';
        if (text.length <= maxLength) return text;
        return text.substring(0, maxLength) + '...';
    }

})();