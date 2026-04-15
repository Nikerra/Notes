// Package api содержит HTTP обработчики для REST API.
// API слой отвечает за сериализацию/десериализацию JSON и преобразование ошибок в HTTP статусы.
package api

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"notes-app/internal/domain"
	"notes-app/internal/service"
)

// NoteHandler обрабатывает HTTP запросы для работы с заметками.
// Является адаптером между HTTP и сервисным слоем.
type NoteHandler struct {
	svc *service.NoteService
}

// NewNoteHandler создает новый экземпляр NoteHandler.
//
// Параметры:
//   - svc: сервис для работы с заметками
//
// Пример:
//
//	svc := service.NewNoteService(repo)
//	handler := api.NewNoteHandler(svc)
func NewNoteHandler(svc *service.NoteService) *NoteHandler {
	return &NoteHandler{svc: svc}
}

// CreateNoteRequest представляет запрос на создание заметки.
type CreateNoteRequest struct {
	// Title - заголовок заметки (обязательное поле)
	Title string `json:"title" binding:"required"`

	// Content - содержание заметки (опциональное)
	Content string `json:"content"`

	// Category - категория заметки ("work" или "personal")
	Category string `json:"category"`

	// Completed - статус выполнения (по умолчанию false)
	Completed bool `json:"completed"`

	// DueDate - срок выполнения в формате RFC3339 (опциональное)
	// Пример: "2024-12-31T23:59:59Z"
	DueDate *time.Time `json:"dueDate"`
}

// UpdateNoteRequest представляет запрос на обновление заметки.
type UpdateNoteRequest struct {
	// Title - новый заголовок (обязательное поле)
	Title string `json:"title" binding:"required"`

	// Content - новое содержание
	Content string `json:"content"`

	// Category - новая категория
	Category string `json:"category"`

	// Completed - новый статус выполнения
	Completed bool `json:"completed"`

	// DueDate - новый срок выполнения (nil для удаления срока)
	DueDate *time.Time `json:"dueDate"`
}

// ErrorResponse представляет ошибку в HTTP ответе.
type ErrorResponse struct {
	// Error - текст ошибки
	Error string `json:"error"`
}

// RegisterRoutes регистрирует маршруты для работы с заметками.
// Создает следующие endpoints:
//   - GET /notes - получить все заметки
//   - POST /notes - создать заметку
//   - GET /notes/:id - получить заметку по ID
//   - PUT /notes/:id - обновить заметку
//   - DELETE /notes/:id - удалить заметку
//   - PATCH /notes/:id/toggle - переключить статус выполнения
//   - GET /notes/upcoming - получить предстоящие заметки
//   - GET /notes/overdue - получить просроченные заметки
//
// Параметры:
//   - r: группа маршрутов Gin
//
// Пример:
//
//	router := gin.Default()
//	api := router.Group("/api")
//	handler.RegisterRoutes(api)
func (h *NoteHandler) RegisterRoutes(r *gin.RouterGroup) {
	notes := r.Group("/notes")
	{
		notes.GET("", h.GetAll)
		notes.POST("", h.Create)
		notes.GET("/upcoming", h.GetUpcoming)
		notes.GET("/overdue", h.GetOverdue)
		notes.GET("/:id", h.GetByID)
		notes.PUT("/:id", h.Update)
		notes.DELETE("/:id", h.Delete)
		notes.PATCH("/:id/toggle", h.ToggleComplete)
	}
}

