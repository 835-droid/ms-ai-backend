// admin-enhanced.js - لوحة إدارة متطورة بالكامل (نسخة محسنة)
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
        const chaptersCount = manga.chapters_count || 0;
        const lastUpdate = manga.updated_at ? new Date(manga.updated_at).toLocaleDateString('ar-EG') : '';
        return `
            <div class="manga-admin-item" data-manga-id="${escapeHtml(id)}">
                <div class="manga-admin-info">
                    <input type="checkbox" class="manga-select-checkbox" data-id="${escapeHtml(id)}" ${isSelected ? 'checked' : ''}>
                    <img class="manga-admin-cover" src="${escapeHtml(manga.cover_image || 'https://via.placeholder.com/50x70?text=No+Image')}" onerror="this.src='https://via.placeholder.com/50x70?text=No+Image'">
                    <div>
                        <div class="manga-admin-title">${escapeHtml(manga.title)}</div>
                        <div class="manga-admin-meta">
                            ✍️ ${manga.author_name ? escapeHtml(manga.author_name) : 'غير معروف'} 
                            📅 ${manga.year || 'غير محدد'}
                            📚 ${chaptersCount} فصل
                            🕒 ${lastUpdate || 'جديد'}
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
    
    document.querySelectorAll('.manga-select-checkbox').forEach(cb => {
        cb.addEventListener('change', (e) => {
            const id = e.target.dataset.id;
            if (e.target.checked) selectedMangas.add(id);
            else selectedMangas.delete(id);
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
    if (totalPages <= 1) { container.innerHTML = ''; return; }
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
    if (!confirm(`⚠️ تحذير: أنت على وشك حذف ${selectedMangas.size} مانجا بشكل دائم. هل أنت متأكد؟`)) return;
    const ids = Array.from(selectedMangas);
    let successCount = 0;
    for (const id of ids) {
        try {
            await apiFetch(`/mangas/${encodeURIComponent(id)}`, { method: 'DELETE' });
            successCount++;
        } catch (e) { console.error(e); }
    }
    showToast(`تم حذف ${successCount} من ${ids.length} مانجا`, successCount === ids.length ? 'success' : 'warning');
    selectedMangas.clear();
    await loadMangasList();
    await loadMangaOptions();
    await loadStats();
}

// ========== إضافة مانجا جديدة ==========
async function handleCreateManga(event) {
    event.preventDefault();
    const title = document.getElementById('manga-title')?.value.trim();
    const description = document.getElementById('manga-description')?.value.trim();
    const coverImage = document.getElementById('manga-cover-url')?.value.trim() || '';
    if (!title || !description) {
        showToast('عنوان المانجا والوصف مطلوبان', 'error');
        return;
    }
    const button = document.getElementById('add-manga-button');
    if (button) button.disabled = true;
    try {
        await apiFetch('/mangas', {
            method: 'POST',
            body: JSON.stringify({
                title, description, cover_image: coverImage,
                tags: tagList,
                year: document.getElementById('manga-year')?.value ? parseInt(document.getElementById('manga-year').value) : null,
                status: document.getElementById('manga-status')?.value,
                categories: Array.from(document.getElementById('manga-categories')?.selectedOptions || []).map(opt => opt.value),
                author_id: document.getElementById('manga-author')?.value || null,
                gallery: splitLines(document.getElementById('manga-gallery')?.value),
                keywords: document.getElementById('manga-keywords')?.value
            })
        });
        showToast('تمت إضافة المانجا بنجاح', 'success');
        event.target.reset();
        tagList = [];
        renderTags();
        await loadMangasList();
        await loadMangaOptions();
        await loadStats();
        document.querySelector('[data-tab="mangas"]').click();
    } catch (error) {
        showToast(error.message, 'error');
    } finally {
        if (button) button.disabled = false;
    }
}

// ========== تعديل المانجا ==========
async function openEditMangaModal(mangaId = null) {
    const id = mangaId || document.getElementById('chapter-manga-id')?.value;
    if (!id) { showToast('اختر مانجا أولاً', 'error'); return; }
    try {
        const manga = await apiFetch(`/mangas/${encodeURIComponent(id)}`);
        document.getElementById('edit-manga-title').value = manga.title;
        document.getElementById('edit-manga-description').value = manga.description || '';
        document.getElementById('edit-manga-cover').value = manga.cover_image || '';
        document.getElementById('edit-manga-year').value = manga.year || '';
        document.getElementById('edit-manga-status').value = manga.status || 'published';
        document.getElementById('edit-manga-gallery').value = (manga.gallery || []).join('\n');
        editTagList = [...(manga.tags || [])];
        renderEditTags();
        // تحديث معاينة الغلاف
        const previewDiv = document.getElementById('edit-cover-preview');
        if (previewDiv) previewDiv.innerHTML = manga.cover_image ? `<img src="${escapeHtml(manga.cover_image)}" style="max-width:100px">` : '';
        await loadAuthorsToSelect(document.getElementById('edit-manga-author'), manga.author_id);
        const categoriesSelect = document.getElementById('edit-manga-categories');
        if (categoriesSelect && manga.categories) {
            Array.from(categoriesSelect.options).forEach(opt => {
                opt.selected = manga.categories.includes(opt.value);
            });
        }
        document.getElementById('editMangaModal').style.display = 'flex';
        currentAdminMangaId = id;
    } catch (error) {
        showToast('فشل تحميل بيانات المانجا: ' + error.message, 'error');
    }
}

async function handleEditMangaSubmit(e) {
    e.preventDefault();
    const mangaId = currentAdminMangaId;
    const title = document.getElementById('edit-manga-title').value.trim();
    const description = document.getElementById('edit-manga-description').value.trim();
    const cover_image = document.getElementById('edit-manga-cover').value.trim();
    if (!title) { showToast('عنوان المانجا مطلوب', 'error'); return; }
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`, {
            method: 'PUT',
            body: JSON.stringify({
                title, description, cover_image,
                tags: editTagList,
                author_id: document.getElementById('edit-manga-author')?.value,
                year: parseInt(document.getElementById('edit-manga-year')?.value) || null,
                status: document.getElementById('edit-manga-status')?.value,
                categories: Array.from(document.getElementById('edit-manga-categories')?.selectedOptions || []).map(opt => opt.value),
                gallery: splitLines(document.getElementById('edit-manga-gallery')?.value)
            })
        });
        showToast('تم تحديث المانجا', 'success');
        document.getElementById('editMangaModal').style.display = 'none';
        await loadMangasList();
        await loadMangaOptions();
        await loadStats();
    } catch (error) {
        showToast(error.message, 'error');
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
    if (authorsList.length === 0) { container.innerHTML = '<div class="empty-state">لا يوجد مؤلفون</div>'; return; }
    container.innerHTML = authorsList.map(author => `
        <div class="admin-chapter-item" data-author-id="${escapeHtml(author.id)}">
            <div>
                <strong>${escapeHtml(author.name)}</strong>
                ${author.bio ? `<p style="font-size:0.75rem;">${escapeHtml(author.bio.substring(0, 100))}</p>` : ''}
            </div>
            <div class="admin-chapter-actions">
                <button class="btn btn-sm btn-secondary edit-author" data-id="${escapeHtml(author.id)}"><i class="fas fa-edit"></i></button>
                <button class="btn btn-sm btn-danger delete-author" data-id="${escapeHtml(author.id)}"><i class="fas fa-trash"></i></button>
            </div>
        </div>
    `).join('');
    document.querySelectorAll('.edit-author').forEach(btn => btn.addEventListener('click', () => openAuthorModal(btn.dataset.id)));
    document.querySelectorAll('.delete-author').forEach(btn => btn.addEventListener('click', () => deleteAuthor(btn.dataset.id)));
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
    if (!name) { showToast('اسم المؤلف مطلوب', 'error'); return; }
    try {
        if (id) {
            await apiFetch(`/admin/authors/${encodeURIComponent(id)}`, { method: 'PUT', body: JSON.stringify({ name, bio, image }) });
            showToast('تم تحديث المؤلف', 'success');
        } else {
            await apiFetch('/admin/authors', { method: 'POST', body: JSON.stringify({ name, bio, image }) });
            showToast('تم إضافة المؤلف', 'success');
        }
        document.getElementById('authorModal').style.display = 'none';
        await loadAuthors();
    } catch (error) { showToast(error.message, 'error'); }
}

