// favorites.js - إدارة قوائم المفضلة

// State
let currentList = null;
let favoriteLists = [];
let mangaInCurrentList = [];

// DOM Elements
const listsContainer = document.getElementById('lists-container');
const mangaGrid = document.getElementById('manga-grid');
const emptyState = document.getElementById('empty-state');
const currentListTitle = document.getElementById('current-list-title');
const btnAddList = document.getElementById('btn-add-list');
const btnEditList = document.getElementById('btn-edit-list');
const btnDeleteList = document.getElementById('btn-delete-list');
const listModal = document.getElementById('list-modal');
const modalTitle = document.getElementById('modal-title');
const modalClose = document.getElementById('modal-close');
const modalCancel = document.getElementById('modal-cancel');
const modalSave = document.getElementById('modal-save');
const listNameInput = document.getElementById('list-name');
const listDescInput = document.getElementById('list-description');
const listPublicCheckbox = document.getElementById('list-public');

// Initialize
document.addEventListener('DOMContentLoaded', () => {
    initializeApp();
});

async function initializeApp() {
    // Check authentication
    const user = await checkAuth();
    if (!user) {
        window.location.href = 'login.html';
        return;
    }

    // Update user display
    updateUserDisplay(user);

    // Load favorite lists
    await loadFavoriteLists();

    // Setup event listeners
    setupEventListeners();

    // Show default view (all lists or first list)
    showAllLists();
}

// Event Listeners
function setupEventListeners() {
    btnAddList.addEventListener('click', () => showCreateModal());
    modalClose.addEventListener('click', hideModal);
    modalCancel.addEventListener('click', hideModal);
    modalSave.addEventListener('click', saveList);
    btnEditList.addEventListener('click', () => editCurrentList());
    btnDeleteList.addEventListener('click', () => deleteCurrentList());

    // Close modal on overlay click
    listModal.addEventListener('click', (e) => {
        if (e.target === listModal) {
            hideModal();
        }
    });

    // Keyboard shortcuts
    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            hideModal();
        }
    });
}

// API Functions
async function loadFavoriteLists() {
    try {
        const response = await api.get('/api/mangas/lists');
        if (response.success) {
            favoriteLists = response.data || [];
            renderLists();
        }
    } catch (error) {
        console.error('Error loading favorite lists:', error);
        showToast('فشل تحميل القوائم', 'error');
    }
}

async function loadMangaInList(listId) {
    try {
        const response = await api.get(`/api/mangas/lists/${listId}/items`);
        if (response.success) {
            mangaInCurrentList = response.data || [];
            renderMangaGrid();
        }
    } catch (error) {
        console.error('Error loading manga in list:', error);
        showToast('فشل تحميل المانجا', 'error');
    }
}

async function createList(name, description, isPublic) {
    try {
        const response = await api.post('/api/mangas/lists', {
            name: name,
            description: description,
            is_public: isPublic
        });
        if (response.success) {
            showToast('تم إنشاء القائمة بنجاح', 'success');
            await loadFavoriteLists();
            return response.data;
        }
    } catch (error) {
        console.error('Error creating list:', error);
        showToast('فشل إنشاء القائمة', 'error');
    }
    return null;
}

async function updateList(listId, name, description, isPublic) {
    try {
        const response = await api.put(`/api/mangas/lists/${listId}`, {
            name: name,
            description: description,
            is_public: isPublic
        });
        if (response.success) {
            showToast('تم تحديث القائمة بنجاح', 'success');
            await loadFavoriteLists();
            return response.data;
        }
    } catch (error) {
        console.error('Error updating list:', error);
        showToast('فشل تحديث القائمة', 'error');
    }
    return null;
}

async function deleteList(listId) {
    if (!confirm('هل أنت متأكد من حذف هذه القائمة؟')) {
        return;
    }

    try {
        const response = await api.delete(`/api/mangas/lists/${listId}`);
        if (response.success) {
            showToast('تم حذف القائمة بنجاح', 'success');
            await loadFavoriteLists();
            showAllLists();
        }
    } catch (error) {
        console.error('Error deleting list:', error);
        showToast('فشل حذف القائمة', 'error');
    }
}

