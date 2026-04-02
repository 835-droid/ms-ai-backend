// admin-enhanced.js - لوحة إدارة متطورة بالكامل

let tagList = [];
let editTagList = [];
let currentAdminMangaId = '';
let allMangas = [];
let selectedMangas = new Set();
let currentPage = 1;
let itemsPerPage = 10;
let authorsList = [];

// ========== دوال العلامات ==========
function renderTags() {
    const container = document.getElementById('tags-container');
    if (!container) return;
    container.innerHTML = tagList.map(tag => `
        <span class="tag-item">
            ${escapeHtml(tag)}
            <button type="button" data-tag="${escapeHtml(tag)}">×</button>
        </span>
    `).join('');
    document.querySelectorAll('[data-tag]').forEach(btn => {
        btn.addEventListener('click', () => removeTag(btn.dataset.tag));
    });
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

function renderEditTags() {
    const container = document.getElementById('edit-tags-container');
    if (!container) return;
    container.innerHTML = editTagList.map(tag => `
        <span class="tag-item">
            ${escapeHtml(tag)}
            <button type="button" data-edit-tag="${escapeHtml(tag)}">×</button>
        </span>
    `).join('');
    document.querySelectorAll('[data-edit-tag]').forEach(btn => {
        btn.addEventListener('click', () => removeEditTag(btn.dataset.editTag));
    });
}

function addEditTag(tag) {
    const cleaned = String(tag || '').trim();
    if (!cleaned || editTagList.includes(cleaned)) return;
    editTagList.push(cleaned);
    renderEditTags();
}

function removeEditTag(tag) {
    editTagList = editTagList.filter(t => t !== tag);
    renderEditTags();
}

// ========== إحصائيات ==========
async function loadStats() {
    try {
        const data = await apiFetch('/admin/metrics');
        document.getElementById('stat-total-mangas').textContent = data.total_mangas || 0;
        document.getElementById('stat-total-chapters').textContent = data.total_chapters || 0;
        document.getElementById('stat-total-views').textContent = formatCompactNumber(data.total_views || 0);
        document.getElementById('stat-total-likes').textContent = formatCompactNumber(data.total_likes || 0);
    } catch (error) {
        console.error('Error loading stats:', error);
        // Silently fail - stats are optional
    }
}

// ========== إدارة المانجا (قائمة مع تحديد) ==========
async function loadMangasList() {
    const container = document.getElementById('manga-list-container');
    if (!container) return;
    
    container.innerHTML = '<div class="loading"><div class="loading-spinner"></div><p>جاري تحميل المانجا...</p></div>';
    
    try {
        const data = await apiFetch('/mangas?limit=1000');
        allMangas = data.items || [];
        renderMangasList();
        loadAuthorsForFilter();
    } catch (error) {
        container.innerHTML = `<div class="error">${escapeHtml(error.message)}</div>`;
    }
}

function renderMangasList() {
    const container = document.getElementById('manga-list-container');
    const searchTerm = document.getElementById('search-manga')?.value.toLowerCase() || '';
    const statusFilter = document.getElementById('filter-status')?.value || 'all';
    const authorFilter = document.getElementById('filter-author')?.value || 'all';
    
    let filtered = allMangas.filter(manga => {
        const matchesSearch = manga.title.toLowerCase().includes(searchTerm) ||
                             (manga.description || '').toLowerCase().includes(searchTerm);
        const matchesStatus = statusFilter === 'all' || manga.status === statusFilter;
        const matchesAuthor = authorFilter === 'all' || manga.author_id === authorFilter;
        return matchesSearch && matchesStatus && matchesAuthor;
    });
    
    // Pagination
    const totalPages = Math.ceil(filtered.length / itemsPerPage);
    const start = (currentPage - 1) * itemsPerPage;
    const paginated = filtered.slice(start, start + itemsPerPage);
    
    if (paginated.length === 0) {
        container.innerHTML = '<div class="empty-state">لا توجد مانجا</div>';
        renderPagination(totalPages);
        return;
    }
    
    container.innerHTML = paginated.map(manga => {
        const id = manga.id || manga._id;
        const isSelected = selectedMangas.has(id);
        return `
            <div class="manga-admin-item" data-manga-id="${escapeHtml(id)}">
                <div class="manga-admin-info">
                    <input type="checkbox" class="manga-select-checkbox" data-id="${escapeHtml(id)}" ${isSelected ? 'checked' : ''}>
                    <img class="manga-admin-cover" src="${escapeHtml(manga.cover_image || 'https://via.placeholder.com/50x70?text=No+Image')}" onerror="this.src='https://via.placeholder.com/50x70?text=No+Image'">
                    <div>
                        <div class="manga-admin-title">${escapeHtml(manga.title)}</div>
                        <div class="manga-admin-meta">
                            ${manga.author_name ? `✍️ ${escapeHtml(manga.author_name)}` : ''} 
                            ${manga.year ? `📅 ${manga.year}` : ''}
                            <span class="status-badge status-${manga.status || 'published'}">${manga.status || 'منشور'}</span>
                        </div>
                    </div>
                </div>
                <div class="manga-admin-actions">
                    <button class="btn btn-sm btn-secondary edit-manga-quick" data-id="${escapeHtml(id)}" title="تعديل">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn btn-sm btn-danger delete-manga-quick" data-id="${escapeHtml(id)}" title="حذف">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </div>
        `;
    }).join('');
    
    // ربط الأحداث
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
    
    document.querySelectorAll('.edit-manga-quick').forEach(btn => {
        btn.addEventListener('click', () => openEditMangaModal(btn.dataset.id));
    });
    
    document.querySelectorAll('.delete-manga-quick').forEach(btn => {
        btn.addEventListener('click', () => deleteManga(btn.dataset.id));
    });
    
    renderPagination(totalPages);
}

function updateSelectedCount() {
    const count = selectedMangas.size;
    document.getElementById('selected-count').textContent = `${count} مانجا محددة`;
    const batchBtn = document.getElementById('batch-delete-btn');
    if (batchBtn) batchBtn.disabled = count === 0;
}

function renderPagination(totalPages) {
    const container = document.getElementById('manga-pagination');
    if (!container) return;
    
    if (totalPages <= 1) {
        container.innerHTML = '';
        return;
    }
    
    let html = '';
    for (let i = 1; i <= totalPages; i++) {
        html += `<button class="btn btn-sm ${i === currentPage ? 'btn-primary' : 'btn-secondary'}" data-page="${i}">${i}</button>`;
    }
    container.innerHTML = html;
    
    container.querySelectorAll('[data-page]').forEach(btn => {
        btn.addEventListener('click', () => {
            currentPage = parseInt(btn.dataset.page);
            renderMangasList();
        });
    });
}

async function deleteManga(mangaId) {
    if (!confirm('⚠️ تحذير: حذف المانجا سيؤدي إلى حذف جميع الفصول والبيانات المرتبطة بها. هل أنت متأكد؟')) return;
    
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`, { method: 'DELETE' });
        showToast('تم حذف المانجا بنجاح', 'success');
        selectedMangas.delete(mangaId);
        await loadMangasList();
        await loadMangaOptions();
        await loadStats();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function batchDeleteMangas() {
    if (selectedMangas.size === 0) return;
    if (!confirm(`⚠️ تحذير: أنت على وشك حذف ${selectedMangas.size} مانجا بشكل دائم. هذا الإجراء لا يمكن التراجع عنه. هل أنت متأكد؟`)) return;
    
    const ids = Array.from(selectedMangas);
    let successCount = 0;
    
    for (const id of ids) {
        try {
            await apiFetch(`/mangas/${encodeURIComponent(id)}`, { method: 'DELETE' });
            successCount++;
        } catch (e) {
            console.error(`Failed to delete ${id}:`, e);
        }
    }
    
    showToast(`تم حذف ${successCount} من ${ids.length} مانجا`, successCount === ids.length ? 'success' : 'warning');
    selectedMangas.clear();
    await loadMangasList();
    await loadMangaOptions();
    await loadStats();
}

// ========== إضافة مانجا جديدة (متطورة) ==========
async function handleCreateManga(event) {
    event.preventDefault();
    
    const title = document.getElementById('manga-title')?.value.trim();
    const description = document.getElementById('manga-description')?.value.trim();
    const coverImage = document.getElementById('manga-cover-url')?.value.trim() || '';
    
    if (!title) {
        showMessage('error', 'عنوان المانجا مطلوب', 'admin-message');
        return;
    }
    if (!description) {
        showMessage('error', 'الوصف مطلوب', 'admin-message');
        return;
    }
    
    const button = document.getElementById('add-manga-button');
    if (button) button.disabled = true;
    
    try {
        await apiFetch('/mangas', {
            method: 'POST',
            body: JSON.stringify({
                title,
                description,
                cover_image: coverImage,
                tags: tagList
            })
        });
        showMessage('success', 'تمت إضافة المانجا بنجاح', 'admin-message');
        event.target.reset();
        tagList = [];
        renderTags();
        await loadMangasList();
        await loadMangaOptions();
        await loadStats();
        
        // تفعيل تبويب المانجا
        document.querySelector('[data-tab="mangas"]').click();
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    } finally {
        if (button) button.disabled = false;
    }
}

// ========== تعديل المانجا ==========
async function openEditMangaModal(mangaId = null) {
    const id = mangaId || document.getElementById('chapter-manga-id')?.value;
    if (!id) {
        showMessage('error', 'اختر مانجا أولاً', 'admin-message');
        return;
    }
    
    try {
        const manga = await apiFetch(`/mangas/${encodeURIComponent(id)}`);
        document.getElementById('edit-manga-title').value = manga.title;
        document.getElementById('edit-manga-description').value = manga.description || '';
        document.getElementById('edit-manga-cover').value = manga.cover_image || '';
        document.getElementById('edit-manga-year').value = manga.year || '';
        document.getElementById('edit-manga-status').value = manga.status || 'published';
        
        editTagList = [...(manga.tags || [])];
        renderEditTags();
        
        // تعيين المؤلف
        const authorSelect = document.getElementById('edit-manga-author');
        if (authorSelect) {
            await loadAuthorsToSelect(authorSelect, manga.author_id);
        }
        
        // تعيين التصنيفات
        const categoriesSelect = document.getElementById('edit-manga-categories');
        if (categoriesSelect && manga.categories) {
            Array.from(categoriesSelect.options).forEach(opt => {
                opt.selected = manga.categories.includes(opt.value);
            });
        }
        
        document.getElementById('editMangaModal').style.display = 'flex';
        currentAdminMangaId = id;
    } catch (error) {
        showMessage('error', 'فشل تحميل بيانات المانجا: ' + error.message, 'admin-message');
    }
}

async function handleEditMangaSubmit(e) {
    e.preventDefault();
    const mangaId = currentAdminMangaId;
    const title = document.getElementById('edit-manga-title').value.trim();
    const description = document.getElementById('edit-manga-description').value.trim();
    const cover_image = document.getElementById('edit-manga-cover').value.trim();
    const authorId = document.getElementById('edit-manga-author')?.value;
    const year = document.getElementById('edit-manga-year')?.value;
    const status = document.getElementById('edit-manga-status')?.value;
    const categories = Array.from(document.getElementById('edit-manga-categories')?.selectedOptions || []).map(opt => opt.value);
    
    if (!title) {
        showMessage('error', 'عنوان المانجا مطلوب', 'admin-message');
        return;
    }
    
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`, {
            method: 'PUT',
            body: JSON.stringify({
                title,
                description,
                cover_image,
                tags: editTagList,
                author_id: authorId,
                year: year ? parseInt(year) : null,
                status,
                categories
            })
        });
        showMessage('success', 'تم تحديث المانجا', 'admin-message');
        document.getElementById('editMangaModal').style.display = 'none';
        await loadMangasList();
        await loadMangaOptions();
        await loadStats();
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    }
}

