# Правила разработки ALFAGEN

## Архитектура

### Обязательная трехслойная архитектура

Проекты должны использовать четкое разделение на три слоя:

```
internal/
├── domain/       # Доменный слой
│   ├── models.go
│   └── repository.go
├── service/      # Бизнес-логика
│   └── service.go
├── api/          # HTTP обработчики
│   └── handler.go
└── repository/   # Реализация репозитория
    └── sqlite.go
```

**Domain (Доменный слой)**:
- Сущности и модели данных
- Интерфейсы репозиториев
- Бизнес-правила предметной области
- Не зависит от внешних библиотек

**Service (Слой бизнес-логики)**:
- Реализация бизнес-логики
- Валидация данных
- Координация между репозиториями
- Возвращает domain-типы и ошибки

**API (Слой представления)**:
- HTTP обработчики
- Сериализация/десериализация JSON
- Валидация входящих запросов
- Преобразование ошибок в HTTP статусы

**Repository (Слой данных)**:
- Реализация интерфейсов domain
- SQL запросы
- Миграции БД
- Только repository зависит от конкретной БД

### Зависимости

```
API → Service → Domain ← Repository
```

- Зависимости направлены внутрь (к domain)
- Domain не знает ни о service, ни о api, ни о repository
- Service зависит только от domain interfaces

## Документация кода

### GoDoc комментарии

Все экспортируемые сущности должны иметь комментарии:

```go
// Note represents a user note with title, content and completion status.
// Notes can be categorized as work or personal.
type Note struct {
    ID        int64     // Unique identifier
    Title     string    // Note title (required, max 100 chars)
    Content   string    // Note content (optional, max 500 chars)
    Category  string    // Category: "work" or "personal"
    Completed bool      // Whether the note is completed
    CreatedAt time.Time // Creation timestamp
    UpdatedAt time.Time // Last update timestamp
}

// Create creates a new note with the given title, content and category.
// Returns ErrInvalidTitle if title is empty.
// Returns ErrInvalidCategory if category is not valid.
//
// Example:
//   note, err := service.Create(ctx, "Buy groceries", "Milk, bread", domain.CategoryPersonal)
func (s *NoteService) Create(ctx context.Context, title, content string, category domain.Category) (*domain.Note, error) {
```

### Требования к комментариям

1. **Пакеты**: Комментарий перед `package` описывает назначение пакета
2. **Типы**: Описание структуры и её полей
3. **Функции**: Описание что делает, параметры, возвращаемые значения, ошибки, примеры
4. **Константы/переменные**: Пояснение назначения

## База данных

### Использование БД вместо файлов

- Всегда использовать БД для хранения данных (SQLite, PostgreSQL, MySQL)
- Не использовать файловую систему для персистентности данных
- Файлы только для конфигурации, логов, временных данных

### Миграции

```go
// Migrate creates the notes table if it doesn't exist.
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
```

## Тестирование

### Обязательное покрытие тестами

1. **Unit тесты** для каждого слоя
2. **Интеграционные тесты** для API endpoints
3. **Использовать стандартный пакет** `testing`
4. **Для asserts** можно использовать `testify/assert`

### Структура тестов

```
tests/
├── repository_test.go  # Тесты репозитория
├── service_test.go     # Тесты бизнес-логики
└── api_test.go         # Тесты HTTP handlers
```

### Именование тестов

```go
func TestCreateNote_Success(t *testing.T) {}
func TestCreateNote_EmptyTitle(t *testing.T) {}
func TestCreateNote_InvalidCategory(t *testing.T) {}
```

### Mock объекты

```go
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
```

## Нативные мобильные приложения

### Выбор технологий

Для нативных мобильных приложений использовать:

**iOS**:
- Swift 5+
- SwiftUI для UI
- URLSession для API
- CoreData/SwiftData для локального хранения

**Android**:
- Kotlin
- Jetpack Compose для UI
- Retrofit для API
- Room для локальной БД

### Общий Backend

Мобильные приложения используют тот же REST API что и веб:
- Единая точка синхронизации
- Один источник истины
- JWT/OAuth для аутентификации

## Обработка ошибок

### Определение ошибок

```go
var (
    ErrNoteNotFound    = errors.New("note not found")
    ErrInvalidTitle    = errors.New("title is required")
    ErrInvalidCategory = errors.New("invalid category")
    ErrInvalidNoteID   = errors.New("invalid note id")
)
```

### Возврат ошибок из Service

```go
func (s *NoteService) Create(ctx context.Context, title string, ...) (*domain.Note, error) {
    if title == "" {
        return nil, ErrInvalidTitle
    }
    // ...
    return note, nil
}
```

### Преобразование в HTTP статусы (в API слое)

```go
func (h *NoteHandler) Create(c *gin.Context) {
    note, err := h.svc.Create(...)
    if err != nil {
        switch err {
        case service.ErrInvalidTitle:
            c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
        default:
            c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
        }
        return
    }
    c.JSON(http.StatusCreated, note)
}
```

## Контекст (Context)

- Всегда передавать `context.Context` первым параметром
- Использовать `ctx` для отмены операций и таймаутов
- Не создавать пустой контекст в сервисах (`context.Background()` только в main)

## Конфигурация

### Переменные окружения

```go
func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

port := getEnv("PORT", "8080")
dataDir := getEnv("DATA_DIR", "./data")
```

### Поддерживаемые переменные

- `PORT` - порт сервера
- `DATA_DIR` - директория для данных
- `GIN_MODE` - режим работы (debug/release)
- `DATABASE_URL` - строка подключения к БД (для PostgreSQL)