async function addMangaToList(listId, mangaId) {
    try {
        const response = await api.post(`/api/mangas/lists/${listId}/items`, {
            manga_id: mangaId
        });
        if (response.success) {
            showToast('تمت إضافة المانجا إلى القائمة', 'success');
            if (currentList && currentList.id === listId) {
                await loadMangaInList(listId);
            }
        }
    } catch (error) {
        console.error('Error adding manga to list:', error);
        showToast('فشل إضافة المانجا', 'error');
    }
}

async function removeMangaFromList(listId, mangaId) {
    try {
        const response = await api.delete(`/api/mangas/lists/${listId}/items/${mangaId}`);
        if (response.success) {
            showToast('تمت إزالة المانجا من القائمة', 'success');
            if (currentList && currentList.id === listId) {
                await loadMangaInList(listId);
            }
        }
    } catch (error) {
        console.error('Error removing manga from list:', error);
        showToast('فشل إزالة المانجا', 'error');
    }
}

// Render Functions
function renderLists() {
    listsContainer.innerHTML = '';

    // Add "All Lists" option
    const allListsItem = createListItem({
        id: 'all',
        name: 'جميع القوائم',
        icon: 'fas fa-th',
        count: favoriteLists.reduce((sum, list) => sum + (list.item_count || 0), 0),
        isAll: true
    });
    listsContainer.appendChild(allListsItem);

    // Render each list
    favoriteLists.forEach(list => {
        const listItem = createListItem(list);
        listsContainer.appendChild(listItem);
    });
}

function createListItem(list) {
    const li = document.createElement('li');
    li.className = 'list-item';
    li.dataset.listId = list.id;

    if (currentList && currentList.id === list.id) {
        li.classList.add('active');
    }

    const icon = list.isAll ? list.icon : (list.is_public ? 'fas fa-globe' : 'fas fa-lock');
    const name = list.name || list.name;

    li.innerHTML = `
        <div class="list-item-info">
            <i class="list-item-icon ${icon}"></i>
            <span class="list-item-name">${name}</span>
        </div>
        <span class="list-item-count">${list.count || 0}</span>
        ${!list.isAll ? `
            <div class="list-item-actions">
                <button class="btn-list-action edit" title="تعديل القائمة">
                    <i class="fas fa-edit"></i>
                </button>
                <button class="btn-list-action delete" title="حذف القائمة">
                    <i class="fas fa-trash"></i>
                </button>
            </div>
        ` : ''}
    `;

    // Click to select list
    li.addEventListener('click', (e) => {
        if (!e.target.closest('.btn-list-action')) {
            selectList(list);
        }
    });

    // Edit button
    const editBtn = li.querySelector('.btn-list-action.edit');
    if (editBtn) {
        editBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            editList(list);
        });
    }

    // Delete button
    const deleteBtn = li.querySelector('.btn-list-action.delete');
    if (deleteBtn) {
        deleteBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            deleteList(list.id);
        });
    }

    return li;
}

function renderMangaGrid() {
    mangaGrid.innerHTML = '';

    if (!mangaInCurrentList || mangaInCurrentList.length === 0) {
        emptyState.style.display = 'block';
        mangaGrid.style.display = 'none';
        return;
    }

    emptyState.style.display = 'none';
    mangaGrid.style.display = 'grid';

    mangaInCurrentList.forEach(item => {
        const manga = item.manga || item;
        const card = createMangaCard(manga, item);
        mangaGrid.appendChild(card);
    });
}

function createMangaCard(manga, listItem) {
    const card = document.createElement('div');
    card.className = 'manga-card';

    const coverImage = manga.cover_image || 'placeholder-manga.jpg';
    const title = manga.title || 'مانجا غير معروفة';

    card.innerHTML = `
        <a href="manga-details.html?id=${manga.id}">
            <img src="${coverImage}" alt="${title}" loading="lazy">
        </a>
        <div class="manga-card-info">
            <h3 class="manga-card-title">
                <a href="manga-details.html?id=${manga.id}">${title}</a>
            </h3>
            <div class="manga-card-actions">
                <span class="manga-card-notes" title="${listItem.notes || ''}">
                    ${listItem.notes ? '<i class="fas fa-sticky-note"></i>' : ''}
                </span>
                <button class="btn-remove-manga" title="إزالة من القائمة">
                    <i class="fas fa-times"></i>
                </button>
            </div>
        </div>
    `;

    // Remove button
    const removeBtn = card.querySelector('.btn-remove-manga');
    removeBtn.addEventListener('click', (e) => {
        e.preventDefault();
        if (currentList) {
            removeMangaFromList(currentList.id, manga.id);
        }
    });

    return card;
}

