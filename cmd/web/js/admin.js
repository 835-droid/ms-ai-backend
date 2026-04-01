let tagList = [];
let currentAdminMangaId = '';

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

function getChapterImagesFromTextarea() {
    return splitLines(document.getElementById('chapter-urls')?.value);
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
                <div class="admin-chapter-item">
                    <div>
                        <strong>الفصل ${escapeHtml(chapter.number)}</strong>
                        <div>${escapeHtml(chapter.title || '')}</div>
                        <small>${Array.isArray(chapter.pages) ? chapter.pages.length : 0} صفحة</small>
                    </div>
                    <div class="admin-chapter-actions">
                        <a class="btn btn-secondary" href="manga-reader.html?mangaId=${encodeURIComponent(mangaId)}&chapterId=${encodeURIComponent(chapterId)}">فتح</a>
                        <button class="btn btn-danger" data-delete-chapter="${escapeHtml(chapterId)}">حذف</button>
                    </div>
                </div>
            `;
        }).join('');

        qsa('[data-delete-chapter]', box).forEach(btn => {
            btn.addEventListener('click', async () => {
                if (!confirm('حذف الفصل نهائياً؟')) return;
                await deleteChapter(mangaId, btn.dataset.deleteChapter);
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
    if (button) {
        button.disabled = true;
        button.textContent = 'جاري الحفظ...';
    }

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

        showMessage('success', 'تمت إضافة المانجا', 'admin-message');
        event.target.reset();
        tagList = [];
        renderTags();
        await loadMangaOptions();
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    } finally {
        if (button) {
            button.disabled = false;
            button.textContent = 'إضافة المانجا';
        }
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
    if (button) {
        button.disabled = true;
        button.textContent = 'جاري الحفظ...';
    }

    try {
        await apiFetch(`/mangas/${encodeURIComponent(mangaId)}/chapters`, {
            method: 'POST',
            body: JSON.stringify({
                title,
                number,
                images
            })
        });

        showMessage('success', 'تمت إضافة الفصل', 'admin-message');
        event.target.reset();
        await loadAdminChapterList(mangaId);
    } catch (error) {
        showMessage('error', error.message, 'admin-message');
    } finally {
        if (button) {
            button.disabled = false;
            button.textContent = 'إضافة الفصل';
        }
    }
}

function initAdminTags() {
    const input = document.getElementById('manga-tags-input');
    if (!input) return;

    input.addEventListener('keydown', (event) => {
        if (event.key === 'Enter') {
            event.preventDefault();
            addTag(input.value);
            input.value = '';
        }
    });
}

function initAdminSelectors() {
    const select = document.getElementById('chapter-manga-id');
    if (!select) return;

    select.addEventListener('change', async () => {
        currentAdminMangaId = select.value;
        await loadAdminChapterList(currentAdminMangaId);
    });
}

async function initAdminPage() {
    if (!requireAuth()) return;

    const isAdmin = await ensureAdmin();
    if (!isAdmin) {
        window.location.href = '/web/control.html';
        return;
    }

    document.getElementById('add-manga-form')?.addEventListener('submit', handleCreateManga);
    document.getElementById('add-chapter-form')?.addEventListener('submit', handleCreateChapter);

    initAdminTags();
    initAdminSelectors();
    renderTags();
    await loadMangaOptions();
}

document.addEventListener('DOMContentLoaded', initAdminPage);