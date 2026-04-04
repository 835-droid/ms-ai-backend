// admin-new.js - لوحة إدارة متطورة بالكامل (النسخة الجديدة)

// ========== State Management ==========
let currentSection = 'dashboard';
let allMangas = [];
let allAuthors = [];
let selectedMangas = new Set();
let currentPage = 1;
let itemsPerPage = 15;
let tagList = [];
let editTagList = [];
let currentEditMangaId = '';
let uploadedFiles = [];
let novelTagList = [];

// ========== Utility Functions ==========
function escapeHtml(text) {
    if (!text) return '';
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function formatCompactNumber(value) {
    const number = Number(value || 0);
    if (number >= 1000000) return `${(number / 1000000).toFixed(1)}M`;
    if (number >= 1000) return `${(number / 1000).toFixed(1)}K`;
    return String(number);
}

function showToast(message, type = 'success') {
    const container = document.getElementById('toast-container');
    const toast = document.createElement('div');
    toast.className = `toast ${type}`;
    toast.textContent = message;
    container.appendChild(toast);
    setTimeout(() => toast.remove(), 3000);
}

// ========== Navigation ==========
function switchSection(sectionId) {
    // Update nav items
    document.querySelectorAll('.admin-nav-item').forEach(item => {
        item.classList.toggle('active', item.dataset.section === sectionId);
    });

    // Update sections
    document.querySelectorAll('.admin-section').forEach(section => {
        section.classList.toggle('active', section.id === `section-${sectionId}`);
    });

    currentSection = sectionId;

    // Load section data
    switch (sectionId) {
        case 'dashboard':
            loadDashboard();
            break;
        case 'mangas':
            loadMangasList();
            break;
        case 'add-manga':
            loadAddMangaData();
            break;
        case 'chapters':
            loadChaptersData();
            break;
        case 'authors':
            loadAuthors();
            break;
    }

    // Close mobile sidebar
    document.getElementById('admin-sidebar').classList.remove('open');
}

// ========== Dashboard ==========
async function loadDashboard() {
    try {
        const data = await apiFetch('/admin/metrics');
        document.getElementById('stat-total-mangas').textContent = data.total_mangas || 0;
        document.getElementById('stat-total-chapters').textContent = data.total_chapters || 0;
        document.getElementById('stat-total-views').textContent = formatCompactNumber(data.total_views || 0);
        document.getElementById('stat-total-likes').textContent = formatCompactNumber(data.total_likes || 0);

        // Load recent activity
        loadRecentActivity();
    } catch (error) {
        console.error('Failed to load dashboard:', error);
    }
}

async function loadRecentActivity() {
    const container = document.getElementById('recent-activity');
    try {
        const data = await apiFetch('/mangas?limit=5');
        const mangas = data.items || [];

        if (mangas.length === 0) {
            container.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>لا توجد نشاطات حديثة</p></div>';
            return;
        }

        container.innerHTML = mangas.map(manga => `
            <div style="display: flex; align-items: center; gap: 1rem; padding: 0.75rem; background: rgba(255,255,255,0.02); border-radius: 10px; margin-bottom: 0.5rem;">
                <img src="${escapeHtml(manga.cover_image || 'https://via.placeholder.com/40x50')}" style="width: 40px; height: 50px; border-radius: 6px; object-fit: cover;" alt="">
                <div style="flex: 1;">
                    <div style="color: var(--admin-text); font-weight: 600;">${escapeHtml(manga.title)}</div>
                    <div style="font-size: 0.75rem; color: var(--admin-text-muted);">
                        👁 ${formatCompactNumber(manga.views_count || 0)} • ⭐ ${Number(manga.average_rating || 0).toFixed(1)}
                    </div>
                </div>
                <span class="status-badge status-${manga.status || 'published'}">${manga.status || 'منشور'}</span>
            </div>
        `).join('');
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-exclamation-circle"></i><p>فشل تحميل النشاطات</p></div>';
    }
}

// ========== Mangas Management ==========
async function loadMangasList() {
    const container = document.getElementById('manga-list-container');
    container.innerHTML = '<div class="loading"><div class="loading-spinner"></div><p style="color: var(--admin-text-muted);">جاري تحميل المانجا...</p></div>';

    try {
        const data = await apiFetch('/mangas?limit=1000');
        allMangas = data.items || [];
        loadAuthorsForFilter();
        renderMangasList();
    } catch (error) {
        container.innerHTML = `<div class="empty-state"><i class="fas fa-exclamation-circle"></i><p>فشل تحميل المانجا: ${escapeHtml(error.message)}</p></div>`;
    }
}

function loadAuthorsForFilter() {
    const filterSelect = document.getElementById('filter-author');
    const uniqueAuthors = [...new Set(allMangas.map(m => m.author_name).filter(Boolean))];
    filterSelect.innerHTML = '<option value="all">جميع المؤلفين</option>' +
        uniqueAuthors.map(name => `<option value="${escapeHtml(name)}">${escapeHtml(name)}</option>`).join('');
}

function renderMangasList() {
    const container = document.getElementById('manga-list-container');
    const searchTerm = document.getElementById('search-manga')?.value.toLowerCase() || '';
    const statusFilter = document.getElementById('filter-status')?.value || 'all';

    let filtered = allMangas.filter(manga => {
        const matchesSearch = manga.title.toLowerCase().includes(searchTerm) ||
            (manga.description || '').toLowerCase().includes(searchTerm);
        const matchesStatus = statusFilter === 'all' || (manga.status || 'published') === statusFilter;
        return matchesSearch && matchesStatus;
    });

    const totalPages = Math.ceil(filtered.length / itemsPerPage);
    const start = (currentPage - 1) * itemsPerPage;
    const paginated = filtered.slice(start, start + itemsPerPage);

    if (paginated.length === 0) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-book"></i><p>لا توجد مانجا</p></div>';
        renderPagination(0);
        return;
    }

    container.innerHTML = paginated.map(manga => {
        const id = manga.id || manga._id;
        const isSelected = selectedMangas.has(id);
        return `
            <div class="manga-item" data-manga-id="${escapeHtml(id)}">
                <input type="checkbox" class="manga-select-checkbox" data-id="${escapeHtml(id)}" ${isSelected ? 'checked' : ''}>
                <img class="manga-item-cover" src="${escapeHtml(manga.cover_image || 'https://via.placeholder.com/50x70')}" onerror="this.src='https://via.placeholder.com/50x70'" alt="">
                <div class="manga-item-info">
                    <div class="manga-item-title">${escapeHtml(manga.title)}</div>
                    <div class="manga-item-meta">
                        <span>✍️ ${escapeHtml(manga.author_name || 'غير معروف')}</span>
                        <span>📚 ${manga.chapters_count || 0} فصل</span>
                        <span>👁 ${formatCompactNumber(manga.views_count || 0)}</span>
                        <span>⭐ ${Number(manga.average_rating || 0).toFixed(1)}</span>
                        <span class="status-badge status-${manga.status || 'published'}">${manga.status || 'منشور'}</span>
                    </div>
                </div>
                <div class="manga-item-actions">
                    <button class="btn btn-sm btn-secondary" onclick="openEditMangaModal('${escapeHtml(id)}')" title="تعديل">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteManga('${escapeHtml(id)}')" title="حذف">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Bind checkbox events
    document.querySelectorAll('.manga-select-checkbox').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const id = e.target.dataset.id;
            if (e.target.checked) {
                selectedMangas.add(id);
            } else {
                selectedMangas.delete(id);
            }
            updateSelectedCount();
        });
    });

    renderPagination(totalPages);
}

function updateSelectedCount() {
    document.getElementById('selected-count').textContent = `${selectedMangas.size} مانجا محددة`;
    document.getElementById('batch-delete-btn').disabled = selectedMangas.size === 0;
}

function renderPagination(totalPages) {
    const container = document.getElementById('manga-pagination');
    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }

    let html = '';
    for (let i = 1; i <= totalPages; i++) {
        html += `<button class="btn btn-sm ${i === currentPage ? 'btn-primary' : 'btn-secondary'}" onclick="goToPage(${i})">${i}</button>`;
    }
    container.innerHTML = html;
}

function goToPage(page) {
    currentPage = page;
    renderMangasList();
}

async function deleteManga(mangaId) {
    if (!confirm('⚠️ هل أنت متأكد من حذف هذه المانجا وجميع فصولها؟')) return;

    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`, { method: 'DELETE' });
        showToast('تم حذف المانجا بنجاح', 'success');
        selectedMangas.delete(mangaId);
        loadMangasList();
    } catch (error) {
        showToast(error.message || 'فشل الحذف', 'error');
    }
}

