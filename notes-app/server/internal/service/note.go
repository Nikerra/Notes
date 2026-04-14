// Package service содержит бизнес-логику приложения.
// Сервисный слой координирует работу репозиториев и реализует бизнес-правила.
package service

import (
	"context"
	"errors"

	"notes-app/internal/domain"
)

// Ошибки, возвращаемые методами сервиса.
var (
	// ErrNoteNotFound - заметка не найдена
	ErrNoteNotFound = errors.New("note not found")

	// ErrInvalidTitle - заголовок пустой или невалидный
	ErrInvalidTitle = errors.New("title is required")

	// ErrInvalidCategory - категория не является допустимой
	ErrInvalidCategory = errors.New("invalid category")

	// ErrInvalidNoteID - идентификатор заметки невалиден (<= 0)
	ErrInvalidNoteID = errors.New("invalid note id")
)

// NoteService предоставляет методы для работы с заметками.
// Реализует бизнес-логику создания, обновления, получения и удаления заметок.
type NoteService struct {
	repo domain.NoteRepository
}

// NewNoteService создает новый экземпляр NoteService.
//
// Параметры:
//   - repo: реализация интерфейса NoteRepository для работы с хранилищем
//
// Пример:
//
//	repo := repository.NewSQLiteRepository(db)
//	svc := service.NewNoteService(repo)
func NewNoteService(repo domain.NoteRepository) *NoteService {
	return &NoteService{repo: repo}
}

// Create создает новую заметку с указанным заголовком, содержанием и категорией.
// Выполняет валидацию входных данных перед сохранением.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - title: заголовок заметки (обязательный, не пустой)
//   - content: содержание заметки (опциональный, может быть пустым)
//   - category: категория заметки (если пустая, используется CategoryPersonal)
//
// Возвращает:
//   - *domain.Note: созданная заметка с ID и временными метками
//   - error: ErrInvalidTitle если title пустой,
//     ErrInvalidCategory если категория невалидна,
//     или ошибка при сохранении в БД
//
// Пример:
//
//	note, err := svc.Create(ctx, "Купить продукты", "Молоко, хлеб", domain.CategoryPersonal)
//	if err != nil {
//	    // обработка ошибки
//	}
//	fmt.Printf("Created note with ID: %d\n", note.ID)
func (s *NoteService) Create(ctx context.Context, title, content string, category domain.Category) (*domain.Note, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}

	if category != "" && !category.IsValid() {
		return nil, ErrInvalidCategory
	}

	if category == "" {
		category = domain.CategoryPersonal
	}

	note := &domain.Note{
		Title:     title,
		Content:   content,
		Category:  string(category),
		Completed: false,
	}

	return s.repo.Create(ctx, note)
}

// GetByID возвращает заметку по её идентификатору.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки (должен быть > 0)
//
// Возвращает:
//   - *domain.Note: найденная заметка
//   - error: ErrInvalidNoteID если id <= 0,
//     ErrNoteNotFound если заметка не найдена
//
// Пример:
//
//	note, err := svc.GetByID(ctx, 1)
//	if errors.Is(err, service.ErrNoteNotFound) {
//	    // заметка не найдена
//	}
func (s *NoteService) GetByID(ctx context.Context, id int64) (*domain.Note, error) {
	if id <= 0 {
		return nil, ErrInvalidNoteID
	}

	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNoteNotFound
	}
	return note, nil
}

// GetAll возвращает все заметки с возможностью фильтрации по категории.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - filter: фильтр для поиска (nil возвращает все заметки)
//
// Возвращает:
//   - []*domain.Note: список заметок (пустой список, если ничего не найдено)
//   - error: ошибка при запросе к БД
//
// Пример:
//
//	// Получить все заметки
//	notes, err := svc.GetAll(ctx, nil)
//
//	// Получить только рабочие заметки
//	filter := &domain.NoteFilter{Category: domain.CategoryWork}
//	workNotes, err := svc.GetAll(ctx, filter)
func (s *NoteService) GetAll(ctx context.Context, filter *domain.NoteFilter) ([]*domain.Note, error) {
	return s.repo.GetAll(ctx, filter)
}

// Update обновляет существующую заметку.
// Выполняет валидацию входных данных перед обновлением.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки для обновления
//   - title: новый заголовок (обязательный, не пустой)
//   - content: новое содержание (опциональный)
//   - category: новая категория (если пустая, старая категория сохраняется)
//   - completed: новый статус выполнения
//
// Возвращает:
//   - *domain.Note: обновленная заметка с новыми временными метками
//   - error: ErrInvalidNoteID если id <= 0,
//     ErrInvalidTitle если title пустой,
//     ErrInvalidCategory если категория невалидна,
//     ErrNoteNotFound если заметка не найдена
//
// Пример:
//
//	updated, err := svc.Update(ctx, 1, "Новый заголовок", "Новое содержание", domain.CategoryWork, true)
func (s *NoteService) Update(ctx context.Context, id int64, title, content string, category domain.Category, completed bool) (*domain.Note, error) {
	if id <= 0 {
		return nil, ErrInvalidNoteID
	}

	if title == "" {
		return nil, ErrInvalidTitle
	}

	if category != "" && !category.IsValid() {
		return nil, ErrInvalidCategory
	}

	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNoteNotFound
	}

	note.Title = title
	note.Content = content
	if category != "" {
		note.Category = string(category)
	}
	note.Completed = completed

	return s.repo.Update(ctx, note)
}

// ToggleComplete переключает статус выполнения заметки.
// Если заметка была невыполненной, становится выполненной, и наоборот.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки
//
// Возвращает:
//   - *domain.Note: обновленная заметка
//   - error: ErrInvalidNoteID если id <= 0,
//     ErrNoteNotFound если заметка не найдена
//
// Пример:
//
//	note, err := svc.ToggleComplete(ctx, 1)
//	fmt.Printf("Completed: %v\n", note.Completed)
func (s *NoteService) ToggleComplete(ctx context.Context, id int64) (*domain.Note, error) {
	if id <= 0 {
		return nil, ErrInvalidNoteID
	}

	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrNoteNotFound
	}

	note.Completed = !note.Completed
	return s.repo.Update(ctx, note)
}

// Delete удаляет заметку по идентификатору.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - id: идентификатор заметки для удаления
//
// Возвращает:
//   - error: ErrInvalidNoteID если id <= 0,
//     ErrNoteNotFound если заметка не найдена,
//     или ошибка при удалении из БД
//
// Пример:
//
//	if err := svc.Delete(ctx, 1); err != nil {
//	    if errors.Is(err, service.ErrNoteNotFound) {
//	        // заметка уже удалена
//	    }
//	}
func (s *NoteService) Delete(ctx context.Context, id int64) error {
	if id <= 0 {
		return ErrInvalidNoteID
	}

	_, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrNoteNotFound
	}

	return s.repo.Delete(ctx, id)
}