async function deleteAuthor(authorId) {
    if (!confirm('هل أنت متأكد من حذف هذا المؤلف؟')) return;
    try {
        await apiFetch(`/admin/authors/${encodeURIComponent(authorId)}`, { method: 'DELETE' });
        showToast('تم حذف المؤلف', 'success');
        await loadAuthors();
    } catch (error) { showToast(error.message, 'error'); }
}

// ========== رفع الصور ومعاينتها ==========
function previewImagesBeforeUpload(files) {
    const container = document.getElementById('upload-preview');
    if (!container) return;
    container.innerHTML = '';
    Array.from(files).forEach((file, index) => {
        const reader = new FileReader();
        reader.onload = (e) => {
            const div = document.createElement('div');
            div.className = 'preview-item';
            div.innerHTML = `
                <img src="${e.target.result}" alt="preview">
                <button type="button" class="remove-preview" data-index="${index}">✖</button>
            `;
            container.appendChild(div);
        };
        reader.readAsDataURL(file);
    });
}

async function handleUploadChapterImages() {
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const input = document.getElementById('chapter-image-files');
    const files = Array.from(input?.files || []);
    if (!mangaId) { showToast('اختر مانجا أولاً', 'error'); return; }
    if (!files.length) { showToast('اختر صوراً لرفعها', 'error'); return; }
    const formData = new FormData();
    files.forEach(file => formData.append('images', file));
    const button = document.getElementById('upload-chapter-images-button');
    const progressDiv = document.getElementById('upload-progress');
    const progressBar = progressDiv?.querySelector('.progress-bar');
    const progressText = progressDiv?.querySelector('.progress-text');
    if (button) button.disabled = true;
    if (progressDiv) progressDiv.style.display = 'block';
    if (progressBar) progressBar.style.width = '30%';
    if (progressText) progressText.textContent = 'جاري رفع الصور...';
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/upload-images`, {
            method: 'POST',
            body: formData
        });
        if (progressBar) progressBar.style.width = '100%';
        if (progressText) progressText.textContent = 'اكتمل!';
        const uploaded = data.images || [];
        const textarea = document.getElementById('chapter-urls');
        const existing = splitLines(textarea.value);
        const merged = [...existing, ...uploaded];
        textarea.value = merged.join('\n');
        input.value = '';
        document.getElementById('upload-preview').innerHTML = '';
        showToast(`تم رفع ${uploaded.length} صورة`, 'success');
    } catch (error) {
        showToast(error.message, 'error');
        if (progressBar) progressBar.style.width = '0%';
        if (progressText) progressText.textContent = 'فشل الرفع';
    } finally {
        if (button) button.disabled = false;
        setTimeout(() => { if (progressDiv) progressDiv.style.display = 'none'; }, 2000);
    }
}

// ========== دوال الفصول ==========
async function loadMangaOptions() {
    const select = document.getElementById('chapter-manga-id');
    try {
        const data = await apiFetch('/mangas?limit=1000');
        const mangas = data.items || [];
        select.innerHTML = '<option value="">اختر المانجا</option>';
        mangas.forEach(manga => {
            const id = manga.id || manga._id;
            const option = document.createElement('option');
            option.value = id;
            option.textContent = `${manga.title} ${manga.year ? `(${manga.year})` : ''}`;
            select.appendChild(option);
        });
    } catch (error) { console.error(error); }
}

async function loadAdminChapterList(mangaId) {
    const box = document.getElementById('admin-chapter-list');
    if (!mangaId) { box.innerHTML = '<div class="empty-state">اختر مانجا لعرض الفصول</div>'; return; }
    box.innerHTML = '<div class="loading"><div class="loading-spinner"></div><p>جاري تحميل الفصول...</p></div>';
    try {
        const data = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`);
        const chapters = data.chapters || [];
        if (!chapters.length) { box.innerHTML = '<div class="empty-state">لا توجد فصول لهذه المانجا</div>'; return; }
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
            el.addEventListener('click', () => {
                const parent = el.closest('.admin-chapter-item');
                window.location.href = `${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(parent.dataset.chapterId)}`;
            });
        });
        document.querySelectorAll('.edit-chapter-btn').forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                await openEditChapterModal(mangaId, btn.dataset.chapterId);
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
        showToast('تم حذف الفصل', 'success');
    } catch (error) { showToast(error.message, 'error'); }
}