async function batchDeleteMangas() {
    if (selectedMangas.size === 0) return;
    if (!confirm(`⚠️ هل أنت متأكد من حذف ${selectedMangas.size} مانجا؟`)) return;

    try {
        for (const id of selectedMangas) {
            await apiFetch(`/mangas/${encodeURIComponent(id)}`, { method: 'DELETE' });
        }
        showToast(`تم حذف ${selectedMangas.size} مانجا`, 'success');
        selectedMangas.clear();
        loadMangasList();
    } catch (error) {
        showToast(error.message || 'فشل الحذف', 'error');
    }
}

// ========== Add/Edit Manga ==========
function loadAddMangaData() {
    loadAuthorsForSelect('manga-author');
}

async function loadAuthorsForSelect(selectId) {
    const select = document.getElementById(selectId);
    try {
        // Try to fetch authors from API
        const data = await apiFetch('/authors');
        allAuthors = data || [];
        select.innerHTML = '<option value="">اختر المؤلف</option>' +
            allAuthors.map(a => `<option value="${escapeHtml(a.id || a._id)}">${escapeHtml(a.name)}</option>`).join('');
    } catch (error) {
        select.innerHTML = '<option value="">اختر المؤلف</option>';
    }
}

function resetAddMangaForm() {
    document.getElementById('add-manga-form').reset();
    tagList = [];
    renderTags();
    document.getElementById('cover-preview').innerHTML = '';
}