// ========== إدارة المؤلفين ==========
async function loadAuthors() {
    try {
        const data = await apiFetch('/admin/authors');
        authorsList = data.authors || [];
        renderAuthorsList();
        loadAuthorsToSelect(document.getElementById('manga-author'));
        loadAuthorsToSelect(document.getElementById('edit-manga-author'));
        loadAuthorsForFilter();
    } catch (error) {
        console.error('Error loading authors:', error);
        authorsList = [];
    }
}

function renderAuthorsList() {
    const container = document.getElementById('authors-list');
    if (!container) return;
    
    if (authorsList.length === 0) {
        container.innerHTML = '<div class="empty-state">لا يوجد مؤلفون</div>';
        return;
    }
    
    container.innerHTML = authorsList.map(author => `
        <div class="admin-chapter-item" data-author-id="${escapeHtml(author.id)}">
            <div>
                <strong>${escapeHtml(author.name)}</strong>
                ${author.bio ? `<p style="font-size:0.75rem; margin-top:0.25rem;">${escapeHtml(author.bio.substring(0, 100))}</p>` : ''}
            </div>
            <div class="admin-chapter-actions">
                <button class="btn btn-sm btn-secondary edit-author" data-id="${escapeHtml(author.id)}">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="btn btn-sm btn-danger delete-author" data-id="${escapeHtml(author.id)}">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        </div>
    `).join('');
    
    document.querySelectorAll('.edit-author').forEach(btn => {
        btn.addEventListener('click', () => openAuthorModal(btn.dataset.id));
    });
    document.querySelectorAll('.delete-author').forEach(btn => {
        btn.addEventListener('click', () => deleteAuthor(btn.dataset.id));
    });
}

