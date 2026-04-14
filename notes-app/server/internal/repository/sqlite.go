// Package repository содержит реализации интерфейсов репозиториев.
// Отвечает за персистентное хранение данных в базе данных.
package repository

import (
	"context"
	"database/sql"
	"time"

	"notes-app/internal/domain"
)

// SQLiteRepository реализует интерфейс domain.NoteRepository для SQLite базы данных.
// Использует стандартный пакет database/sql для работы с БД.
type SQLiteRepository struct {
	db *sql.DB
}

// NewSQLiteRepository создает новый экземпляр SQLiteRepository.
//
// Параметры:
//   - db: открытое соединение с SQLite базой данных
//
// Пример:
//
//	db, err := sql.Open("sqlite3", "./data/notes.db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//	repo := repository.NewSQLiteRepository(db)
func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// Create создает новую заметку в базе данных.
// Автоматически устанавливает ID, CreatedAt и UpdatedAt.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - note: заметка для создания (поля ID, CreatedAt, UpdatedAt будут перезаписаны)
//
// Возвращает:
//   - *domain.Note: созданная заметка с установленным ID и временными метками
//   - error: ошибка при выполнении INSERT запроса
//
// SQL запрос:
//
//	INSERT INTO notes (title, content, category, completed, created_at, updated_at)
//	VALUES (?, ?, ?, ?, ?, ?)
func (r *SQLiteRepository) Create(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	query := `
		INSERT INTO notes (title, content, category, completed, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	now := time.Now()
	result, err := r.db.ExecContext(ctx, query,
		note.Title, note.Content, note.Category, note.Completed, now, now)
	if err != nil {
		return nil, err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, err
	}

	note.ID = id
	note.CreatedAt = now
	note.UpdatedAt = now
	return note, nil
}

// GetByID возвращает заметку по идентификатору из базы данных.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки
//
// Возвращает:
//   - *domain.Note: найденная заметка со всеми полями
//   - error: sql.ErrNoRows если заметка не найдена, или ошибка при выполнении SELECT запроса
//
// SQL запрос:
//
//	SELECT id, title, content, category, completed, created_at, updated_at
//	FROM notes WHERE id = ?
func (r *SQLiteRepository) GetByID(ctx context.Context, id int64) (*domain.Note, error) {
	query := `
		SELECT id, title, content, category, completed, created_at, updated_at
		FROM notes WHERE id = ?
	`
	note := &domain.Note{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&note.ID, &note.Title, &note.Content, &note.Category,
		&note.Completed, &note.CreatedAt, &note.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return note, nil
}

// GetAll возвращает все заметки из базы данных с возможностью фильтрации по категории.
// Заметки сортируются по дате создания в обратном порядке (новые первые).
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - filter: фильтр по категории (nil или пустая категория возвращает все заметки)
//
// Возвращает:
//   - []*domain.Note: список заметок (пустой список если ничего не найдено, но не nil)
//   - error: ошибка при выполнении SELECT запроса
//
// SQL запросы:
//
//	Без фильтра:
//	  SELECT ... FROM notes ORDER BY created_at DESC
//
//	С фильтром:
//	  SELECT ... FROM notes WHERE category = ? ORDER BY created_at DESC
func (r *SQLiteRepository) GetAll(ctx context.Context, filter *domain.NoteFilter) ([]*domain.Note, error) {
	var query string
	var args []interface{}

	if filter != nil && filter.Category != "" {
		query = `
			SELECT id, title, content, category, completed, created_at, updated_at
			FROM notes WHERE category = ? ORDER BY created_at DESC
		`
		args = append(args, filter.Category)
	} else {
		query = `
			SELECT id, title, content, category, completed, created_at, updated_at
			FROM notes ORDER BY created_at DESC
		`
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notes []*domain.Note
	for rows.Next() {
		note := &domain.Note{}
		if err := rows.Scan(
			&note.ID, &note.Title, &note.Content, &note.Category,
			&note.Completed, &note.CreatedAt, &note.UpdatedAt,
		); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}

	if notes == nil {
		notes = []*domain.Note{}
	}
	return notes, nil
}

// Update обновляет существующую заметку в базе данных.
// Автоматически обновляет поле UpdatedAt текущим временем.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - note: заметка с обновленными данными (ID должен быть установлен)
//
// Возвращает:
//   - *domain.Note: обновленная заметка с новым UpdatedAt
//   - error: ошибка при выполнении UPDATE запроса
//
// SQL запрос:
//
//	UPDATE notes
//	SET title = ?, content = ?, category = ?, completed = ?, updated_at = ?
//	WHERE id = ?
func (r *SQLiteRepository) Update(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	query := `
		UPDATE notes
		SET title = ?, content = ?, category = ?, completed = ?, updated_at = ?
		WHERE id = ?
	`
	now := time.Now()
	_, err := r.db.ExecContext(ctx, query,
		note.Title, note.Content, note.Category, note.Completed, now, note.ID)
	if err != nil {
		return nil, err
	}

	note.UpdatedAt = now
	return note, nil
}

// Delete удаляет заметку из базы данных по идентификатору.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки для удаления
//
// Возвращает:
//   - error: ошибка при выполнении DELETE запроса
//
// SQL запрос:
//
//	DELETE FROM notes WHERE id = ?
//
// Примечание: не возвращает ошибку, если заметка не существует.
func (r *SQLiteRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM notes WHERE id = ?`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

// Migrate создает таблицу notes в базе данных, если она не существует.
// Должна вызываться при запуске приложения для инициализации БД.
//
// Возвращает:
//   - error: ошибка при выполнении CREATE TABLE запроса
//
// SQL запрос:
//
//	CREATE TABLE IF NOT EXISTS notes (
//	    id INTEGER PRIMARY KEY AUTOINCREMENT,
//	    title TEXT NOT NULL,
//	    content TEXT,
//	    category TEXT DEFAULT 'personal',
//	    completed BOOLEAN DEFAULT 0,
//	    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
//	    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
//	)
//
// Пример:
//
//	repo := repository.NewSQLiteRepository(db)
//	if err := repo.Migrate(); err != nil {
//	    log.Fatal(err)
//	}
func (r *SQLiteRepository) Migrate() error {
	query := `
	CREATE TABLE IF NOT EXISTS notes (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		content TEXT,
		category TEXT DEFAULT 'personal',
		completed BOOLEAN DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`
	_, err := r.db.Exec(query)
	return err
}