function resetAddNovelForm() {
    document.getElementById('add-novel-form').reset();
    novelTagList = [];
    renderNovelTags();
    document.getElementById('novel-cover-preview').innerHTML = '';
}

function renderNovelTags() {
    const container = document.getElementById('novel-tags-container');
    if (!container) return;
    container.innerHTML = novelTagList.map(tag => `
        <span class="tag-item">
            ${escapeHtml(tag)}
            <button type="button" onclick="removeNovelTag('${escapeHtml(tag)}')">×</button>
        </span>
    `).join('');
}

function addNovelTag(tag) {
    const cleaned = String(tag || '').trim();
    if (!cleaned || novelTagList.includes(cleaned)) return;
    novelTagList.push(cleaned);
    renderNovelTags();
}

function removeNovelTag(tag) {
    novelTagList = novelTagList.filter(t => t !== tag);
    renderNovelTags();
}

// ========== Tags Management ==========
function renderTags() {
    const container = document.getElementById('tags-container');
    if (!container) return;
    container.innerHTML = tagList.map(tag => `
        <span class="tag-item">
            ${escapeHtml(tag)}
            <button type="button" onclick="removeTag('${escapeHtml(tag)}')">×</button>
        </span>
    `).join('');
}

function addTag(tag) {
    const cleaned = String(tag || '').trim();
    if (!cleaned || tagList.includes(cleaned)) return;
    tagList.push(cleaned);
    renderTags();
}

function removeTag(tag) {
    tagList = tagList.filter(t => t !== tag);
    renderTags();
}

// ========== Chapters Management - Improved ==========
let allMangasForChapters = [];

async function loadChaptersData() {
    await loadMangasForChapters();
    renderQuickAddList();
    initQuickAddEvents();
}

async function loadMangasForChapters() {
    try {
        const data = await apiFetch('/mangas?limit=5000');
        allMangasForChapters = data.items || [];
    } catch (error) {
        console.error('Failed to load mangas for chapters:', error);
    }
}

function renderQuickAddList() {
    const container = document.getElementById('quick-add-manga-list');
    const searchTerm = document.getElementById('quick-add-search')?.value.toLowerCase() || '';
    const statusFilter = document.getElementById('quick-add-filter')?.value || 'all';

    let filtered = allMangasForChapters.filter(m => {
        const matchesSearch = m.title.toLowerCase().includes(searchTerm) ||
            (m.author_name || '').toLowerCase().includes(searchTerm);
        const matchesStatus = statusFilter === 'all' || (m.status || 'published') === statusFilter;
        return matchesSearch && matchesStatus;
    }).slice(0, 50);

    if (filtered.length === 0) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-search"></i><p>لا توجد مانجا</p></div>';
        return;
    }

    container.innerHTML = filtered.map(m => `
        <div class="manga-item" style="cursor: pointer;" data-manga-id="${escapeHtml(m.id || m._id)}">
            <img class="manga-item-cover" src="${escapeHtml(m.cover_image || 'https://via.placeholder.com/50x70')}" onerror="this.src='https://via.placeholder.com/50x70'" alt="">
            <div class="manga-item-info">
                <div class="manga-item-title">${escapeHtml(m.title)}</div>
                <div class="manga-item-meta">
                    <span>📚 ${m.chapters_count || 0} فصل</span>
                    <span class="status-badge status-${m.status || 'published'}">${m.status || 'منشور'}</span>
                </div>
            </div>
            <div class="manga-item-actions">
                <button class="btn btn-sm btn-primary quick-add-btn" data-manga-id="${escapeHtml(m.id || m._id)}" title="إضافة فصل">
                    <i class="fas fa-plus"></i>
                </button>
                <button class="btn btn-sm btn-secondary view-chapters-btn" data-manga-id="${escapeHtml(m.id || m._id)}" title="عرض الفصول">
                    <i class="fas fa-list"></i>
                </button>
            </div>
        </div>
    `).join('');

    // Bind events
    container.querySelectorAll('.quick-add-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            openAddChapterForm(btn.dataset.mangaId);
        });
    });

    container.querySelectorAll('.view-chapters-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            viewChapters(btn.dataset.mangaId);
        });
    });
}

