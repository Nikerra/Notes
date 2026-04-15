const API_URL = window.location.origin + '/api';
let currentFilter = 'all';
let notes = [];
let currentView = 'list';
let currentDate = new Date();
let editingNoteId = null;
let pickerCurrentDate = new Date();
let pickerSelectedDate = null;
let pickerTargetInput = null;
let pickerClearBtn = null;

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
        let url = `${API_URL}/notes`;
        if (currentFilter === 'overdue') {
            url = `${API_URL}/notes/overdue`;
        } else if (currentFilter !== 'all') {
            url = `${API_URL}/notes?category=${currentFilter}`;
        }
        
        const response = await fetch(url);
        if (!response.ok) throw new Error('Server error');
        
        notes = await response.json();
        await db.saveNotes(notes);
        renderNotes();
        if (currentView === 'calendar') {
            renderCalendar();
        }
        showToast('Синхронизация успешна');
    } catch (error) {
        console.log('Using offline data');
        notes = await db.getNotes();
        renderNotes();
        if (currentView === 'calendar') {
            renderCalendar();
        }
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
        if (currentView === 'calendar') {
            renderCalendar();
        }
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
        if (currentView === 'calendar') {
            renderCalendar();
        }
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
    if (currentView === 'calendar') {
        renderCalendar();
    }
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
    if (currentView === 'calendar') {
        renderCalendar();
    }
}

function formatDate(dateStr) {
    if (!dateStr) return '';
    const date = new Date(dateStr);
    const now = new Date();
    const diff = date - now;
    const days = Math.ceil(diff / (1000 * 60 * 60 * 24));
    
    if (days < 0) {
        return { text: 'Просрочено', class: 'overdue' };
    } else if (days === 0) {
        return { text: 'Сегодня', class: 'upcoming' };
    } else if (days === 1) {
        return { text: 'Завтра', class: 'upcoming' };
    } else if (days <= 7) {
        return { text: `Через ${days} дн.`, class: 'upcoming' };
    }
    
    return { 
        text: date.toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' }), 
        class: '' 
    };
}

function formatDisplayDate(date) {
    if (!date) return '';
    const d = new Date(date);
    return d.toLocaleDateString('ru-RU', { 
        day: 'numeric', 
        month: 'short',
        year: 'numeric',
        hour: '2-digit',
        minute: '2-digit'
    });
}

// Custom Date Picker Functions
function openDatePicker(inputId, clearBtnId) {
    pickerTargetInput = document.getElementById(inputId);
    pickerClearBtn = document.getElementById(clearBtnId);
    
    const currentValue = pickerTargetInput.dataset.value;
    if (currentValue) {
        pickerSelectedDate = new Date(currentValue);
        pickerCurrentDate = new Date(pickerSelectedDate);
        document.getElementById('pickerTime').value = 
            String(pickerSelectedDate.getHours()).padStart(2, '0') + ':' + 
            String(pickerSelectedDate.getMinutes()).padStart(2, '0');
    } else {
        pickerSelectedDate = null;
        pickerCurrentDate = new Date();
        document.getElementById('pickerTime').value = '12:00';
    }
    
    renderPicker();
    document.getElementById('datePickerOverlay').style.display = 'flex';
}

function closeDatePicker() {
    document.getElementById('datePickerOverlay').style.display = 'none';
}