// GetAll обрабатывает GET /api/notes запрос.
// Возвращает список всех заметок с возможностью фильтрации.
//
// Query параметры:
//   - category: категория для фильтрации ("work", "personal" или "all")
//   - dueDateFrom: начальная дата срока выполнения (RFC3339)
//   - dueDateTo: конечная дата срока выполнения (RFC3339)
//   - hasDueDate: "true" или "false" - фильтр по наличию срока
//
// Ответы:
//   - 200 OK: массив заметок в JSON
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	GET /api/notes
//	GET /api/notes?category=work
//	GET /api/notes?dueDateFrom=2024-01-01T00:00:00Z&dueDateTo=2024-12-31T23:59:59Z
func (h *NoteHandler) GetAll(c *gin.Context) {
	filter := &domain.NoteFilter{}

	category := c.Query("category")
	if category != "" && category != "all" {
		filter.Category = domain.Category(category)
	}

	if dueDateFrom := c.Query("dueDateFrom"); dueDateFrom != "" {
		if t, err := time.Parse(time.RFC3339, dueDateFrom); err == nil {
			filter.DueDateFrom = &t
		}
	}

	if dueDateTo := c.Query("dueDateTo"); dueDateTo != "" {
		if t, err := time.Parse(time.RFC3339, dueDateTo); err == nil {
			filter.DueDateTo = &t
		}
	}

	if hasDueDate := c.Query("hasDueDate"); hasDueDate != "" {
		val := hasDueDate == "true"
		filter.HasDueDate = &val
	}

	notes, err := h.svc.GetAll(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

// GetByID обрабатывает GET /api/notes/:id запрос.
// Возвращает заметку по указанному идентификатору.
//
// URL параметры:
//   - id: идентификатор заметки
//
// Ответы:
//   - 200 OK: заметка в JSON
//   - 400 Bad Request: невалидный ID
//   - 404 Not Found: заметка не найдена
//
// Пример:
//
//	GET /api/notes/1
func (h *NoteHandler) GetByID(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	note, err := h.svc.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

// Create обрабатывает POST /api/notes запрос.
// Создает новую заметку с указанными данными.
//
// Тело запроса:
//   - title (обязательное): заголовок заметки
//   - content (опциональное): содержание
//   - category (опциональное): "work" или "personal" (по умолчанию "personal")
//   - dueDate (опциональное): срок выполнения в RFC3339
//
// Ответы:
//   - 201 Created: созданная заметка в JSON
//   - 400 Bad Request: невалидные данные
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	POST /api/notes
//	{
//	    "title": "Купить продукты",
//	    "content": "Молоко, хлеб",
//	    "category": "personal",
//	    "dueDate": "2024-12-31T18:00:00Z"
//	}
func (h *NoteHandler) Create(c *gin.Context) {
	var req CreateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	category := domain.Category(req.Category)
	if category == "" {
		category = domain.CategoryPersonal
	}

	note, err := h.svc.Create(c.Request.Context(), req.Title, req.Content, category, req.DueDate)
	if err != nil {
		if err == service.ErrInvalidTitle || err == service.ErrInvalidCategory || err == service.ErrInvalidDueDate {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusCreated, note)
}

// Update обрабатывает PUT /api/notes/:id запрос.
// Обновляет существующую заметку с указанными данными.
//
// URL параметры:
//   - id: идентификатор заметки
//
// Тело запроса:
//   - title (обязательное): новый заголовок
//   - content (опциональное): новое содержание
//   - category (опциональное): новая категория
//   - completed: новый статус выполнения
//   - dueDate (опциональное): новый срок выполнения (null для удаления)
//
// Ответы:
//   - 200 OK: обновленная заметка в JSON
//   - 400 Bad Request: невалидные данные
//   - 404 Not Found: заметка не найдена
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	PUT /api/notes/1
//	{
//	    "title": "Обновленный заголовок",
//	    "content": "Обновленное содержание",
//	    "category": "work",
//	    "completed": true,
//	    "dueDate": null
//	}
func (h *NoteHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	var req UpdateNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
		return
	}

	category := domain.Category(req.Category)
	note, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Content, category, req.Completed, req.DueDate)
	if err != nil {
		if err == service.ErrNoteNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err == service.ErrInvalidTitle || err == service.ErrInvalidCategory || err == service.ErrInvalidDueDate {
			c.JSON(http.StatusBadRequest, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

// ToggleComplete обрабатывает PATCH /api/notes/:id/toggle запрос.
// Переключает статус выполнения заметки (выполнено <-> не выполнено).
//
// URL параметры:
//   - id: идентификатор заметки
//
// Ответы:
//   - 200 OK: обновленная заметка в JSON
//   - 400 Bad Request: невалидный ID
//   - 404 Not Found: заметка не найдена
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	PATCH /api/notes/1/toggle
func (h *NoteHandler) ToggleComplete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	note, err := h.svc.ToggleComplete(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrNoteNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, note)
}

// Delete обрабатывает DELETE /api/notes/:id запрос.
// Удаляет заметку по указанному идентификатору.
//
// URL параметры:
//   - id: идентификатор заметки
//
// Ответы:
//   - 200 OK: сообщение об успешном удалении
//   - 400 Bad Request: невалидный ID
//   - 404 Not Found: заметка не найдена
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	DELETE /api/notes/1
func (h *NoteHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse{Error: "invalid id"})
		return
	}

	if err := h.svc.Delete(c.Request.Context(), id); err != nil {
		if err == service.ErrNoteNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "note deleted"})
}

// GetUpcoming обрабатывает GET /api/notes/upcoming запрос.
// Возвращает заметки с предстоящим сроком выполнения.
//
// Query параметры:
//   - days: количество дней вперед (по умолчанию 7)
//
// Ответы:
//   - 200 OK: массив заметок в JSON
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	GET /api/notes/upcoming
//	GET /api/notes/upcoming?days=14
func (h *NoteHandler) GetUpcoming(c *gin.Context) {
	days := 7
	if d := c.Query("days"); d != "" {
		if parsed, err := strconv.Atoi(d); err == nil && parsed > 0 {
			days = parsed
		}
	}

	notes, err := h.svc.GetUpcoming(c.Request.Context(), days)
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}

// GetOverdue обрабатывает GET /api/notes/overdue запрос.
// Возвращает просроченные невыполненные заметки.
//
// Ответы:
//   - 200 OK: массив заметок в JSON
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	GET /api/notes/overdue
func (h *NoteHandler) GetOverdue(c *gin.Context) {
	notes, err := h.svc.GetOverdue(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
		return
	}

	c.JSON(http.StatusOK, notes)
}
