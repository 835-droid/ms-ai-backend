// Favorites Page Logic - Full Favorite Lists Implementation

let allLists = [];
let currentList = null;
let currentListManga = [];

// ========== API Helpers ==========
async function fetchLists() {
    try {
        const data = await apiFetch('/mangas/lists');
        return data.items || [];
    } catch (error) {
        console.error('Failed to fetch lists:', error);
        return [];
    }
}

async function fetchListItems(listID) {
    try {
        const data = await apiFetch(`/mangas/lists/${encodeURIComponent(listID)}/items`);
        return data.items || [];
    } catch (error) {
        console.error('Failed to fetch list items:', error);
        return [];
    }
}

async function createList(name, description, isPublic) {
    try {
        const data = await apiFetch('/mangas/lists', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description, is_public: isPublic })
        });
        return data;
    } catch (error) {
        console.error('Failed to create list:', error);
        throw error;
    }
}

async function updateList(listID, name, description, isPublic) {
    try {
        const data = await apiFetch(`/mangas/lists/${encodeURIComponent(listID)}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ name, description, is_public: isPublic })
        });
        return data;
    } catch (error) {
        console.error('Failed to update list:', error);
        throw error;
    }
}

async function deleteList(listID) {
    try {
        await apiFetch(`/mangas/lists/${encodeURIComponent(listID)}`, {
            method: 'DELETE'
        });
        return true;
    } catch (error) {
        console.error('Failed to delete list:', error);
        throw error;
    }
}

async function addMangaToList(listID, mangaID, notes = '') {
    try {
        await apiFetch(`/mangas/lists/${encodeURIComponent(listID)}/items`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ manga_id: mangaID, notes })
        });
        return true;
    } catch (error) {
        console.error('Failed to add manga to list:', error);
        throw error;
    }
}

async function removeMangaFromList(listID, mangaID) {
    try {
        await apiFetch(`/mangas/lists/${encodeURIComponent(listID)}/items/${encodeURIComponent(mangaID)}`, {
            method: 'DELETE'
        });
        return true;
    } catch (error) {
        console.error('Failed to remove manga from list:', error);
        throw error;
    }
}

// ========== Render Functions ==========
function renderLists(lists) {
    const container = document.getElementById('lists-container');
    if (!container) return;

    if (!lists || lists.length === 0) {
        container.innerHTML = `
            <li style="padding: 1rem; text-align: center; color: var(--text-muted); font-size: 0.9rem;">
                لا توجد قوائم بعد.<br>أنشئ قائمة جديدة!
            </li>
        `;
        return;
    }

    container.innerHTML = lists.map(list => {
        const isActive = currentList && currentList.id === list.id;
        return `
            <li class="list-item ${isActive ? 'active' : ''}" data-list-id="${escapeHtml(list.id)}">
                <div class="list-item-info">
                    <span class="list-item-icon">📋</span>
                    <span class="list-item-name">${escapeHtml(list.name)}</span>
                    <span class="list-item-count">${list.manga_count || 0}</span>
                </div>
                <div class="list-item-actions">
                    <button class="btn-list-action edit-list-btn" title="تعديل" data-list-id="${escapeHtml(list.id)}">
                        <i class="fas fa-edit"></i>
                    </button>
                    <button class="btn-list-action delete delete-list-btn" title="حذف" data-list-id="${escapeHtml(list.id)}">
                        <i class="fas fa-trash"></i>
                    </button>
                </div>
            </li>
        `;
    }).join('');

    // Add event listeners
    container.querySelectorAll('.list-item').forEach(item => {
        item.addEventListener('click', (e) => {
            // Don't trigger if clicking on action buttons
            if (e.target.closest('.list-item-actions')) return;
            const listID = item.dataset.listId;
            selectList(listID);
        });
    });

    container.querySelectorAll('.edit-list-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const listID = btn.dataset.listId;
            openEditModal(listID);
        });
    });

    container.querySelectorAll('.delete-list-btn').forEach(btn => {
        btn.addEventListener('click', (e) => {
            e.stopPropagation();
            const listID = btn.dataset.listId;
            confirmDeleteList(listID);
        });
    });
}

function renderMangaGrid(mangaItems) {
    const grid = document.getElementById('manga-grid');
    const emptyState = document.getElementById('empty-state');
    if (!grid) return;

    if (!mangaItems || mangaItems.length === 0) {
        grid.style.display = 'none';
        if (emptyState) emptyState.style.display = 'block';
        return;
    }

    grid.style.display = 'grid';
    if (emptyState) emptyState.style.display = 'none';

    grid.innerHTML = mangaItems.map(item => {
        const manga = item.manga;
        if (!manga) return '';
        const mangaID = manga.id || manga._id;
        const cover = manga.cover_image || 'https://via.placeholder.com/400x560?text=Manga';
        const notes = item.notes || '';

        return `
            <div class="manga-card" data-manga-id="${escapeHtml(mangaID)}">
                <a class="manga-card-link" href="manga-details.html?id=${encodeURIComponent(mangaID)}">
                    <img src="${escapeHtml(cover)}" alt="${escapeHtml(manga.title)}" onerror="this.src='https://via.placeholder.com/400x560?text=Manga'">
                    <div class="manga-card-info">
                        <h4 class="manga-card-title">${escapeHtml(manga.title)}</h4>
                        ${notes ? `<p style="font-size: 0.75rem; color: var(--text-secondary); margin: 0.25rem 0;">${escapeHtml(notes.substring(0, 50))}${notes.length > 50 ? '...' : ''}</p>` : ''}
                    </div>
                </a>
                <div class="manga-card-actions" style="padding: 0.5rem;">
                    <button class="btn-remove-manga" data-list-id="${currentList ? escapeHtml(currentList.id) : ''}" data-manga-id="${escapeHtml(mangaID)}" title="إزالة من القائمة">
                        <i class="fas fa-times"></i> إزالة
                    </button>
                </div>
            </div>
        `;
    }).join('');

    // Add remove event listeners
    grid.querySelectorAll('.btn-remove-manga').forEach(btn => {
        btn.addEventListener('click', async (e) => {
            e.preventDefault();
            e.stopPropagation();
            const listID = btn.dataset.listId;
            const mangaID = btn.dataset.mangaId;
            if (listID && mangaID) {
                await handleRemoveManga(listID, mangaID);
            }
        });
    });
}

// ========== Actions ==========
async function selectList(listID) {
    const list = allLists.find(l => l.id === listID);
    if (!list) return;

    currentList = list;

    // Update UI
    document.querySelectorAll('.list-item').forEach(item => {
        item.classList.toggle('active', item.dataset.listId === listID);
    });

    document.getElementById('current-list-title').textContent = list.name;

    // Show edit/delete buttons
    document.getElementById('btn-edit-list').style.display = 'inline-flex';
    document.getElementById('btn-delete-list').style.display = 'inline-flex';

    // Load list items
    const items = await fetchListItems(listID);
    currentListManga = items;
    renderMangaGrid(items);
}

async function handleCreateList() {
    const nameInput = document.getElementById('list-name');
    const descInput = document.getElementById('list-description');
    const publicCheck = document.getElementById('list-public');

    const name = nameInput.value.trim();
    const description = descInput.value.trim();
    const isPublic = publicCheck.checked;

    if (!name) {
        showToast('الرجاء إدخال اسم القائمة', 'error');
        return;
    }

    try {
        await createList(name, description, isPublic);
        showToast('تم إنشاء القائمة بنجاح', 'success');
        closeModal();
        await refreshLists();
    } catch (error) {
        showToast('فشل إنشاء القائمة: ' + (error.message || 'خطأ غير معروف'), 'error');
    }
}

async function handleUpdateList() {
    if (!currentList) return;

    const nameInput = document.getElementById('list-name');
    const descInput = document.getElementById('list-description');
    const publicCheck = document.getElementById('list-public');

    const name = nameInput.value.trim();
    const description = descInput.value.trim();
    const isPublic = publicCheck.checked;

    if (!name) {
        showToast('الرجاء إدخال اسم القائمة', 'error');
        return;
    }

    try {
        await updateList(currentList.id, name, description, isPublic);
        showToast('تم تعديل القائمة بنجاح', 'success');
        closeModal();
        await refreshLists();
    } catch (error) {
        showToast('فشل تعديل القائمة: ' + (error.message || 'خطأ غير معروف'), 'error');
    }
}

function openEditModal(listID) {
    const list = allLists.find(l => l.id === listID);
    if (!list) return;

    currentList = list;

    document.getElementById('modal-title').textContent = 'تعديل القائمة';
    document.getElementById('list-name').value = list.name;
    document.getElementById('list-description').value = list.description || '';
    document.getElementById('list-public').checked = list.is_public || false;

    document.getElementById('list-modal').classList.add('active');
}

function confirmDeleteList(listID) {
    if (!confirm('هل أنت متأكد من حذف هذه القائمة؟ لا يمكن التراجع عن هذا الإجراء.')) {
        return;
    }

    deleteList(listID).then(() => {
        showToast('تم حذف القائمة بنجاح', 'success');
        currentList = null;
        document.getElementById('current-list-title').textContent = 'جميع القوائم';
        document.getElementById('btn-edit-list').style.display = 'none';
        document.getElementById('btn-delete-list').style.display = 'none';
        document.getElementById('manga-grid').innerHTML = '';
        document.getElementById('empty-state').style.display = 'block';
        refreshLists();
    }).catch(err => {
        showToast('فشل حذف القائمة: ' + (err.message || 'خطأ غير معروف'), 'error');
    });
}

async function handleRemoveManga(listID, mangaID) {
    if (!confirm('هل تريد إزالة هذه المانجا من القائمة؟')) {
        return;
    }

    try {
        await removeMangaFromList(listID, mangaID);
        showToast('تمت الإزالة بنجاح', 'success');
        // Refresh current list
        const items = await fetchListItems(listID);
        currentListManga = items;
        renderMangaGrid(items);
        // Also refresh the list counts
        await refreshLists();
    } catch (error) {
        showToast('فشل الإزالة: ' + (error.message || 'خطأ غير معروف'), 'error');
    }
}

async function refreshLists() {
    allLists = await fetchLists();
    renderLists(allLists);
}

// ========== Modal Functions ==========
function openModal(mode = 'create') {
    if (mode === 'create') {
        document.getElementById('modal-title').textContent = 'إنشاء قائمة جديدة';
        document.getElementById('list-name').value = '';
        document.getElementById('list-description').value = '';
        document.getElementById('list-public').checked = false;
        currentList = null;
    }
    document.getElementById('list-modal').classList.add('active');
}

function closeModal() {
    document.getElementById('list-modal').classList.remove('active');
    currentList = null;
}

// ========== Toast ==========
function showToast(message, type = 'info') {
    if (typeof window.showToast === 'function') {
        window.showToast(message, type);
    } else {
        alert(message);
    }
}

// ========== Initialization ==========
document.addEventListener('DOMContentLoaded', () => {
    if (!requireAuth()) return;

    // Load lists
    refreshLists();

    // Add list button
    document.getElementById('btn-add-list')?.addEventListener('click', () => openModal('create'));

    // Edit list button
    document.getElementById('btn-edit-list')?.addEventListener('click', () => {
        if (currentList) {
            openEditModal(currentList.id);
        }
    });

    // Delete list button
    document.getElementById('btn-delete-list')?.addEventListener('click', () => {
        if (currentList) {
            confirmDeleteList(currentList.id);
        }
    });

    // Modal close
    document.getElementById('modal-close')?.addEventListener('click', closeModal);
    document.getElementById('modal-cancel')?.addEventListener('click', closeModal);

    // Modal overlay click to close
    document.getElementById('list-modal')?.addEventListener('click', (e) => {
        if (e.target === e.currentTarget) {
            closeModal();
        }
    });

    // Save button - handles both create and update
    document.getElementById('modal-save')?.addEventListener('click', () => {
        if (currentList) {
            handleUpdateList();
        } else {
            handleCreateList();
        }
    });

    // Logout button
    document.getElementById('logout-button')?.addEventListener('click', async () => {
        try {
            await apiFetch('/auth/logout', { method: 'POST' });
        } catch { /* ignore */ }
        logoutLocal(true);
    });
});