function initQuickAddEvents() {
    document.getElementById('quick-add-search')?.addEventListener('input', debounce(() => {
        renderQuickAddList();
    }, 300));

    document.getElementById('quick-add-filter')?.addEventListener('change', () => {
        renderQuickAddList();
    });

    document.getElementById('cancel-chapter-add')?.addEventListener('click', closeAddChapterForm);
    document.getElementById('close-chapters-view')?.addEventListener('click', closeChaptersView);
}

function openAddChapterForm(mangaId) {
    const manga = allMangasForChapters.find(m => (m.id || m._id) === mangaId);
    if (!manga) return;

    document.getElementById('chapter-manga-id').value = mangaId;
    document.getElementById('selected-manga-cover-lg').src = manga.cover_image || 'https://via.placeholder.com/60x80';
    document.getElementById('selected-manga-title-lg').textContent = manga.title;
    document.getElementById('selected-manga-meta').textContent = `✍️ ${manga.author_name || 'غير معروف'} • 📚 ${manga.chapters_count || 0} فصل`;

    document.getElementById('add-chapter-panel').style.display = 'block';
    document.getElementById('add-chapter-form').reset();

    // Scroll to form
    document.getElementById('add-chapter-panel').scrollIntoView({ behavior: 'smooth' });
}

function closeAddChapterForm() {
    document.getElementById('add-chapter-panel').style.display = 'none';
    document.getElementById('chapter-manga-id').value = '';
}

function viewChapters(mangaId) {
    document.getElementById('chapters-view-panel').style.display = 'block';
    loadChaptersForManga(mangaId);
    document.getElementById('chapters-view-panel').scrollIntoView({ behavior: 'smooth' });
}

function closeChaptersView() {
    document.getElementById('chapters-view-panel').style.display = 'none';
}

async function loadChaptersForManga(mangaId) {
    const container = document.getElementById('admin-chapter-list');
    if (!mangaId) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-layer-group"></i><p>اختر مانجا لعرض فصولها</p></div>';
        return;
    }

    container.innerHTML = '<div class="loading"><div class="loading-spinner"></div></div>';

    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        const chapters = data.chapters || [];

        if (chapters.length === 0) {
            container.innerHTML = '<div class="empty-state"><i class="fas fa-inbox"></i><p>لا توجد فصول</p></div>';
            return;
        }

        container.innerHTML = chapters.map(chapter => `
            <div class="chapter-item">
                <div class="chapter-info">
                    <span class="chapter-number">#${chapter.number}</span>
                    <span class="chapter-title">${escapeHtml(chapter.title || `الفصل ${chapter.number}`)}</span>
                    <div class="chapter-meta">
                        <span>📄 ${chapter.pages?.length || 0} صفحة</span>
                        <span>👁 ${formatCompactNumber(chapter.views_count || 0)}</span>
                    </div>
                </div>
                <div class="chapter-actions">
                    <button class="btn btn-sm btn-secondary" onclick="openEditChapterModal('${escapeHtml(chapter.id || chapter._id)}')">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteChapter('${escapeHtml(mangaId)}', '${escapeHtml(chapter.id || chapter._id)}')">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-exclamation-circle"></i><p>فشل تحميل الفصول</p></div>';
    }
}

async function deleteChapter(mangaId, chapterId) {
    if (!confirm('هل أنت متأكد من حذف هذا الفصل؟')) return;

    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`, { method: 'DELETE' });
        showToast('تم حذف الفصل', 'success');
        loadChaptersForManga(mangaId);
    } catch (error) {
        showToast(error.message || 'فشل الحذف', 'error');
    }
}

// ========== Authors Management ==========
async function loadAuthors() {
    const container = document.getElementById('authors-list');
    container.innerHTML = '<div class="loading"><div class="loading-spinner"></div></div>';

    try {
        const data = await apiFetch('/authors');
        allAuthors = data || [];

        if (allAuthors.length === 0) {
            container.innerHTML = '<div class="empty-state"><i class="fas fa-users"></i><p>لا توجد مؤلفين</p></div>';
            return;
        }

        container.innerHTML = allAuthors.map(author => `
            <div class="manga-item">
                <img src="${escapeHtml(author.image || 'https://via.placeholder.com/50x50')}" style="width: 50px; height: 50px; border-radius: 50%; object-fit: cover;" alt="">
                <div class="manga-item-info">
                    <div class="manga-item-title">${escapeHtml(author.name)}</div>
                    <div class="manga-item-meta">${escapeHtml(author.bio || 'لا توجد سيرة ذاتية')}</div>
                </div>
                <div class="manga-item-actions">
                    <button class="btn btn-sm btn-secondary" onclick="openEditAuthorModal('${escapeHtml(author.id || author._id)}')">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger" onclick="deleteAuthor('${escapeHtml(author.id || author._id)}')">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `).join('');
    } catch (error) {
        container.innerHTML = '<div class="empty-state"><i class="fas fa-exclamation-circle"></i><p>فشل تحميل المؤلفين</p></div>';
    }
}

