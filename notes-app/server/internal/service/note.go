// Package service содержит бизнес-логику приложения.
// Сервисный слой координирует работу репозиториев и реализует бизнес-правила.
package service

import (
	"context"
	"errors"
	"time"

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

	// ErrInvalidDueDate - срок выполнения в прошлом (для невыполненных заметок)
	ErrInvalidDueDate = errors.New("due date cannot be in the past for incomplete notes")
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

// Create создает новую заметку с указанным заголовком, содержанием, категорией и сроком выполнения.
// Выполняет валидацию входных данных перед сохранением.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - title: заголовок заметки (обязательный, не пустой)
//   - content: содержание заметки (опциональный, может быть пустым)
//   - category: категория заметки (если пустая, используется CategoryPersonal)
//   - dueDate: срок выполнения заметки (опциональный, может быть nil)
//
// Возвращает:
//   - *domain.Note: созданная заметка с ID и временными метками
//   - error: ErrInvalidTitle если title пустой,
//     ErrInvalidCategory если категория невалидна,
//     ErrInvalidDueDate если срок в прошлом,
//     или ошибка при сохранении в БД
//
// Пример:
//
//	dueDate := time.Now().Add(24 * time.Hour)
//	note, err := svc.Create(ctx, "Купить продукты", "Молоко, хлеб", domain.CategoryPersonal, &dueDate)
//	if err != nil {
//	    // обработка ошибки
//	}
//	fmt.Printf("Created note with ID: %d\n", note.ID)
func (s *NoteService) Create(ctx context.Context, title, content string, category domain.Category, dueDate *time.Time) (*domain.Note, error) {
	if title == "" {
		return nil, ErrInvalidTitle
	}

	if category != "" && !category.IsValid() {
		return nil, ErrInvalidCategory
	}

	if category == "" {
		category = domain.CategoryPersonal
	}

	// Валидация срока выполнения (не должен быть в прошлом для невыполненных заметок)
	if dueDate != nil && dueDate.Before(time.Now()) {
		return nil, ErrInvalidDueDate
	}

	note := &domain.Note{
		Title:     title,
		Content:   content,
		Category:  string(category),
		Completed: false,
		DueDate:   dueDate,
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

// GetAll возвращает все заметки с возможностью фильтрации.
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
//
//	// Получить заметки на сегодня
//	today := time.Now()
//	startOfDay := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, today.Location())
//	endOfDay := startOfDay.Add(24 * time.Hour)
//	dateFilter := &domain.NoteFilter{
//	    DueDateFrom: &startOfDay,
//	    DueDateTo:   &endOfDay,
//	}
//	todayNotes, err := svc.GetAll(ctx, dateFilter)
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
//   - dueDate: новый срок выполнения (может быть nil для удаления срока)
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
//	dueDate := time.Now().Add(48 * time.Hour)
//	updated, err := svc.Update(ctx, 1, "Новый заголовок", "Новое содержание", domain.CategoryWork, true, &dueDate)
func (s *NoteService) Update(ctx context.Context, id int64, title, content string, category domain.Category, completed bool, dueDate *time.Time) (*domain.Note, error) {
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
	note.DueDate = dueDate

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

// GetUpcoming возвращает заметки с предстоящим сроком выполнения.
// Полезно для отображения в календаре или напоминаниях.
//
// Параметры:
//   - ctx: контекст для отмены операции
//   - days: количество дней вперед (например, 7 для недели)
//
// Возвращает:
//   - []*domain.Note: список заметок с предстоящим сроком
//   - error: ошибка при запросе к БД
//
// Пример:
//
//	// Получить заметки на ближайшую неделю
//	upcoming, err := svc.GetUpcoming(ctx, 7)
func (s *NoteService) GetUpcoming(ctx context.Context, days int) ([]*domain.Note, error) {
	now := time.Now()
	endDate := now.AddDate(0, 0, days)

	hasDueDate := true
	filter := &domain.NoteFilter{
		DueDateFrom: &now,
		DueDateTo:   &endDate,
		HasDueDate:  &hasDueDate,
	}

	return s.repo.GetAll(ctx, filter)
}

// GetOverdue возвращает просроченные невыполненные заметки.
//
// Параметры:
//   - ctx: контекст для отмены операции
//
// Возвращает:
//   - []*domain.Note: список просроченных заметок
//   - error: ошибка при запросе к БД
//
// Пример:
//
//	overdue, err := svc.GetOverdue(ctx)
//	if len(overdue) > 0 {
//	    // есть просроченные задачи
//	}
func (s *NoteService) GetOverdue(ctx context.Context) ([]*domain.Note, error) {
	now := time.Now()
	hasDueDate := true
	
	filter := &domain.NoteFilter{
		DueDateTo:  &now,
		HasDueDate: &hasDueDate,
	}

	notes, err := s.repo.GetAll(ctx, filter)
	if err != nil {
		return nil, err
	}

	// Фильтруем только невыполненные
	var overdue []*domain.Note
	for _, note := range notes {
		if !note.Completed {
			overdue = append(overdue, note)
		}
	}

	return overdue, nil
}