function loadAuthorsToSelect(selectElement, selectedId = null) {
    if (!selectElement) return;
    selectElement.innerHTML = '<option value="">اختر المؤلف</option>';
    authorsList.forEach(author => {
        const option = document.createElement('option');
        option.value = author.id;
        option.textContent = author.name;
        if (selectedId && selectedId === author.id) option.selected = true;
        selectElement.appendChild(option);
    });
}

function loadAuthorsForFilter() {
    const select = document.getElementById('filter-author');
    if (!select) return;
    select.innerHTML = '<option value="all">جميع المؤلفين</option>';
    authorsList.forEach(author => {
        const option = document.createElement('option');
        option.value = author.id;
        option.textContent = author.name;
        select.appendChild(option);
    });
}

async function openAuthorModal(authorId = null) {
    const modal = document.getElementById('authorModal');
    const title = document.getElementById('author-modal-title');
    const idField = document.getElementById('author-id');
    const nameField = document.getElementById('author-name');
    const bioField = document.getElementById('author-bio');
    const imageField = document.getElementById('author-image');
    
    if (authorId) {
        const author = authorsList.find(a => a.id === authorId);
        if (author) {
            title.textContent = 'تعديل المؤلف';
            idField.value = author.id;
            nameField.value = author.name;
            bioField.value = author.bio || '';
            imageField.value = author.image || '';
        }
    } else {
        title.textContent = 'إضافة مؤلف جديد';
        idField.value = '';
        nameField.value = '';
        bioField.value = '';
        imageField.value = '';
    }
    
    modal.style.display = 'flex';
}