function renderPicker() {
    const monthNames = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
                        'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
    
    const year = pickerCurrentDate.getFullYear();
    const month = pickerCurrentDate.getMonth();
    
    document.getElementById('pickerMonth').textContent = `${monthNames[month]} ${year}`;
    
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const startDay = (firstDay.getDay() + 6) % 7;
    
    const container = document.getElementById('pickerDays');
    container.innerHTML = '';
    
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    // Previous month
    const prevMonth = new Date(year, month, 0);
    for (let i = startDay - 1; i >= 0; i--) {
        const day = prevMonth.getDate() - i;
        const dayEl = document.createElement('div');
        dayEl.className = 'date-picker-day other-month';
        dayEl.textContent = day;
        container.appendChild(dayEl);
    }
    
    // Current month
    for (let day = 1; day <= lastDay.getDate(); day++) {
        const dayEl = document.createElement('div');
        dayEl.className = 'date-picker-day';
        dayEl.textContent = day;
        
        const date = new Date(year, month, day);
        if (date.getTime() === today.getTime()) {
            dayEl.classList.add('today');
        }
        
        if (pickerSelectedDate) {
            const selectedDateOnly = new Date(pickerSelectedDate);
            selectedDateOnly.setHours(0, 0, 0, 0);
            if (date.getTime() === selectedDateOnly.getTime()) {
                dayEl.classList.add('selected');
            }
        }
        
        dayEl.onclick = () => selectPickerDate(year, month, day);
        container.appendChild(dayEl);
    }
    
    // Next month
    const remaining = 42 - (startDay + lastDay.getDate());
    for (let day = 1; day <= remaining; day++) {
        const dayEl = document.createElement('div');
        dayEl.className = 'date-picker-day other-month';
        dayEl.textContent = day;
        container.appendChild(dayEl);
    }
}

function selectPickerDate(year, month, day) {
    pickerSelectedDate = new Date(year, month, day);
    renderPicker();
}

function confirmPickerDate() {
    if (!pickerSelectedDate) {
        closeDatePicker();
        return;
    }
    
    const timeInput = document.getElementById('pickerTime').value;
    const [hours, minutes] = timeInput.split(':');
    
    pickerSelectedDate.setHours(parseInt(hours), parseInt(minutes), 0, 0);
    
    pickerTargetInput.value = formatDisplayDate(pickerSelectedDate);
    pickerTargetInput.dataset.value = pickerSelectedDate.toISOString();
    pickerClearBtn.style.display = 'flex';
    
    closeDatePicker();
}

function setToday() {
    pickerSelectedDate = new Date();
    pickerCurrentDate = new Date();
    document.getElementById('pickerTime').value = '12:00';
    renderPicker();
}

function clearDate() {
    pickerTargetInput.value = '';
    pickerTargetInput.dataset.value = '';
    pickerClearBtn.style.display = 'none';
}

function openEditModal(id) {
    const note = notes.find(n => n.id == id);
    if (!note) return;
    
    editingNoteId = id;
    
    document.getElementById('editTitleInput').value = note.title;
    document.getElementById('editContentInput').value = note.content || '';
    document.getElementById('editCategorySelect').value = note.category;
    
    const editDateInput = document.getElementById('editDueDateInput');
    const editClearBtn = document.getElementById('editClearDateBtn');
    
    if (note.dueDate) {
        editDateInput.value = formatDisplayDate(note.dueDate);
        editDateInput.dataset.value = note.dueDate;
        editClearBtn.style.display = 'flex';
    } else {
        editDateInput.value = '';
        editDateInput.dataset.value = '';
        editClearBtn.style.display = 'none';
    }
    
    document.getElementById('editModal').style.display = 'flex';
    document.getElementById('editTitleInput').focus();
}

function closeEditModal() {
    editingNoteId = null;
    document.getElementById('editModal').style.display = 'none';
    document.getElementById('editTitleInput').value = '';
    document.getElementById('editContentInput').value = '';
    document.getElementById('editDueDateInput').value = '';
    document.getElementById('editDueDateInput').dataset.value = '';
}

async function saveEdit() {
    if (editingNoteId === null) return;
    
    const title = document.getElementById('editTitleInput').value.trim();
    if (!title) {
        document.getElementById('editTitleInput').focus();
        return;
    }
    
    const note = notes.find(n => n.id == editingNoteId);
    if (!note) return;
    
    const dueDateValue = document.getElementById('editDueDateInput').dataset.value;
    
    await updateNote(editingNoteId, {
        title: title,
        content: document.getElementById('editContentInput').value.trim(),
        category: document.getElementById('editCategorySelect').value,
        completed: note.completed,
        dueDate: dueDateValue || null
    });
    
    closeEditModal();
    showToast('Заметка обновлена');
}

