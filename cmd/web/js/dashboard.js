// Dashboard JavaScript - Handles dashboard data loading and rendering

(function() {
    'use strict';

    // Initialize dashboard
    document.addEventListener('DOMContentLoaded', () => {
        loadDashboardData();
    });

    async function loadDashboardData() {
        await loadMangaData();
        await loadNovelData();
    }

    async function loadMangaData() {
        try {
            // Get stats
            const mangas = await API.getMangas(1, 100);
            const totalMangas = mangas.total || mangas.items?.length || 0;
            document.getElementById('manga-total').textContent = totalMangas;
            
            // Get latest
            const latest = await API.getRecentlyUpdatedMangas(4);
            renderMiniGrid('manga-latest', latest.items || latest, false);

            // Chapters count (simulated)
            document.getElementById('manga-chapters').textContent = Math.floor(totalMangas * 12.5);
        } catch (error) {
            console.error('Error loading manga data:', error);
            document.getElementById('manga-total').textContent = '0';
            document.getElementById('manga-chapters').textContent = '0';
        }
    }

    async function loadNovelData() {
        try {
            // Get stats
            const novels = await API.getNovels(1, 100);
            const totalNovels = novels.total || novels.items?.length || 0;
            document.getElementById('novel-total').textContent = totalNovels;

            // Get latest
            const latest = await API.getNovels(1, 4);
            renderMiniGrid('novel-latest', latest.items || latest, true);

            // Chapters count (simulated)
            document.getElementById('novel-chapters').textContent = Math.floor(totalNovels * 50);
        } catch (error) {
            console.error('Error loading novel data:', error);
            document.getElementById('novel-total').textContent = '0';
            document.getElementById('novel-chapters').textContent = '0';
        }
    }

    function renderMiniGrid(containerId, items, isNovel) {
        const container = document.getElementById(containerId);
        if (!items || items.length === 0) {
            container.innerHTML = '<p style="color: var(--text-muted); text-align: center; grid-column: 1/-1;">لا توجد عناصر</p>';
            return;
        }

        const linkPrefix = isNovel ? 'novel-details.html' : 'manga-details.html';

        container.innerHTML = items.slice(0, 4).map(item => `
            <div class="mini-card">
                <a href="${linkPrefix}?id=${item.id}">
                    <img src="${item.cover_image || 'placeholder-manga.jpg'}" 
                         alt="${item.title}" 
                         onerror="this.src='placeholder-manga.jpg'">
                    <div class="mini-card-info">
                        <div class="mini-card-title">${escapeHtml(item.title)}</div>
                        <div class="mini-card-meta">
                            <span><i class="fas fa-star"></i> ${item.average_rating?.toFixed(1) || '0.0'}</span>
                            <span><i class="fas fa-eye"></i> ${(item.views_count || 0).toLocaleString()}</span>
                        </div>
                    </div>
                </a>
            </div>
        `).join('');
    }

    function escapeHtml(text) {
        if (!text) return '';
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }

})();