async function handleAuthorSubmit(e) {
    e.preventDefault();
    const id = document.getElementById('author-id').value;
    const name = document.getElementById('author-name').value.trim();
    const bio = document.getElementById('author-bio').value.trim();
    const image = document.getElementById('author-image').value.trim();
    
    if (!name) {
        showToast('اسم المؤلف مطلوب', 'error');
        return;
    }
    
    try {
        if (id) {
            await apiFetch(`/admin/authors/${encodeURIComponent(id)}`, {
                method: 'PUT',
                body: JSON.stringify({ name, bio, image })
            });
            showToast('تم تحديث المؤلف', 'success');
        } else {
            await apiFetch('/admin/authors', {
                method: 'POST',
                body: JSON.stringify({ name, bio, image })
            });
            showToast('تم إضافة المؤلف', 'success');
        }
        document.getElementById('authorModal').style.display = 'none';
        await loadAuthors();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

async function deleteAuthor(authorId) {
    if (!confirm('هل أنت متأكد من حذف هذا المؤلف؟')) return;
    
    try {
        await apiFetch(`/admin/authors/${encodeURIComponent(authorId)}`, { method: 'DELETE' });
        showToast('تم حذف المؤلف', 'success');
        await loadAuthors();
    } catch (error) {
        showToast(error.message, 'error');
    }
}

// ========== الرفع الجماعي والتصدير ==========
function setupBulkUpload() {
    const area = document.getElementById('bulk-upload-area');
    const fileInput = document.getElementById('bulk-file-input');
    
    if (!area) return;
    
    area.addEventListener('click', () => fileInput.click());
    area.addEventListener('dragover', (e) => {
        e.preventDefault();
        area.style.borderColor = 'var(--primary)';
    });
    area.addEventListener('dragleave', () => {
        area.style.borderColor = 'var(--border-light)';
    });
    area.addEventListener('drop', (e) => {
        e.preventDefault();
        area.style.borderColor = 'var(--border-light)';
        const files = e.dataTransfer.files;
        if (files.length) handleBulkFile(files[0]);
    });
    
    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length) handleBulkFile(e.target.files[0]);
    });
}