// ========== Edit Manga Modal ==========
function openEditMangaModal(mangaId) {
    const manga = allMangas.find(m => (m.id || m._id) === mangaId);
    if (!manga) return;

    currentEditMangaId = mangaId;
    document.getElementById('edit-manga-id').value = mangaId;
    document.getElementById('edit-manga-title').value = manga.title || '';
    document.getElementById('edit-manga-description').value = manga.description || '';
    document.getElementById('edit-manga-year').value = manga.year || '';
    document.getElementById('edit-manga-status').value = manga.status || 'published';

    document.getElementById('editMangaModal').classList.add('active');
}

// ========== Export Data ==========
async function exportData() {
    try {
        showToast('جاري تجهيز البيانات...', 'info');
        const data = await apiFetch('/mangas?limit=1000');
        const exportData = {
            mangas: data.items || [],
            exportedAt: new Date().toISOString()
        };

        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `manga-export-${new Date().toISOString().split('T')[0]}.json`;
        a.click();
        URL.revokeObjectURL(url);

        showToast('تم تصدير البيانات بنجاح', 'success');
    } catch (error) {
        showToast(error.message || 'فشل التصدير', 'error');
    }
}

// ========== Event Listeners ==========
document.addEventListener('DOMContentLoaded', () => {
    // Navigation
    document.querySelectorAll('.admin-nav-item').forEach(item => {
        item.addEventListener('click', () => {
            switchSection(item.dataset.section);
        });
    });

    // Search & Filter
    document.getElementById('search-manga')?.addEventListener('input', () => {
        currentPage = 1;
        renderMangasList();
    });

    document.getElementById('filter-status')?.addEventListener('change', () => {
        currentPage = 1;
        renderMangasList();
    });

    document.getElementById('filter-author')?.addEventListener('change', () => {
        currentPage = 1;
        renderMangasList();
    });

    // Select All
    document.getElementById('select-all-mangas')?.addEventListener('change', (e) => {
        if (e.target.checked) {
            allMangas.forEach(m => selectedMangas.add(m.id || m._id));
        } else {
            selectedMangas.clear();
        }
        renderMangasList();
        updateSelectedCount();
    });

    // Batch Delete
    document.getElementById('batch-delete-btn')?.addEventListener('click', batchDeleteMangas);

    // Manga Tags Input
    document.getElementById('manga-tags-input')?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            const input = e.target;
            const tags = input.value.split(',').map(t => t.trim()).filter(Boolean);
            tags.forEach(tag => addTag(tag));
            input.value = '';
        }
    });

    // Novel Tags Input
    document.getElementById('novel-tags-input')?.addEventListener('keydown', (e) => {
        if (e.key === 'Enter') {
            e.preventDefault();
            const input = e.target;
            const tags = input.value.split(',').map(t => t.trim()).filter(Boolean);
            tags.forEach(tag => addNovelTag(tag));
            input.value = '';
        }
    });

    // Add Manga Form
    document.getElementById('add-manga-form')?.addEventListener('submit', async (e) => {
        e.preventDefault();

        const mangaData = {
            title: document.getElementById('manga-title').value,
            description: document.getElementById('manga-description').value,
            author_id: document.getElementById('manga-author').value,
            cover_image: document.getElementById('manga-cover-url').value,
            year: document.getElementById('manga-year').value,
            status: document.getElementById('manga-status').value,
            tags: tagList,
            categories: Array.from(document.getElementById('manga-categories').selectedOptions).map(o => o.value)
        };

        try {
            await apiFetch('/mangas', {
                method: 'POST',
                body: JSON.stringify(mangaData)
            });
            showToast('تمت إضافة المانجا بنجاح', 'success');
            resetAddMangaForm();
            switchSection('mangas');
        } catch (error) {
            showToast(error.message || 'فشل الإضافة', 'error');
        }
    });

    // Add Novel Form
    document.getElementById('add-novel-form')?.addEventListener('submit', async (e) => {
        e.preventDefault();

        const novelData = {
            title: document.getElementById('novel-title').value,
            description: document.getElementById('novel-description').value,
            author_id: document.getElementById('novel-author').value,
            cover_image: document.getElementById('novel-cover-url').value,
            status: document.getElementById('novel-status').value,
            tags: novelTagList,
            categories: Array.from(document.getElementById('novel-categories').selectedOptions).map(o => o.value)
        };

        try {
            await apiFetch('/novels', {
                method: 'POST',
                body: JSON.stringify(novelData)
            });
            showToast('تمت إضافة الرواية بنجاح', 'success');
            resetAddNovelForm();
            switchSection('novels');
        } catch (error) {
            showToast(error.message || 'فشل الإضافة', 'error');
        }
    });

    // Edit Manga Form
    document.getElementById('edit-manga-form')?.addEventListener('submit', async (e) => {
        e.preventDefault();

        const mangaData = {
            title: document.getElementById('edit-manga-title').value,
            description: document.getElementById('edit-manga-description').value,
            year: document.getElementById('edit-manga-year').value,
            status: document.getElementById('edit-manga-status').value
        };

        try {
            await apiFetch(`/mangas/${encodeURIComponent(currentEditMangaId)}`, {
                method: 'PUT',
                body: JSON.stringify(mangaData)
            });
            showToast('تم تحديث المانجا بنجاح', 'success');
            document.getElementById('editMangaModal').classList.remove('active');
            loadMangasList();
        } catch (error) {
            showToast(error.message || 'فشل التحديث', 'error');
        }
    });

    // Delete Manga Permanent
    document.getElementById('delete-manga-permanent-btn')?.addEventListener('click', async () => {
        if (!confirm('هل أنت متأكد من الحذف النهائي؟')) return;
        try {
            await deleteManga(currentEditMangaId);
            document.getElementById('editMangaModal').classList.remove('active');
        } catch (error) {
            showToast(error.message, 'error');
        }
    });

    // Close Modals
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', () => {
            btn.closest('.modal').classList.remove('active');
        });
    });

    document.querySelectorAll('.modal').forEach(modal => {
        modal.addEventListener('click', (e) => {
            if (e.target === modal) {
                modal.classList.remove('active');
            }
        });
    });

    // Chapter Manga Select
    document.getElementById('chapter-manga-id')?.addEventListener('change', (e) => {
        loadChaptersForManga(e.target.value);
    });

    // Add Chapter Form
    document.getElementById('add-chapter-form')?.addEventListener('submit', async (e) => {
        e.preventDefault();

        const mangaId = document.getElementById('chapter-manga-id').value;
        if (!mangaId) {
            showToast('يرجى اختيار مانجا', 'warning');
            return;
        }

        const chapterData = {
            number: parseInt(document.getElementById('chapter-number').value),
            title: document.getElementById('chapter-title').value,
            pages: document.getElementById('chapter-urls').value.split('\n').filter(u => u.trim())
        };

        try {
            await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`, {
                method: 'POST',
                body: JSON.stringify(chapterData)
            });
            showToast('تمت إضافة الفصل بنجاح', 'success');
            document.getElementById('add-chapter-form').reset();
            loadChaptersForManga(mangaId);
        } catch (error) {
            showToast(error.message || 'فشل الإضافة', 'error');
        }
    });

    // Logout
    document.getElementById('logout-button')?.addEventListener('click', async () => {
        try {
            await apiFetch('/auth/logout', { method: 'POST' });
        } catch {}
        window.location.href = 'dashboard.html';
    });

    // File Upload Area
    document.getElementById('chapter-upload-area')?.addEventListener('click', () => {
        document.getElementById('chapter-image-files').click();
    });

    document.getElementById('chapter-image-files')?.addEventListener('change', (e) => {
        const files = Array.from(e.target.files);
        uploadedFiles = files;
        // Here you would typically upload files and get URLs
        showToast(`تم اختيار ${files.length} ملف`, 'info');
    });

    // Bulk Upload
    document.getElementById('bulk-upload-area')?.addEventListener('click', () => {
        document.getElementById('bulk-file-input').click();
    });

    document.getElementById('bulk-file-input')?.addEventListener('change', handleBulkUpload);

    // Initial load
    loadDashboard();

    // Load users only when users section is active
    const usersSection = document.getElementById('section-users');
    if (usersSection && usersSection.classList.contains('active')) {
        loadUsers();
    }

    // Load users when switching to users section
    document.querySelectorAll('.admin-nav-item').forEach(item => {
        item.addEventListener('click', () => {
            if (item.dataset.section === 'users') {
                setTimeout(() => loadUsers(), 100);
            }
        });
    });
});

// ========== Users Management ==========
let usersCache = null;
let usersCacheTime = 0;
const CACHE_DURATION = 30000; // 30 seconds cache

async function loadUsers(forceRefresh = false) {
    const container = document.getElementById('users-list-container');
    if (!container) return;

    // Use cache if available and not expired
    if (!forceRefresh && usersCache && (Date.now() - usersCacheTime) < CACHE_DURATION) {
        renderUsersTable(usersCache);
        return;
    }

    try {
        const data = await apiFetch('/admin/users?page=1&limit=50');
        usersCache = data.users || [];
        usersCacheTime = Date.now();
        renderUsersTable(usersCache);
    } catch (error) {
        handleUsersError(error, container);
    }
}

function renderUsersTable(users) {
    const container = document.getElementById('users-list-container');
    if (!container) return;

    if (users.length === 0) {
        container.innerHTML = '<div class="empty-state"><p>لا يوجد مستخدمين</p></div>';
        return;
    }

    container.innerHTML = `
        <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 1rem;">
            <button class="btn btn-sm btn-secondary" onclick="loadUsers(true)" title="تحديث">
                <i class="fas fa-sync"></i> تحديث
            </button>
        </div>
        <table class="admin-table">
                <thead>
                    <tr>
                        <th>المستخدم</th>
                        <th>الرتبة</th>
                        <th>الحالة</th>
                        <th>تاريخ التسجيل</th>
                        <th>الإجراءات</th>
                    </tr>
                </thead>
                <tbody>
                    ${users.map(user => `
                        <tr>
                            <td>
                                <div style="display: flex; align-items: center; gap: 0.5rem;">
                                    <div class="user-avatar" style="width: 36px; height: 36px; border-radius: 50%; background: var(--admin-primary); display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; font-size: 0.875rem;">
                                        ${(user.username || 'U').charAt(0).toUpperCase()}
                                    </div>
                                    <div>
                                        <div style="font-weight: 600;">${escapeHtml(user.username || 'مستخدم')}</div>
                                        <div style="font-size: 0.75rem; color: var(--admin-text-muted);">ID: ${user.id}</div>
                                    </div>
                                </div>
                            </td>
                            <td>
                                <span class="role-badge ${user.roles?.includes('admin') ? 'role-admin' : 'role-user'}">
                                    ${user.roles?.includes('admin') ? '👑 مشرف' : '👤 مستخدم'}
                                </span>
                            </td>
                            <td>
                                <span class="status-badge status-${user.is_active !== false ? 'published' : 'draft'}">
                                    ${user.is_active !== false ? 'نشط' : 'غير نشط'}
                                </span>
                            </td>
                            <td style="direction: ltr; text-align: right;">
                                ${new Date(user.created_at).toLocaleDateString('ar-SA')}
                            </td>
                            <td>
                                <div style="display: flex; gap: 0.5rem;">
                                    ${user.roles?.includes('admin') ? `
                                        <button class="btn btn-sm btn-warning" onclick="demoteUser('${user.id}')" title="تخفيض الرتبة">
                                            <i class="fas fa-user-minus"></i>
                                        </button>
                                    ` : `
                                        <button class="btn btn-sm btn-success" onclick="promoteUser('${user.id}')" title="ترقية لمشرف">
                                            <i class="fas fa-user-plus"></i>
                                        </button>
                                    `}
                                    <button class="btn btn-sm btn-info" onclick="changeUserPassword('${user.id}', '${escapeHtml(user.username || 'مستخدم')}')" title="تغيير كلمة المرور">
                                        <i class="fas fa-key"></i>
                                    </button>
                                    <button class="btn btn-sm btn-danger" onclick="deleteUser('${user.id}')" title="حذف المستخدم">
                                        <i class="fas fa-trash"></i>
                                    </button>
                                </div>
                            </td>
                        </tr>
                    `).join('')}
                </tbody>
            </table>
    `;

    // Add CSS for admin table
        if (!document.getElementById('admin-table-styles')) {
            const style = document.createElement('style');
            style.id = 'admin-table-styles';
            style.textContent = `
                .admin-table {
                    width: 100%;
                    border-collapse: collapse;
                    background: rgba(255, 255, 255, 0.05);
                    border-radius: 12px;
                    overflow: hidden;
                }
                .admin-table th, .admin-table td {
                    padding: 1rem;
                    text-align: right;
                    border-bottom: 1px solid rgba(255, 255, 255, 0.1);
                }
                .admin-table th {
                    background: rgba(124, 58, 237, 0.2);
                    font-weight: 600;
                    color: var(--admin-text);
                }
                .admin-table tr:hover {
                    background: rgba(124, 58, 237, 0.1);
                }
                .role-badge {
                    display: inline-flex;
                    align-items: center;
                    gap: 0.25rem;
                    padding: 0.25rem 0.75rem;
                    border-radius: 20px;
                    font-size: 0.875rem;
                    font-weight: 500;
                }
                .role-admin {
                    background: rgba(234, 179, 8, 0.2);
                    color: #fbbf24;
                }
                .role-user {
                    background: rgba(156, 163, 175, 0.2);
                    color: #9ca3af;
                }
            `;
            document.head.appendChild(style);
        }
}

function handleUsersError(error, container) {
    console.error('Failed to load users:', error);
    if (error.message.includes('401') || error.message.includes('unauthorized')) {
        container.innerHTML = '<div class="error-state"><p>⚠️ يجب تسجيل الدخول كـ مشرف لعرض المستخدمين</p><button class="btn btn-primary" onclick="window.location.href=\'dashboard.html\'" style="margin-top:1rem;">تسجيل الدخول</button></div>';
    } else if (error.message.includes('429') || error.message.includes('too many requests')) {
        container.innerHTML = '<div class="error-state"><p>⚠️ تم إرسال طلبات كثيرة، يرجى الانتظار 30 ثانية</p><button class="btn btn-primary" onclick="loadUsers(true)" style="margin-top:1rem;">إعادة المحاولة</button></div>';
    } else {
        container.innerHTML = '<div class="error-state"><p>فشل تحميل المستخدمين: ' + escapeHtml(error.message) + '</p></div>';
    }
}

async function promoteUser(userId) {
    if (!confirm('هل أنت متأكد من ترقية هذا المستخدم إلى مشرف؟')) return;

    try {
        await apiFetch(`/admin/users/${encodeURIComponent(userId)}/promote`, { method: 'PUT' });
        showToast('تم ترقية المستخدم إلى مشرف', 'success');
        loadUsers();
    } catch (error) {
        showToast(error.message || 'فشل الترقية', 'error');
    }
}

async function demoteUser(userId) {
    if (!confirm('هل أنت متأكد من تخفيض رتبة هذا المشرف إلى مستخدم عادي؟')) return;

    try {
        await apiFetch(`/admin/users/${encodeURIComponent(userId)}/demote`, { method: 'PUT' });
        showToast('تم تخفيض رتبة المستخدم', 'success');
        loadUsers();
    } catch (error) {
        showToast(error.message || 'فشل التخفيض', 'error');
    }
}

async function deleteUser(userId) {
    if (!confirm('هل أنت متأكد من حذف هذا المستخدم؟ هذا الإجراء لا يمكن التراجع عنه.')) return;

    try {
        await apiFetch(`/admin/users/${encodeURIComponent(userId)}`, { method: 'DELETE' });
        showToast('تم حذف المستخدم', 'success');
        loadUsers();
    } catch (error) {
        showToast(error.message || 'فشل الحذف', 'error');
    }
}

async function changeUserPassword(userId, username) {
    const newPassword = prompt(`تغيير كلمة المرور للمستخدم: ${username}\n\nأدخل كلمة المرور الجديدة:`);
    if (!newPassword) return;
    if (newPassword.length < 6) {
        showToast('كلمة المرور يجب أن تكون 6 أحرف على الأقل', 'error');
        return;
    }

    if (!confirm('هل أنت متأكد من تغيير كلمة المرور لهذا المستخدم؟')) return;

    try {
        await apiFetch(`/admin/users/${encodeURIComponent(userId)}/password`, {
            method: 'PUT',
            body: JSON.stringify({ password: newPassword })
        });
        showToast('تم تغيير كلمة المرور بنجاح', 'success');
    } catch (error) {
        showToast(error.message || 'فشل تغيير كلمة المرور', 'error');
    }
}

// ========== Bulk Upload ==========
async function handleBulkUpload(e) {
    const file = e.target.files[0];
    if (!file) return;

    if (!file.name.endsWith('.json')) {
        showToast('يرجى اختيار ملف JSON', 'error');
        return;
    }

    try {
        const content = await file.text();
        const data = JSON.parse(content);

        if (!data.mangas || !Array.isArray(data.mangas)) {
            showToast('ملف غير صالح', 'error');
            return;
        }

        const container = document.getElementById('bulk-progress');
        const fill = document.getElementById('bulk-progress-fill');
        const status = document.getElementById('bulk-status');

        container.style.display = 'block';

        for (let i = 0; i < data.mangas.length; i++) {
            const manga = data.mangas[i];
            const progress = ((i + 1) / data.mangas.length) * 100;
            fill.style.width = `${progress}%`;
            status.textContent = `جاري استيراد ${i + 1} من ${data.mangas.length}...`;

            try {
                await apiFetch('/mangas', {
                    method: 'POST',
                    body: JSON.stringify(manga)
                });
            } catch (error) {
                console.error(`Failed to import manga ${manga.title}:`, error);
            }
        }

        status.textContent = 'تم الاستيراد بنجاح!';
        showToast(`تم استيراد ${data.mangas.length} مانجا`, 'success');

        setTimeout(() => {
            container.style.display = 'none';
        }, 3000);
    } catch (error) {
        showToast('فشل قراءة الملف', 'error');
    }
}

// Make functions globally available
window.switchSection = switchSection;
window.exportData = exportData;
window.resetAddMangaForm = resetAddMangaForm;
window.resetAddNovelForm = resetAddNovelForm;
window.openEditMangaModal = openEditMangaModal;
window.deleteManga = deleteManga;
window.deleteChapter = deleteChapter;
window.goToPage = goToPage;
window.removeNovelTag = removeNovelTag;
window.promoteUser = promoteUser;
window.demoteUser = demoteUser;
window.deleteUser = deleteUser;
window.loadUsers = loadUsers;
window.changeUserPassword = changeUserPassword;
