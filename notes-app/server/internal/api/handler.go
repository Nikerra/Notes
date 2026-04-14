// Package api содержит HTTP обработчики для REST API.
// API слой отвечает за сериализацию/десериализацию JSON и преобразование ошибок в HTTP статусы.
package api

import (
	"net/http"
	"strconv"

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
		notes.GET("/:id", h.GetByID)
		notes.PUT("/:id", h.Update)
		notes.DELETE("/:id", h.Delete)
		notes.PATCH("/:id/toggle", h.ToggleComplete)
	}
}

// GetAll обрабатывает GET /api/notes запрос.
// Возвращает список всех заметок с возможностью фильтрации по категории.
//
// Query параметры:
//   - category: категория для фильтрации ("work", "personal" или "all")
//
// Ответы:
//   - 200 OK: массив заметок в JSON
//   - 500 Internal Server Error: ошибка сервера
//
// Пример:
//
//	GET /api/notes
//	GET /api/notes?category=work
func (h *NoteHandler) GetAll(c *gin.Context) {
	category := c.Query("category")

	var filter *domain.NoteFilter
	if category != "" && category != "all" {
		filter = &domain.NoteFilter{Category: domain.Category(category)}
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
//	    "category": "personal"
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

	note, err := h.svc.Create(c.Request.Context(), req.Title, req.Content, category)
	if err != nil {
		if err == service.ErrInvalidTitle || err == service.ErrInvalidCategory {
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
//	    "completed": true
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
	note, err := h.svc.Update(c.Request.Context(), id, req.Title, req.Content, category, req.Completed)
	if err != nil {
		if err == service.ErrNoteNotFound {
			c.JSON(http.StatusNotFound, ErrorResponse{Error: err.Error()})
			return
		}
		if err == service.ErrInvalidTitle || err == service.ErrInvalidCategory {
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