function renderNotes() {
    const container = document.getElementById('notesList');
    
    let filtered = notes;
    if (currentFilter === 'overdue') {
        const now = new Date();
        filtered = notes.filter(n => n.dueDate && new Date(n.dueDate) < now && !n.completed);
    } else if (currentFilter !== 'all') {
        filtered = notes.filter(n => n.category === currentFilter);
    }
    
    if (filtered.length === 0) {
        container.innerHTML = `
            <div class="empty-state">
                <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5">
                    <path d="M9 5H7a2 2 0 00-2 2v12a2 2 0 002 2h10a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"/>
                </svg>
                <h3>Нет заметок</h3>
                <p>${currentFilter === 'overdue' ? 'Нет просроченных задач' : 'Добавьте первую заметку'}</p>
            </div>
        `;
        return;
    }
    
    container.innerHTML = filtered.map(note => {
        const dueDateInfo = note.dueDate ? formatDate(note.dueDate) : null;
        return `
            <div class="note-card ${note.completed ? 'completed' : ''}" data-id="${note.id}">
                <div class="note-header">
                    <div class="note-checkbox ${note.completed ? 'checked' : ''}" onclick="toggleComplete(${note.id})">
                        ${note.completed ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M5 13l4 4L19 7"/></svg>' : ''}
                    </div>
                    <div class="note-title">${escapeHtml(note.title)}</div>
                </div>
                ${note.content ? `<div class="note-content">${escapeHtml(note.content)}</div>` : ''}
                ${dueDateInfo ? `
                    <div class="note-due-date ${dueDateInfo.class}">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                            <circle cx="12" cy="12" r="10"></circle>
                            <polyline points="12 6 12 12 16 14"></polyline>
                        </svg>
                        ${dueDateInfo.text}
                    </div>
                ` : ''}
                <div class="note-footer">
                    <span class="note-category ${note.category}">${note.category === 'work' ? 'Работа' : 'Личное'}</span>
                    <div class="note-actions">
                        <button class="note-edit" onclick="openEditModal(${note.id})" title="Редактировать">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                                <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                            </svg>
                        </button>
                        <button class="note-delete" onclick="handleDelete(${note.id})" title="Удалить">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        `;
    }).join('');
}

function renderCalendar() {
    const monthNames = ['Январь', 'Февраль', 'Март', 'Апрель', 'Май', 'Июнь',
                        'Июль', 'Август', 'Сентябрь', 'Октябрь', 'Ноябрь', 'Декабрь'];
    
    const year = currentDate.getFullYear();
    const month = currentDate.getMonth();
    
    document.getElementById('currentMonth').textContent = `${monthNames[month]} ${year}`;
    
    const firstDay = new Date(year, month, 1);
    const lastDay = new Date(year, month + 1, 0);
    const startDay = (firstDay.getDay() + 6) % 7;
    
    const daysContainer = document.getElementById('calendarDays');
    daysContainer.innerHTML = '';
    
    const today = new Date();
    today.setHours(0, 0, 0, 0);
    
    const prevMonth = new Date(year, month, 0);
    for (let i = startDay - 1; i >= 0; i--) {
        const day = prevMonth.getDate() - i;
        const dayEl = createDayElement(day, true, false);
        daysContainer.appendChild(dayEl);
    }
    
    for (let day = 1; day <= lastDay.getDate(); day++) {
        const date = new Date(year, month, day);
        const isToday = date.getTime() === today.getTime();
        const hasNotes = notes.some(n => {
            if (!n.dueDate) return false;
            const dueDate = new Date(n.dueDate);
            return dueDate.toDateString() === date.toDateString();
        });
        
        const dayEl = createDayElement(day, false, isToday, hasNotes, date);
        daysContainer.appendChild(dayEl);
    }
    
    const remaining = 42 - (startDay + lastDay.getDate());
    for (let day = 1; day <= remaining; day++) {
        const dayEl = createDayElement(day, true, false);
        daysContainer.appendChild(dayEl);
    }
    
    showNotesForDate(today);
}