async function handleBulkFile(file) {
    if (!file.name.endsWith('.json')) {
        showToast('الرجاء اختيار ملف JSON', 'error');
        return;
    }
    
    const reader = new FileReader();
    reader.onload = async (e) => {
        try {
            const data = JSON.parse(e.target.result);
            const mangas = data.mangas || data;
            
            if (!Array.isArray(mangas)) {
                showToast('الملف يجب أن يحتوي على مصفوفة من المانجا', 'error');
                return;
            }
            
            const progressDiv = document.getElementById('bulk-progress');
            const progressFill = document.getElementById('bulk-progress-fill');
            const statusText = document.getElementById('bulk-status');
            
            progressDiv.style.display = 'block';
            let successCount = 0;
            
            for (let i = 0; i < mangas.length; i++) {
                const manga = mangas[i];
                const percent = ((i + 1) / mangas.length) * 100;
                progressFill.style.width = percent + '%';
                statusText.textContent = `جاري رفع ${i + 1} من ${mangas.length}: ${manga.title}`;
                
                try {
                    await apiFetch('/mangas', {
                        method: 'POST',
                        body: JSON.stringify(manga)
                    });
                    successCount++;
                } catch (err) {
                    console.error(`Failed to upload ${manga.title}:`, err);
                }
            }
            
            statusText.textContent = `اكتمل! تم رفع ${successCount} من ${mangas.length} مانجا`;
            setTimeout(() => {
                progressDiv.style.display = 'none';
            }, 3000);
            
            await loadMangasList();
            await loadMangaOptions();
            await loadStats();
            
        } catch (err) {
            showToast('خطأ في قراءة الملف: ' + err.message, 'error');
            document.getElementById('bulk-progress').style.display = 'none';
        }
    };
    reader.readAsText(file);
}

async function exportAllData() {
    try {
        const data = await apiFetch('/mangas?limit=10000');
        const mangas = data.items || [];
        
        const exportData = {
            exported_at: new Date().toISOString(),
            total: mangas.length,
            mangas: mangas
        };
        
        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `mangas_export_${new Date().toISOString().slice(0, 19)}.json`;
        a.click();
        URL.revokeObjectURL(url);
        
        showToast('تم تصدير البيانات بنجاح', 'success');
    } catch (error) {
        showToast(error.message, 'error');
    }
}

