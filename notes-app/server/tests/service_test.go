package tests

import (
	"context"
	"testing"

	"notes-app/internal/domain"
	"notes-app/internal/service"
)

type mockRepository struct {
	notes  map[int64]*domain.Note
	nextID int64
}

func newMockRepository() *mockRepository {
	return &mockRepository{
		notes:  make(map[int64]*domain.Note),
		nextID: 1,
	}
}

func (m *mockRepository) Create(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	note.ID = m.nextID
	m.notes[m.nextID] = note
	m.nextID++
	return note, nil
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*domain.Note, error) {
	if note, ok := m.notes[id]; ok {
		return note, nil
	}
	return nil, domain.ErrNoteNotFound
}

func (m *mockRepository) GetAll(ctx context.Context, filter *domain.NoteFilter) ([]*domain.Note, error) {
	var notes []*domain.Note
	for _, note := range m.notes {
		if filter != nil && filter.Category != "" && note.Category != string(filter.Category) {
			continue
		}
		notes = append(notes, note)
	}
	return notes, nil
}

func (m *mockRepository) Update(ctx context.Context, note *domain.Note) (*domain.Note, error) {
	if _, ok := m.notes[note.ID]; !ok {
		return nil, domain.ErrNoteNotFound
	}
	m.notes[note.ID] = note
	return note, nil
}

func (m *mockRepository) Delete(ctx context.Context, id int64) error {
	if _, ok := m.notes[id]; !ok {
		return domain.ErrNoteNotFound
	}
	delete(m.notes, id)
	return nil
}

func TestCreateNote_Success(t *testing.T) {
	svc := service.NewNoteService(newMockRepository())
	ctx := context.Background()

	note, err := svc.Create(ctx, "Test Note", "Content", domain.CategoryPersonal)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if note.Title != "Test Note" {
		t.Errorf("Expected title 'Test Note', got %s", note.Title)
	}
	if note.Completed {
		t.Error("Expected completed to be false by default")
	}
}

func TestCreateNote_EmptyTitle(t *testing.T) {
	svc := service.NewNoteService(newMockRepository())
	ctx := context.Background()

	_, err := svc.Create(ctx, "", "Content", domain.CategoryPersonal)
	if err != service.ErrInvalidTitle {
		t.Errorf("Expected ErrInvalidTitle, got %v", err)
	}
}

func TestCreateNote_InvalidCategory(t *testing.T) {
	svc := service.NewNoteService(newMockRepository())
	ctx := context.Background()

	_, err := svc.Create(ctx, "Test", "Content", "invalid")
	if err != service.ErrInvalidCategory {
		t.Errorf("Expected ErrInvalidCategory, got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	mock := newMockRepository()
	svc := service.NewNoteService(mock)
	ctx := context.Background()

	created, _ := mock.Create(ctx, &domain.Note{Title: "Test", Category: "personal"})

	found, err := svc.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := service.NewNoteService(newMockRepository())
	ctx := context.Background()

	_, err := svc.GetByID(ctx, 999)
	if err != service.ErrNoteNotFound {
		t.Errorf("Expected ErrNoteNotFound, got %v", err)
	}
}

func TestUpdate_Success(t *testing.T) {
	mock := newMockRepository()
	svc := service.NewNoteService(mock)
	ctx := context.Background()

	created, _ := mock.Create(ctx, &domain.Note{Title: "Original", Category: "personal"})

	updated, err := svc.Update(ctx, created.ID, "Updated", "Content", domain.CategoryWork, true)
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updated.Title != "Updated" {
		t.Errorf("Expected title 'Updated', got %s", updated.Title)
	}
	if updated.Category != "work" {
		t.Errorf("Expected category 'work', got %s", updated.Category)
	}
	if !updated.Completed {
		t.Error("Expected completed to be true")
	}
}

func TestToggleComplete_Success(t *testing.T) {
	mock := newMockRepository()
	svc := service.NewNoteService(mock)
	ctx := context.Background()

	created, _ := mock.Create(ctx, &domain.Note{Title: "Test", Category: "personal", Completed: false})

	toggled, err := svc.ToggleComplete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to toggle note: %v", err)
	}

	if !toggled.Completed {
		t.Error("Expected completed to be true after toggle")
	}

	toggledAgain, _ := svc.ToggleComplete(ctx, created.ID)
	if toggledAgain.Completed {
		t.Error("Expected completed to be false after second toggle")
	}
}

func TestDelete_Success(t *testing.T) {
	mock := newMockRepository()
	svc := service.NewNoteService(mock)
	ctx := context.Background()

	created, _ := mock.Create(ctx, &domain.Note{Title: "Test", Category: "personal"})

	if err := svc.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	_, err := svc.GetByID(ctx, created.ID)
	if err != service.ErrNoteNotFound {
		t.Error("Expected note to be deleted")
	}
}
