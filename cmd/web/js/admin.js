// admin.js - لوحة الإدارة الكاملة مع التعديلات والمودالات
// تم تحسين معالجة الأخطاء ورفع الصور

let tagList = [];
let currentAdminMangaId = '';
let editTagList = [];

function renderTags() {
    const container = document.getElementById('tags-container');
    if (!container) return;
    container.innerHTML = tagList.map(tag => `
        <span class="tag-item">
            ${escapeHtml(tag)}
            <button type="button" data-tag="${escapeHtml(tag)}">×</button>
        </span>
    `).join('');
    qsa('[data-tag]', container).forEach(btn => {
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
    qsa('[data-edit-tag]', container).forEach(btn => {
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

function getChapterImagesFromTextarea(textareaId = 'chapter-urls') {
    const raw = String(document.getElementById(textareaId)?.value || '');
    return raw
        .split(/\n|,/)
        .map(value => String(value || '').trim().replace(/^["']|["']$/g, ''))
        .filter(Boolean);
}

function appendChapterImagesToTextarea(urls, textareaId = 'chapter-urls') {
    const textarea = document.getElementById(textareaId);
    if (!textarea || !Array.isArray(urls) || !urls.length) return;
    const existing = getChapterImagesFromTextarea(textareaId);
    const merged = [...existing];
    urls.forEach((url) => {
        if (url && !merged.includes(url)) merged.push(url);
    });
    textarea.value = merged.join('\n');
}

async function handleUploadChapterImages() {
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const input = document.getElementById('chapter-image-files');
    const button = document.getElementById('upload-chapter-images-button');
    const files = Array.from(input?.files || []);
    const progressContainer = document.getElementById('upload-progress');
    const progressBar = progressContainer?.querySelector('.progress-bar');
    const progressText = progressContainer?.querySelector('.progress-text');

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

    if (button) {
        button.disabled = true;
        button.textContent = 'جاري الرفع...';
    }
    if (progressContainer) progressContainer.style.display = 'block';
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
        appendChapterImagesToTextarea(uploaded);
        if (input) input.value = '';
        const countSpan = document.getElementById('selected-files-count');
        if (countSpan) countSpan.textContent = 'لم يتم اختيار أي ملف';
        showMessage('success', `تم رفع ${uploaded.length} صورة وإضافتها إلى الروابط`, 'admin-message');
    } catch (error) {
        console.error('Upload error:', error);
        let errorMsg = error.message;
        if (errorMsg.includes('انتهت صلاحية الجلسة')) {
            errorMsg = 'انتهت صلاحية الجلسة. سيتم إعادة توجيهك لتسجيل الدخول.';
            showMessage('error', errorMsg, 'admin-message');
            setTimeout(() => logoutLocal(true), 2000);
        } else {
            showMessage('error', errorMsg, 'admin-message');
        }
        if (progressBar) progressBar.style.width = '0%';
        if (progressText) progressText.textContent = 'فشل الرفع';
    } finally {
        if (button) {
            button.disabled = false;
            button.textContent = 'رفع الصور وإضافة الروابط';
        }
        if (progressContainer) {
            setTimeout(() => {
                progressContainer.style.display = 'none';
                if (progressBar) progressBar.style.width = '0%';
                if (progressText) progressText.textContent = '';
            }, 2500);
        }
    }
}

async function loadMangaOptions() {
    const select = document.getElementById('chapter-manga-id');
    const listBox = document.getElementById('admin-chapter-list');
    try {
        const data = await apiFetch('/mangas?limit=100');
        const mangas = data.items || [];
        select.innerHTML = `<option value="">اختر المانجا</option>`;
        mangas.forEach(manga => {
            const id = manga.id || manga._id;
            const option = document.createElement('option');
            option.value = id;
            option.textContent = manga.title;
            select.appendChild(option);
        });
        if (!currentAdminMangaId && mangas.length) {
            currentAdminMangaId = mangas[0].id || mangas[0]._id;
            select.value = currentAdminMangaId;
            await loadAdminChapterList(currentAdminMangaId);
        } else if (currentAdminMangaId) {
            select.value = currentAdminMangaId;
            await loadAdminChapterList(currentAdminMangaId);
        } else {
            listBox.innerHTML = `<div class="empty-state">لا توجد مانجا بعد</div>`;
        }
    } catch (error) {
        listBox.innerHTML = `<div class="error">${escapeHtml(error.message)}</div>`;
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

        qsa('.chapter-info', box).forEach(el => {
            el.addEventListener('click', (e) => {
                const parent = el.closest('.admin-chapter-item');
                const chapterId = parent.dataset.chapterId;
                const mangaId = parent.dataset.mangaId;
                window.location.href = `${webPagePath('manga-reader.html')}?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}`;
            });
        });

        qsa('.edit-chapter-btn', box).forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                const chapterId = btn.dataset.chapterId;
                const mangaId = btn.dataset.mangaId;
                await openEditChapterModal(mangaId, chapterId);
            });
        });

        qsa('.delete-chapter-btn', box).forEach(btn => {
            btn.addEventListener('click', async (e) => {
                e.stopPropagation();
                if (!confirm('حذف الفصل نهائياً؟')) return;
                const chapterId = btn.dataset.chapterId;
                await deleteChapter(mangaId, chapterId);
            });
        });
    } catch (error) {
        box.innerHTML = `<div class="error">${escapeHtml(error.message)}</div>`;
    }
}

async function deleteChapter(mangaId, chapterId) {
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters/${encodeURIComponent(chapterId)}`, {
            method: 'DELETE'
        });
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
    const images = getChapterImagesFromTextarea('edit-chapter-urls');
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

async function openEditMangaModal() {
    const mangaId = document.getElementById('chapter-manga-id').value;
    if (!mangaId) {
        showMessage('error', 'اختر مانجا أولاً', 'admin-message');
        return;
    }
    try {
        const manga = await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`);
        document.getElementById('edit-manga-title').value = manga.title;
        document.getElementById('edit-manga-description').value = manga.description || '';
        document.getElementById('edit-manga-cover').value = manga.cover_image || '';
        editTagList = [...(manga.tags || [])];
        renderEditTags();
        document.getElementById('editMangaModal').style.display = 'flex';
        currentAdminMangaId = mangaId;
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
    if (!title) {
        showMessage('error', 'عنوان المانجا مطلوب', 'admin-message');
        return;
    }
    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}`, {
            method: 'PUT',
            body: JSON.stringify({ title, description, cover_image, tags: editTagList })
        });
        showMessage('success', 'تم تحديث المانجا', 'admin-message');
        document.getElementById('editMangaModal').style.display = 'none';
        await loadMangaOptions();
        if (document.getElementById('chapter-manga-id').value === mangaId) {
            await loadAdminChapterList(mangaId);
        }
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    }
}

async function handleCreateManga(event) {
    event.preventDefault();
    const title = document.getElementById('manga-title')?.value.trim();
    const description = document.getElementById('manga-description')?.value.trim();
    const coverImage = document.getElementById('manga-cover-url')?.value.trim() || '';
    if (!title) {
        showMessage('error', 'عنوان المانجا مطلوب', 'admin-message');
        return;
    }
    const button = document.getElementById('add-manga-button');
    if (button) button.disabled = true;
    try {
        await apiFetch('/mangas', {
            method: 'POST',
            body: JSON.stringify({ title, description, cover_image: coverImage, tags: tagList })
        });
        showMessage('success', 'تمت إضافة المانجا', 'admin-message');
        event.target.reset();
        tagList = [];
        renderTags();
        await loadMangaOptions();
    } catch (error) {
        let errorMsg = error.message;
        if (errorMsg.includes('انتهت صلاحية الجلسة')) {
            errorMsg = 'انتهت صلاحية الجلسة. سيتم إعادة توجيهك لتسجيل الدخول.';
            showMessage('error', errorMsg, 'admin-message');
            setTimeout(() => logoutLocal(true), 2000);
        } else {
            showMessage('error', errorMsg, 'admin-message');
        }
    } finally {
        if (button) button.disabled = false;
    }
}

async function handleCreateChapter(event) {
    event.preventDefault();
    const mangaId = document.getElementById('chapter-manga-id')?.value;
    const title = document.getElementById('chapter-title')?.value.trim();
    const number = Number(document.getElementById('chapter-number')?.value);
    const images = getChapterImagesFromTextarea();
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
    } catch (error) {
        let errorMsg = error.message;
        if (errorMsg.includes('انتهت صلاحية الجلسة')) {
            errorMsg = 'انتهت صلاحية الجلسة. سيتم إعادة توجيهك لتسجيل الدخول.';
            showMessage('error', errorMsg, 'admin-message');
            setTimeout(() => logoutLocal(true), 2000);
        } else {
            showMessage('error', errorMsg, 'admin-message');
        }
    } finally {
        if (button) button.disabled = false;
    }
}

function initAdminTags() {
    const input = document.getElementById('manga-tags-input');
    if (input) {
        input.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addTag(input.value);
                input.value = '';
            }
        });
    }
    const editInput = document.getElementById('edit-manga-tags-input');
    if (editInput) {
        editInput.addEventListener('keydown', (e) => {
            if (e.key === 'Enter') {
                e.preventDefault();
                addEditTag(editInput.value);
                editInput.value = '';
            }
        });
    }
}

function initAdminSelectors() {
    const select = document.getElementById('chapter-manga-id');
    if (select) {
        select.addEventListener('change', async () => {
            currentAdminMangaId = select.value;
            await loadAdminChapterList(currentAdminMangaId);
        });
    }
    const editMangaBtn = document.getElementById('edit-manga-btn');
    if (editMangaBtn) editMangaBtn.addEventListener('click', openEditMangaModal);
}

function initFileUpload() {
    const fileInput = document.getElementById('chapter-image-files');
    const selectButton = document.getElementById('select-files-button');
    if (selectButton && fileInput) {
        selectButton.addEventListener('click', () => fileInput.click());
        fileInput.addEventListener('change', () => {
            const countSpan = document.getElementById('selected-files-count');
            if (countSpan) {
                countSpan.textContent = fileInput.files.length ? `${fileInput.files.length} ملف(ات) مختارة` : 'لم يتم اختيار أي ملف';
            }
        });
    }
}

function initModals() {
    const closeBtns = document.querySelectorAll('.close-modal');
    closeBtns.forEach(btn => {
        btn.addEventListener('click', () => {
            document.getElementById('editMangaModal').style.display = 'none';
            document.getElementById('editChapterModal').style.display = 'none';
        });
    });
    window.addEventListener('click', (e) => {
        if (e.target.classList.contains('modal')) {
            e.target.style.display = 'none';
        }
    });
    const editMangaForm = document.getElementById('edit-manga-form');
    if (editMangaForm) editMangaForm.addEventListener('submit', handleEditMangaSubmit);
    const editChapterForm = document.getElementById('edit-chapter-form');
    if (editChapterForm) editChapterForm.addEventListener('submit', handleEditChapterSubmit);
}

async function initAdminPage() {
    if (!requireAuth()) return;

    const isAdmin = await ensureAdmin();
    if (!isAdmin) {
        window.location.href = webPagePath('control.html');
        return;
    }

    const addMangaForm = document.getElementById('add-manga-form');
    if (addMangaForm) addMangaForm.addEventListener('submit', handleCreateManga);
    const addChapterForm = document.getElementById('add-chapter-form');
    if (addChapterForm) addChapterForm.addEventListener('submit', handleCreateChapter);
    const uploadBtn = document.getElementById('upload-chapter-images-button');
    if (uploadBtn) uploadBtn.addEventListener('click', handleUploadChapterImages);

    initAdminTags();
    initAdminSelectors();
    initFileUpload();
    initModals();
    renderTags();
    await loadMangaOptions();
}

document.addEventListener('DOMContentLoaded', () => {
    const logoutBtn = document.getElementById('logout-button');
    if (logoutBtn) {
        logoutBtn.addEventListener('click', async () => {
            try {
                await apiFetch('/auth/logout', { method: 'POST' });
            } catch (e) { /* تجاهل */ }
            logoutLocal(true);
        });
    }
});

document.addEventListener('DOMContentLoaded', initAdminPage);