// ========== دوال الفصول (محسنة) ==========
async function loadMangaOptions() {
    const select = document.getElementById('chapter-manga-id');
    try {
        const data = await apiFetch('/mangas?limit=1000');
        const mangas = data.items || [];
        select.innerHTML = `<option value="">اختر المانجا</option>`;
        mangas.forEach(manga => {
            const id = manga.id || manga._id;
            const option = document.createElement('option');
            option.value = id;
            option.textContent = `${manga.title} ${manga.year ? `(${manga.year})` : ''}`;
            select.appendChild(option);
        });
    } catch (error) {
        console.error(error);
    }
}

async function loadAdminChapterList(mangaId) {
    const box = document.getElementById('admin-chapter-list');
    if (!mangaId) {
        box.innerHTML = `<div class="empty-state">اختر مانجا لعرض الفصول</div>`;
        return;
    }
    box.innerHTML = `<div class="loading"><div class="loading-spinner"></div><p>جاري تحميل الفصول...</p></div>`;
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        const chapters = data.chapters || [];
        if (!chapters.length) {
            box.innerHTML = `<div class="empty-state">لا توجد فصول لهذه المانجا</div>`;
            return;
        }
        box.innerHTML = chapters.map(chapter => {
            const chapterId = chapter.id || chapter._id;
            return `
                <div class="admin-chapter-item" data-chapter-id="${escapeHtml(chapterId)}" data-manga-id="${escapeHtml(mangaId)}">
                    <div class="chapter-info" style="flex:1; cursor:pointer;">
                        <strong>الفصل ${escapeHtml(chapter.number)}</strong>
                        <div>${escapeHtml(chapter.title || '')}</div>
                        <small>${Array.isArray(chapter.pages) ? chapter.pages.length : 0} صفحة</small>
                    </div>
                    <div class="admin-chapter-actions">
                        <button class="btn btn-secondary edit-chapter-btn" data-chapter-id="${escapeHtml(chapterId)}" data-manga-id="${escapeHtml(mangaId)}">✏️ تعديل</button>
                        <button class="btn btn-danger delete-chapter-btn" data-chapter-id="${escapeHtml(chapterId)}">🗑️ حذف</button>
                    </div>
                </div>
            `;
        }).join('');
        
        document.querySelectorAll('.chapter-info').forEach(el => {
            el.addEventListener('click', (e) => {
                const parent = el.closest('.admin-chapter-item');
                const chapterId = parent.dataset.chapterId;
                const mangaId = parent.dataset.mangaId;
                window.location.href = `${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}`;
            });
        });
        
        document.querySelectorAll('.edit-chapter-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                await openEditChapterModal(btn.dataset.mangaId, btn.dataset.chapterId);
            });
        });
        
        document.querySelectorAll('.delete-chapter-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                if (!confirm('حذف الفصل نهائياً؟')) return;
                await deleteChapter(mangaId, btn.dataset.chapterId);
            });
        });
    } catch (error) {
        box.innerHTML = `<div class="error">${escapeHtml(error.message)}</div>`;
    }
}

async function deleteChapter(mangaId, chapterId) {
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`, { method: 'DELETE' });
        await loadAdminChapterList(mangaId);
        showMessage('success', 'تم حذف الفصل', 'admin-message');
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    }
}

async function openEditChapterModal(mangaId, chapterId) {
    try {
        const chapter = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`);
        document.getElementById('edit-chapter-id').value = chapterId;
        document.getElementById('edit-chapter-title').value = chapter.title || '';
        document.getElementById('edit-chapter-number').value = chapter.number || '';
        const pagesText = (chapter.pages || []).join('\n');
        document.getElementById('edit-chapter-urls').value = pagesText;
        document.getElementById('editChapterModal').style.display = 'flex';
    } catch (error) {
        showMessage('error', 'فشل تحميل بيانات الفصل: ' + error.message, 'admin-message');
    }
}

