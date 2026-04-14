package tests

import (
	"context"
	"database/sql"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"notes-app/internal/domain"
	"notes-app/internal/repository"
)

func setupTestDB(t *testing.T) *repository.SQLiteRepository {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("Failed to open test database: %v", err)
	}

	repo := repository.NewSQLiteRepository(db)
	if err := repo.Migrate(); err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return repo
}

func TestCreateNote(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	note := &domain.Note{
		Title:    "Test Note",
		Content:  "Test Content",
		Category: "personal",
	}

	created, err := repo.Create(ctx, note)
	if err != nil {
		t.Fatalf("Failed to create note: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected note ID to be set")
	}
	if created.Title != note.Title {
		t.Errorf("Expected title %s, got %s", note.Title, created.Title)
	}
}

func TestGetByID(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	note := &domain.Note{
		Title:    "Test Note",
		Content:  "Test Content",
		Category: "work",
	}

	created, _ := repo.Create(ctx, note)

	found, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("Failed to get note: %v", err)
	}

	if found.ID != created.ID {
		t.Errorf("Expected ID %d, got %d", created.ID, found.ID)
	}
}

func TestGetAll(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	for i := 1; i <= 3; i++ {
		note := &domain.Note{
			Title:    "Test Note",
			Category: "personal",
		}
		repo.Create(ctx, note)
	}

	notes, err := repo.GetAll(ctx, nil)
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) != 3 {
		t.Errorf("Expected 3 notes, got %d", len(notes))
	}
}

func TestGetAllWithFilter(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	personal := &domain.Note{Title: "Personal", Category: "personal"}
	work := &domain.Note{Title: "Work", Category: "work"}
	repo.Create(ctx, personal)
	repo.Create(ctx, work)

	filter := &domain.NoteFilter{Category: domain.CategoryWork}
	notes, err := repo.GetAll(ctx, filter)
	if err != nil {
		t.Fatalf("Failed to get notes: %v", err)
	}

	if len(notes) != 1 {
		t.Errorf("Expected 1 note, got %d", len(notes))
	}
	if notes[0].Category != "work" {
		t.Errorf("Expected work category, got %s", notes[0].Category)
	}
}

func TestUpdate(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	note := &domain.Note{
		Title:    "Original",
		Category: "personal",
	}
	created, _ := repo.Create(ctx, note)

	created.Title = "Updated"
	created.Completed = true
	updated, err := repo.Update(ctx, created)
	if err != nil {
		t.Fatalf("Failed to update note: %v", err)
	}

	if updated.Title != "Updated" {
		t.Errorf("Expected title Updated, got %s", updated.Title)
	}
	if !updated.Completed {
		t.Error("Expected completed to be true")
	}
}

func TestDelete(t *testing.T) {
	repo := setupTestDB(t)
	ctx := context.Background()

	note := &domain.Note{
		Title:    "To Delete",
		Category: "personal",
	}
	created, _ := repo.Create(ctx, note)

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Failed to delete note: %v", err)
	}

	_, err := repo.GetByID(ctx, created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted note")
	}
}