// View Functions
function selectList(list) {
    currentList = list;

    // Update active state
    document.querySelectorAll('.list-item').forEach(item => {
        item.classList.remove('active');
        if (item.dataset.listId === list.id) {
            item.classList.add('active');
        }
    });

    // Update title
    currentListTitle.textContent = list.name || 'جميع القوائم';

    // Show/hide edit/delete buttons
    if (list.isAll) {
        btnEditList.style.display = 'none';
        btnDeleteList.style.display = 'none';
        showAllListsView();
    } else {
        btnEditList.style.display = 'inline-flex';
        btnDeleteList.style.display = 'inline-flex';
        loadMangaInList(list.id);
    }
}

function showAllLists() {
    currentList = { id: 'all', name: 'جميع القوائم', isAll: true };
    selectList(currentList);
}

function showAllListsView() {
    mangaGrid.innerHTML = '';
    emptyState.style.display = 'block';
    emptyState.innerHTML = `
        <i class="fas fa-heart"></i>
        <p>اختر قائمة لعرض المانجا فيها، أو أنشئ قائمة جديدة.</p>
    `;
    mangaGrid.style.display = 'none';
}

// Modal Functions
let editingList = null;

function showCreateModal() {
    editingList = null;
    modalTitle.textContent = 'إنشاء قائمة جديدة';
    modalSave.textContent = 'إنشاء';
    listNameInput.value = '';
    listDescInput.value = '';
    listPublicCheckbox.checked = false;
    listModal.classList.add('active');
    listNameInput.focus();
}

function editList(list) {
    editingList = list;
    modalTitle.textContent = 'تعديل القائمة';
    modalSave.textContent = 'حفظ';
    listNameInput.value = list.name || '';
    listDescInput.value = list.description || '';
    listPublicCheckbox.checked = list.is_public || false;
    listModal.classList.add('active');
    listNameInput.focus();
}

function editCurrentList() {
    if (currentList && !currentList.isAll) {
        editList(currentList);
    }
}

function deleteCurrentList() {
    if (currentList && !currentList.isAll) {
        deleteList(currentList.id);
    }
}

function hideModal() {
    listModal.classList.remove('active');
    editingList = null;
}

async function saveList() {
    const name = listNameInput.value.trim();
    if (!name) {
        showToast('الرجاء إدخال اسم القائمة', 'error');
        listNameInput.focus();
        return;
    }

    const description = listDescInput.value.trim();
    const isPublic = listPublicCheckbox.checked;

    if (editingList) {
        await updateList(editingList.id, name, description, isPublic);
        if (currentList && currentList.id === editingList.id) {
            currentList.name = name;
            currentList.description = description;
            currentList.is_public = isPublic;
            currentListTitle.textContent = name;
        }
    } else {
        const newList = await createList(name, description, isPublic);
        if (newList) {
            selectList(newList);
        }
    }

    hideModal();
}

// Utility Functions
function updateUserDisplay(user) {
    const displayName = document.getElementById('user-display-name');
    if (displayName) {
        displayName.textContent = user.username || user.email || 'مستخدم';
    }

    const avatar = document.querySelector('.navbar-user-avatar');
    if (avatar && user.username) {
        avatar.textContent = user.username.substring(0, 2).toUpperCase();
    }
}

function showToast(message, type = 'info') {
    if (typeof window.showToast === 'function') {
        window.showToast(message, type);
    } else {
        console.log(`[${type.toUpperCase()}] ${message}`);
    }
}

async function checkAuth() {
    // Simple auth check - in real app, use proper auth system
    const token = localStorage.getItem('token');
    if (!token) {
        return null;
    }

    try {
        // Try to get user profile
        const user = JSON.parse(localStorage.getItem('user') || '{}');
        return user;
    } catch {
        return null;
    }
}

// Export functions for use in other files
window.favorites = {
    addMangaToList,
    removeMangaFromList,
    loadFavoriteLists
};