function createDayElement(day, isOtherMonth, isToday, hasNotes = false, date = null) {
    const dayEl = document.createElement('div');
    dayEl.className = 'calendar-day';
    if (isOtherMonth) dayEl.classList.add('other-month');
    if (isToday) dayEl.classList.add('today');
    if (hasNotes) dayEl.classList.add('has-notes');
    dayEl.textContent = day;
    
    if (date) {
        dayEl.onclick = () => showNotesForDate(date);
    }
    
    return dayEl;
}

function showNotesForDate(date) {
    const notesContainer = document.getElementById('calendarNotes');
    const dateNotes = notes.filter(n => {
        if (!n.dueDate) return false;
        const dueDate = new Date(n.dueDate);
        return dueDate.toDateString() === date.toDateString();
    });
    
    const dateStr = date.toLocaleDateString('ru-RU', { 
        day: 'numeric', 
        month: 'long', 
        year: 'numeric' 
    });
    
    let html = `<h3>Заметки на ${dateStr}</h3>`;
    
    if (dateNotes.length === 0) {
        html += '<p style="color: #999; font-size: 14px;">Нет заметок</p>';
    } else {
        html += dateNotes.map(note => `
            <div class="note-card ${note.completed ? 'completed' : ''}" style="margin-top: 10px;">
                <div class="note-header">
                    <div class="note-checkbox ${note.completed ? 'checked' : ''}" onclick="toggleComplete(${note.id})">
                        ${note.completed ? '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><path d="M5 13l4 4L19 7"/></svg>' : ''}
                    </div>
                    <div class="note-title">${escapeHtml(note.title)}</div>
                </div>
                ${note.content ? `<div class="note-content">${escapeHtml(note.content)}</div>` : ''}
                <div class="note-footer">
                    <span class="note-category ${note.category}">${note.category === 'work' ? 'Работа' : 'Личное'}</span>
                    <div class="note-actions">
                        <button class="note-edit" onclick="openEditModal(${note.id})" title="Редактировать">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M11 4H4a2 2 0 00-2 2v14a2 2 0 002 2h14a2 2 0 002-2v-7"/>
                                <path d="M18.5 2.5a2.121 2.121 0 013 3L12 15l-4 1 1-4 9.5-9.5z"/>
                            </svg>
                        </button>
                        <button class="note-delete" onclick="handleDelete(${note.id})" title="Удалить">
                            <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
                                <path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2m3 0v14a2 2 0 01-2 2H7a2 2 0 01-2-2V6h14"/>
                            </svg>
                        </button>
                    </div>
                </div>
            </div>
        `).join('');
    }
    
    notesContainer.innerHTML = html;
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
        z-index: 3000;
        animation: fadeIn 0.3s ease;
    `;
    toast.textContent = message;
    document.body.appendChild(toast);
    
    setTimeout(() => {
        toast.style.opacity = '0';
        setTimeout(() => toast.remove(), 300);
    }, 2000);
}

// Event Listeners

// Add new note
document.getElementById('addBtn').addEventListener('click', async () => {
    const titleInput = document.getElementById('titleInput');
    const contentInput = document.getElementById('contentInput');
    const categorySelect = document.getElementById('categorySelect');
    const dueDateInput = document.getElementById('dueDateInput');
    
    const title = titleInput.value.trim();
    if (!title) {
        titleInput.focus();
        return;
    }
    
    const dueDateValue = dueDateInput.dataset.value;
    
    await createNote({
        title: title,
        content: contentInput.value.trim(),
        category: categorySelect.value,
        completed: false,
        dueDate: dueDateValue || null
    });
    
    titleInput.value = '';
    contentInput.value = '';
    dueDateInput.value = '';
    dueDateInput.dataset.value = '';
    document.getElementById('clearDateBtn').style.display = 'none';
    titleInput.focus();
});

// Date picker events
document.getElementById('dueDateInput').addEventListener('click', () => {
    openDatePicker('dueDateInput', 'clearDateBtn');
});

document.getElementById('editDueDateInput').addEventListener('click', () => {
    openDatePicker('editDueDateInput', 'editClearDateBtn');
});

document.getElementById('clearDateBtn').addEventListener('click', (e) => {
    e.stopPropagation();
    clearDate();
});

document.getElementById('editClearDateBtn').addEventListener('click', (e) => {
    e.stopPropagation();
    pickerTargetInput = document.getElementById('editDueDateInput');
    pickerClearBtn = document.getElementById('editClearDateBtn');
    clearDate();
});

document.getElementById('prevMonthPicker').addEventListener('click', () => {
    pickerCurrentDate.setMonth(pickerCurrentDate.getMonth() - 1);
    renderPicker();
});

document.getElementById('nextMonthPicker').addEventListener('click', () => {
    pickerCurrentDate.setMonth(pickerCurrentDate.getMonth() + 1);
    renderPicker();
});

document.getElementById('pickerTodayBtn').addEventListener('click', setToday);

document.getElementById('pickerOkBtn').addEventListener('click', confirmPickerDate);

document.getElementById('datePickerOverlay').addEventListener('click', (e) => {
    if (e.target.id === 'datePickerOverlay') {
        closeDatePicker();
    }
});

// Enter key for new note form
document.getElementById('titleInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        document.getElementById('addBtn').click();
    }
});

// Enter key for edit modal
document.getElementById('editTitleInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter') {
        saveEdit();
    }
});

document.getElementById('editContentInput').addEventListener('keypress', (e) => {
    if (e.key === 'Enter' && e.ctrlKey) {
        saveEdit();
    }
});

// Close modals on Escape key
document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') {
        if (editingNoteId !== null) {
            closeEditModal();
        }
        if (document.getElementById('datePickerOverlay').style.display === 'flex') {
            closeDatePicker();
        }
    }
});

// Filter buttons
document.querySelectorAll('.filter-btn').forEach(btn => {
    btn.addEventListener('click', async () => {
        document.querySelectorAll('.filter-btn').forEach(b => b.classList.remove('active'));
        btn.classList.add('active');
        currentFilter = btn.dataset.filter;
        
        try {
            let url = `${API_URL}/notes`;
            if (currentFilter === 'overdue') {
                url = `${API_URL}/notes/overdue`;
            } else if (currentFilter !== 'all') {
                url = `${API_URL}/notes?category=${currentFilter}`;
            }
            
            const response = await fetch(url);
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

// View switching
document.getElementById('listViewBtn').addEventListener('click', () => {
    currentView = 'list';
    document.getElementById('listViewBtn').classList.add('active');
    document.getElementById('calendarViewBtn').classList.remove('active');
    document.getElementById('calendarView').style.display = 'none';
    document.getElementById('notesList').style.display = 'block';
    document.getElementById('listControls').style.display = 'flex';
    renderNotes();
});

document.getElementById('calendarViewBtn').addEventListener('click', () => {
    currentView = 'calendar';
    document.getElementById('calendarViewBtn').classList.add('active');
    document.getElementById('listViewBtn').classList.remove('active');
    document.getElementById('calendarView').style.display = 'block';
    document.getElementById('notesList').style.display = 'none';
    document.getElementById('listControls').style.display = 'none';
    renderCalendar();
});

// Calendar navigation
document.getElementById('prevMonth').addEventListener('click', () => {
    currentDate.setMonth(currentDate.getMonth() - 1);
    renderCalendar();
});

document.getElementById('nextMonth').addEventListener('click', () => {
    currentDate.setMonth(currentDate.getMonth() + 1);
    renderCalendar();
});

// Service Worker
if ('serviceWorker' in navigator) {
    navigator.serviceWorker.register('/static/sw.js').catch(() => {});
}

// Initial sync
syncWithServer();