async function handleEditChapterSubmit(e) {
    e.preventDefault();
    const chapterId = document.getElementById('edit-chapter-id').value;
    const mangaId = document.getElementById('chapter-manga-id').value;
    const title = document.getElementById('edit-chapter-title').value.trim();
    const number = parseInt(document.getElementById('edit-chapter-number').value, 10);
    const rawImages = document.getElementById('edit-chapter-urls').value;
    const images = splitLines(rawImages);
    
    if (!title || !number || !images.length) {
        showMessage('error', 'جميع الحقول مطلوبة ويلزم وجود صورة واحدة على الأقل', 'admin-message');
        return;
    }
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`, {
            method: 'PUT',
            body: JSON.stringify({ title, number, images })
        });
        showMessage('success', 'تم تحديث الفصل', 'admin-message');
        document.getElementById('editChapterModal').style.display = 'none';
        await loadAdminChapterList(mangaId);
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    }
}

async function handleCreateChapter(event) {
    event.preventDefault();
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const title = document.getElementById('chapter-title')?.value.trim();
    const number = Number(document.getElementById('chapter-number')?.value);
    const rawImages = document.getElementById('chapter-urls')?.value;
    const images = splitLines(rawImages);
    
    if (!mangaId) return showMessage('error', 'اختر مانجا أولاً', 'admin-message');
    if (!title) return showMessage('error', 'عنوان الفصل مطلوب', 'admin-message');
    if (!Number.isFinite(number) || number < 1) return showMessage('error', 'رقم الفصل غير صحيح', 'admin-message');
    if (!images.length) return showMessage('error', 'أدخل روابط الصور، كل رابط في سطر', 'admin-message');
    
    const button = document.getElementById('add-chapter-button');
    if (button) button.disabled = true;
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`, {
            method: 'POST',
            body: JSON.stringify({ title, number, images })
        });
        showMessage('success', 'تمت إضافة الفصل', 'admin-message');
        event.target.reset();
        await loadAdminChapterList(mangaId);
        await loadStats();
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    } finally {
        if (button) button.disabled = false;
    }
}

