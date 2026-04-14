const API_URL = window.location.origin + '/api';
let currentFilter = 'all';
let notes = [];

const db = {
    async getNotes() {
        return new Promise((resolve) => {
            const data = localStorage.getItem('notes');
            resolve(data ? JSON.parse(data) : []);
        });
    },
    async saveNotes(data) {
        localStorage.setItem('notes', JSON.stringify(data));
    }
};

async function syncWithServer() {
    const syncBtn = document.getElementById('syncBtn');
    syncBtn.classList.add('syncing');
    
    try {
        const response = await fetch(`${API_URL}/notes?category=${currentFilter}`);
        if (!response.ok) throw new Error('Server error');
        
        notes = await response.json();
        await db.saveNotes(notes);
        renderNotes();
        showToast('Синхронизация успешна');
    } catch (error) {
        console.log('Using offline data');
        notes = await db.getNotes();
        renderNotes();
        showToast('Режим офлайн');
    } finally {
        syncBtn.classList.remove('syncing');
    }
}

async function createNote(note) {
    try {
        const response = await fetch(`${API_URL}/notes`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(note)
        });
        
        if (!response.ok) throw new Error('Server error');
        
        const newNote = await response.json();
        notes.unshift(newNote);
        await db.saveNotes(notes);
        renderNotes();
        return newNote;
    } catch (error) {
        const newNote = {
            ...note,
            id: Date.now(),
            createdAt: new Date().toISOString(),
            updatedAt: new Date().toISOString()
        };
        notes.unshift(newNote);
        await db.saveNotes(notes);
        renderNotes();
        return newNote;
    }
}

async function updateNote(id, updates) {
    try {
        const response = await fetch(`${API_URL}/notes/${id}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(updates)
        });
        
        if (!response.ok) throw new Error('Server error');
        
        const updated = await response.json();
        notes = notes.map(n => n.id == id ? updated : n);
    } catch (error) {
        notes = notes.map(n => {
            if (n.id == id) {
                return { ...n, ...updates, updatedAt: new Date().toISOString() };
            }
            return n;
        });
    }
    
    await db.saveNotes(notes);
    renderNotes();
}

async function deleteNote(id) {
    try {
        await fetch(`${API_URL}/notes/${id}`, { method: 'DELETE' });
    } catch (error) {
        console.log('Offline delete');
    }
    
    notes = notes.filter(n => n.id != id);
    await db.saveNotes(notes);
    renderNotes();
}

function renderNotes() {
    const container = document.getElementById('notesList');
    
    let filtered = notes;
    if (currentFilter !== 'all') {
        filtered = notes.filter(n => n.category === currentFilter);
    }
    
    if (filtered.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
                </svg>
                <h3>Нет заметок</h3>
                <p>Добавьте первую заметку</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = filtered.map(note => `
        <div class="note-card ${note.completed ? 'completed' : ''}" data-id="${note.id}">
            <div class="note-header">
                <div class="note-checkbox ${note.completed ? 'checked' : ''}" onclick="toggleComplete(${note.id})">
                    ${note.completed ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M5 13l4 4L19 7"/></svg>' : ''}
                </div>
                <div class="note-title">${escapeHtml(note.title)}</div>
            </div>
            ${note.content ? `<div class="note-content">${escapeHtml(note.content)}</div>` : ''}
            <div class="note-footer">
                <span class="note-category ${note.category}">${note.category === 'work' ? 'Работа' : 'Личное'}</span>
                <button class="note-delete" onclick="handleDelete(${note.id})" title="Удалить">
                    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                        <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14"/>
                    </svg>
                </button>
            </div>
        </div>
    `).join('');
}

function escapeHtml(text) {
    const div = document.createElement('div');
    div.textContent = text;
    return div.innerHTML;
}

function toggleComplete(id) {
    const note = notes.find(n => n.id == id);
    if (note) {
        updateNote(id, { ...note, completed: !note.completed });
    }
}

async function handleDelete(id) {
    if (confirm('Удалить заметку?')) {
        await deleteNote(id);
    }
}

function showToast(message) {
    const toast = document.createElement('div');
    toast.style.cssText = `
        position: fixed;
        bottom: 20px;
        left: 50%;
        transform: translateX(-50%);
        background: rgba(0,0,0,0.8);
        color: white;
        padding: 12px 24px;
        border-radius: 8px;
        z-index: 1000;
        animation: fadeIn 0.3s ease;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 300);
    }, 2000);
}

document.getElementById('addBtn').addEventListener('click', async () => {
    const titleInput = document.getElementById('titleInput');
    const contentInput = document.getElementById('contentInput');
    const categorySelect = document.getElementById('categorySelect');
    
    const title = titleInput.value.trim();
    if (!title) {
        titleInput.focus();
        return;
    }
    
    await createNote({
        title: title,
        content: contentInput.value.trim(),
        category: categorySelect.value,
        completed: false
    });
    
    titleInput.value = '';
    contentInput.value = '';
    titleInput.focus();
});

document.getElementById('titleInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        document.getElementById('addBtn').click();
    }
});

document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentFilter = btn.dataset.filter;
        
        try {
            const response = await fetch(`${API_URL}/notes?category=${currentFilter}`);
            if (response.ok) {
                notes = await response.json();
                await db.saveNotes(notes);
            }
        } catch (error) {
            notes = await db.getNotes();
        }
        
        renderNotes();
    });
});

document.getElementById('syncBtn').addEventListener('click', syncWithServer);

if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/static/sw.js').catch(() => {});
}

syncWithServer();
