// History Page Logic

let currentPage = 1;
let totalPages = 1;
let limit = 20;

// ========== API Helpers ==========
async function fetchHistory(page = 1, limitVal = limit) {
    try {
        const data = await apiFetch(`/mangas/history?page=${page}&limit=${limitVal}`);
        return data;
    } catch (error) {
        console.error('Failed to fetch history:', error);
        return { items: [], total: 0, total_pages: 0, current_page: 1 };
    }
}

async function fetchStats() {
    try {
        const data = await apiFetch('/mangas/history/stats');
        return data;
    } catch (error) {
        console.error('Failed to fetch stats:', error);
        return { total_views: 0, unique_manga: 0, unique_chapters: 0, total_duration: 0 };
    }
}

async function deleteHistoryItem(id) {
    try {
        await apiFetch(`/mangas/history/${encodeURIComponent(id)}`, {
            method: 'DELETE'
        });
        return true;
    } catch (error) {
        console.error('Failed to delete history item:', error);
        throw error;
    }
}

async function cleanOldHistory(days = 90) {
    try {
        const data = await apiFetch(`/mangas/history/clean?days=${days}`, {
            method: 'DELETE'
        });
        return data;
    } catch (error) {
        console.error('Failed to clean old history:', error);
        throw error;
    }
}

// ========== Render Functions ==========
function renderStats(stats) {
    document.getElementById('stat-total').textContent = stats.unique_manga || 0;
    document.getElementById('stat-chapters').textContent = stats.unique_chapters || 0;
    document.getElementById('stat-recent').textContent = stats.total_views || 0;
}

function renderHistory(historyData) {
    const container = document.getElementById('history-list');
    const emptyState = document.getElementById('history-empty');
    const countLabel = document.getElementById('history-count');
    
    if (!container) return;

    const items = historyData.items || [];
    
    if (items.length === 0) {
        container.style.display = 'none';
        if (emptyState) emptyState.style.display = 'block';
        if (countLabel) countLabel.textContent = '';
        return;
    }

    container.style.display = 'flex';
    if (emptyState) emptyState.style.display = 'none';
    
    if (countLabel) {
        countLabel.textContent = `عرض ${items.length} من ${historyData.total || 0} عنصر`;
    }

    container.innerHTML = items.map(item => {
        const manga = item.manga;
        if (!manga) return '';
        
        const mangaID = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/60x80?text=M';
        const viewedAt = item.viewed_at ? new Date(item.viewed_at).toLocaleDateString('ar-EG') : '';
        const chapterID = item.chapter_id;
        
        return `
            <div class="history-item" data-id="${item.id}">
                <img class="history-cover" src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" 
                     onerror="this.src='https://via.placeholder.com/60x80?text=M'">
                <div class="history-info">
                    <div class="history-title">
                        <a href="manga-details.html?id=${encodeURIComponent(mangaID)}">${escapeHtml(manga.title)}</a>
                    </div>
                    <div class="history-meta">
                        ${chapterID ? `<span class="chapter-badge"><i class="fas fa-book-open"></i> فصل</span>` : ''}
                        <span><i class="fas fa-clock"></i> ${viewedAt}</span>
                    </div>
                </div>
                <div class="history-actions">
                    <button class="btn-icon delete-history-btn" data-id="${item.id}" title="حذف من السجل">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Add event listeners for delete buttons
    container.querySelectorAll('.delete-history-btn').forEach(btn => {
        btn.addEventListener('click', async () => {
            const id = btn.dataset.id;
            if (confirm('هل تريد حذف هذا العنصر من السجل؟')) {
                try {
                    await deleteHistoryItem(id);
                    showToast('تم الحذف بنجاح', 'success');
                    loadHistory(currentPage);
                } catch (error) {
                    showToast('فشل الحذف: ' + (error.message || 'خطأ غير معروف'), 'error');
                }
            }
        });
    });
}

function renderPagination(data) {
    const container = document.getElementById('pagination');
    if (!container) return;

    totalPages = data.total_pages || 1;
    currentPage = data.current_page || 1;

    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }

    let html = '';
    
    // Previous button
    html += `<button ${currentPage <= 1 ? 'disabled' : ''} onclick="loadHistory(${currentPage - 1})">
        <i class="fas fa-chevron-right"></i> السابق
    </button>`;

    // Page numbers
    const maxVisible = 5;
    let startPage = Math.max(1, currentPage - Math.floor(maxVisible / 2));
    let endPage = Math.min(totalPages, startPage + maxVisible - 1);

    if (endPage - startPage + 1 < maxVisible) {
        startPage = Math.max(1, endPage - maxVisible + 1);
    }

    if (startPage > 1) {
        html += `<button onclick="loadHistory(1)">1</button>`;
        if (startPage > 2) {
            html += `<button disabled>...</button>`;
        }
    }

    for (let i = startPage; i <= endPage; i++) {
        html += `<button class="${i === currentPage ? 'active' : ''}" onclick="loadHistory(${i})">${i}</button>`;
    }

    if (endPage < totalPages) {
        if (endPage < totalPages - 1) {
            html += `<button disabled>...</button>`;
        }
        html += `<button onclick="loadHistory(${totalPages})">${totalPages}</button>`;
    }

    // Next button
    html += `<button ${currentPage >= totalPages ? 'disabled' : ''} onclick="loadHistory(${currentPage + 1})">
        التالي <i class="fas fa-chevron-left"></i>
    </button>`;

    container.innerHTML = html;
}

// ========== Load Functions ==========
async function loadHistory(page = 1) {
    currentPage = page;
    
    const data = await fetchHistory(page, limit);
    renderHistory(data);
    renderPagination(data);
}

async function loadStats() {
    const stats = await fetchStats();
    renderStats(stats);
}

// ========== Event Handlers ==========
async function handleCleanOldHistory() {
    const days = prompt('كم يوماً من السجل تريد الاحتفاظ به؟ (افتراضي: 90)', '90');
    if (!days) return;
    
    const daysNum = parseInt(days);
    if (isNaN(daysNum) || daysNum <= 0) {
        showToast('الرجاء إدخال رقم صحيح', 'error');
        return;
    }

    if (!confirm(`هل أنت متأكد من حذف جميع السجلات الأقدم من ${daysNum} يوماً؟`)) {
        return;
    }

    try {
        const result = await cleanOldHistory(daysNum);
        showToast(`تم حذف ${result.deleted_count || 0} عنصر من السجل`, 'success');
        loadHistory(currentPage);
        loadStats();
    } catch (error) {
        showToast('فشل تنظيف السجل: ' + (error.message || 'خطأ غير معروف'), 'error');
    }
}

// ========== Initialization ==========
document.addEventListener('DOMContentLoaded', () => {
    if (!requireAuth()) return;

    // Load initial data
    loadHistory();
    loadStats();

    // Clean old history button
    document.getElementById('btn-clean-old')?.addEventListener('click', handleCleanOldHistory);

    // Logout button
    document.getElementById('logout-button')?.addEventListener('click', async () => {
        try {
            await apiFetch('/auth/logout', { method: 'POST' });
        } catch { /* ignore */ }
        logoutLocal(true);
    });
});

// Make loadHistory globally available for pagination
window.loadHistory = loadHistory;