async function openEditChapterModal(mangaId, chapterId) {
    try {
        const chapter = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`);
        document.getElementById('edit-chapter-id').value = chapterId;
        document.getElementById('edit-chapter-title').value = chapter.title || '';
        document.getElementById('edit-chapter-number').value = chapter.number || '';
        document.getElementById('edit-chapter-urls').value = (chapter.pages || []).join('\n');
        document.getElementById('editChapterModal').style.display = 'flex';
    } catch (error) { showToast('فشل تحميل بيانات الفصل: ' + error.message, 'error'); }
}

async function handleEditChapterSubmit(e) {
    e.preventDefault();
    const chapterId = document.getElementById('edit-chapter-id').value;
    const mangaId = document.getElementById('chapter-manga-id').value;
    const title = document.getElementById('edit-chapter-title').value.trim();
    const number = parseInt(document.getElementById('edit-chapter-number').value, 10);
    const images = splitLines(document.getElementById('edit-chapter-urls').value);
    if (!title || !number || !images.length) {
        showToast('جميع الحقول مطلوبة ويلزم وجود صورة واحدة على الأقل', 'error');
        return;
    }
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`, {
            method: 'PUT',
            body: JSON.stringify({ title, number, images })
        });
        showToast('تم تحديث الفصل', 'success');
        document.getElementById('editChapterModal').style.display = 'none';
        await loadAdminChapterList(mangaId);
    } catch (error) { showToast(error.message, 'error'); }
}