async function handleUploadChapterImages() {
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const input = document.getElementById('chapter-image-files');
    const files = Array.from(input?.files || []);
    
    if (!mangaId) {
        showMessage('error', 'اختر مانجا أولاً', 'admin-message');
        return;
    }
    if (!files.length) {
        showMessage('error', 'اختر صوراً لرفعها', 'admin-message');
        return;
    }
    
    const formData = new FormData();
    files.forEach(file => formData.append('images', file));
    
    const button = document.getElementById('upload-chapter-images-button');
    if (button) button.disabled = true;
    
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/upload-images`, {
            method: 'POST',
            body: formData
        });
        const uploaded = data.images || [];
        
        const textarea = document.getElementById('chapter-urls');
        const existing = splitLines(textarea.value);
        const merged = [...existing, ...uploaded];
        textarea.value = merged.join('\n');
        
        input.value = '';
        document.getElementById('selected-files-count').textContent = 'لم يتم اختيار أي ملف';
        showMessage('success', `تم رفع ${uploaded.length} صورة وإضافتها إلى الروابط`, 'admin-message');
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    } finally {
        if (button) button.disabled = false;
    }
}

// ========== معاينة الغلاف ==========
function initCoverPreview() {
    const coverInput = document.getElementById('manga-cover-url');
    const previewDiv = document.getElementById('cover-preview');
    if (!coverInput || !previewDiv) return;
    
    coverInput.addEventListener('input', () => {
        const url = coverInput.value.trim();
        if (url) {
            previewDiv.innerHTML = `<img src="${escapeHtml(url)}" style="max-width: 100px; max-height: 100px; border-radius: 8px;" onerror="this.src='https://via.placeholder.com/100x140?text=Invalid+URL'">`;
        } else {
            previewDiv.innerHTML = '';
        }
    });
}

// ========== التهيئة ==========
async function initAdminPage() {
    if (!requireAuth()) return;
    
    const isAdmin = await ensureAdmin();
    if (!isAdmin) {
        window.location.href = webPagePath('control.html');
        return;
    }
    
    // ربط الأحداث
    document.getElementById('add-manga-form')?.addEventListener('submit', handleCreateManga);
    document.getElementById('add-chapter-form')?.addEventListener('submit', handleCreateChapter);
    document.getElementById('upload-chapter-images-button')?.addEventListener('click', handleUploadChapterImages);
    document.getElementById('edit-manga-form')?.addEventListener('submit', handleEditMangaSubmit);
    document.getElementById('edit-chapter-form')?.addEventListener('submit', handleEditChapterSubmit);
    document.getElementById('batch-delete-btn')?.addEventListener('click', batchDeleteMangas);
    document.getElementById('select-all-mangas')?.addEventListener('change', (e) => {
        document.querySelectorAll('.manga-select-checkbox').forEach(cb => {
            cb.checked = e.target.checked;
            const id = cb.dataset.id;
            if (e.target.checked) selectedMangas.add(id);
            else selectedMangas.delete(id);
        });
        updateSelectedCount();
    });
    document.getElementById('export-data-btn')?.addEventListener('click', exportAllData);
    document.getElementById('add-author-btn')?.addEventListener('click', () => openAuthorModal());
    document.getElementById('author-form')?.addEventListener('submit', handleAuthorSubmit);
    document.getElementById('delete-manga-btn')?.addEventListener('click', () => {
        const mangaId = document.getElementById('chapter-manga-id')?.value;
        if (mangaId) deleteManga(mangaId);
    });
    document.getElementById('delete-manga-permanent-btn')?.addEventListener('click', () => {
        if (currentAdminMangaId) deleteManga(currentAdminMangaId);
    });
    document.getElementById('add-new-author-btn')?.addEventListener('click', () => openAuthorModal());
    
    // شريط البحث والفلترة
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
    
    // دوال العلامات
    const tagsInput = document.getElementById('manga-tags-input');
    if (tagsInput) {
        tagsInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addTag(tagsInput.value);
                tagsInput.value = '';
            }
        });
    }
    
    const editTagsInput = document.getElementById('edit-manga-tags-input');
    if (editTagsInput) {
        editTagsInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addEditTag(editTagsInput.value);
                editTagsInput.value = '';
            }
        });
    }
    
    // ربط المودالات
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', () => {
            document.getElementById('editMangaModal').style.display = 'none';
            document.getElementById('editChapterModal').style.display = 'none';
            document.getElementById('authorModal').style.display = 'none';
        });
    });
    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            e.target.style.display = 'none';
        }
    });
    
    // ربط تبويب الفصول
    const chapterSelect = document.getElementById('chapter-manga-id');
    if (chapterSelect) {
        chapterSelect.addEventListener('change', async () => {
            await loadAdminChapterList(chapterSelect.value);
        });
    }
    document.getElementById('edit-manga-btn')?.addEventListener('click', () => openEditMangaModal());
    
    // ربط التبويبات
    document.querySelectorAll('.tab-button[data-tab]').forEach(btn => {
        btn.addEventListener('click', () => {
            const tabName = btn.dataset.tab;
            // إزالة الكلاس النشط من جميع الأزرار والمحتويات
            document.querySelectorAll('.tab-button').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            // إضافة الكلاس النشط للزر والمحتوى المختار
            btn.classList.add('active');
            const tabElement = document.getElementById(`tab-${tabName}`);
            if (tabElement) {
                tabElement.classList.add('active');
            }
        });
    });
    
    // ربط زر اختيار الملفات
    document.getElementById('select-files-button')?.addEventListener('click', () => {
        document.getElementById('chapter-image-files')?.click();
    });
    
    // تحميل البيانات
    await loadStats();
    await loadMangasList();
    await loadMangaOptions();
    try {
        await loadAuthors();
    } catch (error) {
        console.warn('Authors endpoint not available, continuing without authors');
        authorsList = [];
    }
    setupBulkUpload();
    initCoverPreview();
    renderTags();
}

// تسجيل الخروج
document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logout-button');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            try { await apiFetch('/auth/logout', { method: 'POST' }); } catch(e) {}
            logoutLocal(true);
        });
    }
});

document.addEventListener('DOMContentLoaded', initAdminPage);