async function handleCreateChapter(event) {
    event.preventDefault();
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const title = document.getElementById('chapter-title')?.value.trim();
    const number = Number(document.getElementById('chapter-number')?.value);
    const images = splitLines(document.getElementById('chapter-urls')?.value);
    if (!mangaId) return showToast('اختر مانجا أولاً', 'error');
    if (!number) return showToast('رقم الفصل مطلوب', 'error');
    if (!images.length) return showToast('أدخل روابط الصور', 'error');
    const button = document.getElementById('add-chapter-button');
    if (button) button.disabled = true;
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`, {
            method: 'POST',
            body: JSON.stringify({ title, number, images })
        });
        showToast('تمت إضافة الفصل', 'success');
        event.target.reset();
        document.getElementById('chapter-urls').value = '';
        document.getElementById('upload-preview').innerHTML = '';
        await loadAdminChapterList(mangaId);
        await loadStats();
    } catch (error) { showToast(error.message, 'error'); }
    finally { if (button) button.disabled = false; }
}

// ========== الرفع الجماعي والتصدير ==========
function setupBulkUpload() {
    const area = document.getElementById('bulk-upload-area');
    const fileInput = document.getElementById('bulk-file-input');
    if (!area) return;
    area.addEventListener('click', () => fileInput.click());
    area.addEventListener('dragover', (e) => { e.preventDefault(); area.style.borderColor = 'var(--admin-primary)'; });
    area.addEventListener('dragleave', () => area.style.borderColor = 'var(--admin-border)');
    area.addEventListener('drop', (e) => {
        e.preventDefault();
        area.style.borderColor = 'var(--admin-border)';
        if (e.dataTransfer.files.length) handleBulkFile(e.dataTransfer.files[0]);
    });
    fileInput.addEventListener('change', (e) => { if (e.target.files.length) handleBulkFile(e.target.files[0]); });
}

async function handleBulkFile(file) {
    if (!file.name.endsWith('.json')) { showToast('الرجاء اختيار ملف JSON', 'error'); return; }
    const reader = new FileReader();
    reader.onload = async (e) => {
        try {
            const data = JSON.parse(e.target.result);
            const mangas = data.mangas || data;
            if (!Array.isArray(mangas)) throw new Error('الملف يجب أن يحتوي على مصفوفة');
            const progressDiv = document.getElementById('bulk-progress');
            const progressFill = document.getElementById('bulk-progress-fill');
            const statusText = document.getElementById('bulk-status');
            progressDiv.style.display = 'block';
            let successCount = 0;
            for (let i = 0; i < mangas.length; i++) {
                const percent = ((i + 1) / mangas.length) * 100;
                progressFill.style.width = percent + '%';
                statusText.textContent = `جاري رفع ${i + 1} من ${mangas.length}: ${mangas[i].title}`;
                try {
                    await apiFetch('/mangas', { method: 'POST', body: JSON.stringify(mangas[i]) });
                    successCount++;
                } catch (err) { console.error(err); }
            }
            statusText.textContent = `اكتمل! تم رفع ${successCount} من ${mangas.length} مانجا`;
            setTimeout(() => progressDiv.style.display = 'none', 3000);
            await loadMangasList();
            await loadMangaOptions();
            await loadStats();
        } catch (err) { showToast('خطأ في قراءة الملف: ' + err.message, 'error'); }
    };
    reader.readAsText(file);
}

async function exportAllData() {
    try {
        const data = await apiFetch('/mangas?limit=10000');
        const mangas = data.items || [];
        const exportData = { exported_at: new Date().toISOString(), total: mangas.length, mangas };
        const blob = new Blob([JSON.stringify(exportData, null, 2)], { type: 'application/json' });
        const url = URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.href = url;
        a.download = `mangas_export_${new Date().toISOString().slice(0, 19)}.json`;
        a.click();
        URL.revokeObjectURL(url);
        showToast('تم تصدير البيانات بنجاح', 'success');
    } catch (error) { showToast(error.message, 'error'); }
}

// ========== معاينة الغلاف ==========
function initCoverPreview() {
    const coverInput = document.getElementById('manga-cover-url');
    const previewDiv = document.getElementById('cover-preview');
    if (!coverInput || !previewDiv) return;
    coverInput.addEventListener('input', () => {
        const url = coverInput.value.trim();
        if (url) previewDiv.innerHTML = `<img src="${escapeHtml(url)}" style="max-width:100px; border-radius:8px;" onerror="this.src='https://via.placeholder.com/100x140?text=Invalid'">`;
        else previewDiv.innerHTML = '';
    });
}

// ========== التهيئة ==========
async function initAdminPage() {
    if (!requireAuth()) return;
    try {
        const isAdmin = await ensureAdmin();
        if (!isAdmin) { window.location.href = webPagePath('control.html'); return; }
    } catch (err) {
        if (err.message === 'admin_check_network_error') showToast('فشل في التحقق من صلاحيات الأدمن', 'error');
        else window.location.href = webPagePath('control.html');
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
    
    // البحث والفلترة
    document.getElementById('search-manga')?.addEventListener('input', () => { currentPage = 1; renderMangasList(); });
    document.getElementById('filter-status')?.addEventListener('change', () => { currentPage = 1; renderMangasList(); });
    document.getElementById('filter-author')?.addEventListener('change', () => { currentPage = 1; renderMangasList(); });
    
    // العلامات
    const tagsInput = document.getElementById('manga-tags-input');
    if (tagsInput) {
        tagsInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') { e.preventDefault(); addTag(tagsInput.value); tagsInput.value = ''; }
        });
    }
    const editTagsInput = document.getElementById('edit-manga-tags-input');
    if (editTagsInput) {
        editTagsInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') { e.preventDefault(); addEditTag(editTagsInput.value); editTagsInput.value = ''; }
        });
    }
    
    // المودالات
    document.querySelectorAll('.close-modal').forEach(btn => {
        btn.addEventListener('click', () => {
            document.getElementById('editMangaModal').style.display = 'none';
            document.getElementById('editChapterModal').style.display = 'none';
            document.getElementById('authorModal').style.display = 'none';
        });
    });
    window.addEventListener('click', (e) => { if (e.target.classList.contains('modal')) e.target.style.display = 'none'; });
    
    // التبويبات الداخلية في مودال التعديل
    document.querySelectorAll('.modal-tab').forEach(tab => {
        tab.addEventListener('click', () => {
            const tabName = tab.dataset.editTab;
            document.querySelectorAll('.modal-tab').forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            document.querySelectorAll('.modal-tab-content').forEach(c => c.classList.remove('active'));
            document.getElementById(`edit-${tabName}`).classList.add('active');
        });
    });
    
    // ربط تبويب الفصول
    const chapterSelect = document.getElementById('chapter-manga-id');
    if (chapterSelect) chapterSelect.addEventListener('change', async () => await loadAdminChapterList(chapterSelect.value));
    document.getElementById('edit-manga-btn')?.addEventListener('click', () => openEditMangaModal());
    
    // ربط التبويبات الرئيسية
    document.querySelectorAll('.tab-button[data-tab]').forEach(btn => {
        btn.addEventListener('click', () => {
            const tabName = btn.dataset.tab;
            document.querySelectorAll('.tab-button').forEach(b => b.classList.remove('active'));
            document.querySelectorAll('.tab-content').forEach(c => c.classList.remove('active'));
            btn.classList.add('active');
            document.getElementById(`tab-${tabName}`).classList.add('active');
        });
    });
    
    // رفع الصور ومعاينة
    document.getElementById('select-files-button')?.addEventListener('click', () => document.getElementById('chapter-image-files')?.click());
    document.getElementById('chapter-image-files')?.addEventListener('change', (e) => previewImagesBeforeUpload(e.target.files));
    
    // تحميل البيانات
    await loadStats();
    await loadMangasList();
    await loadMangaOptions();
    try { await loadAuthors(); } catch(e) { console.warn(e); }
    setupBulkUpload();
    initCoverPreview();
    renderTags();
}

document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logout-button');
    if (logoutBtn) logoutBtn.addEventListener('click', async () => { try { await apiFetch('/auth/logout', { method: 'POST' }); } catch(e) {} logoutLocal(true); });
    initAdminPage();
});