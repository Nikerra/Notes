// Package domain содержит основные доменные сущности и интерфейсы приложения.
// Этот пакет не имеет внешних зависимостей и определяет ядро бизнес-логики.
package domain

import "time"

// Note представляет заметку пользователя с заголовком, содержанием и статусом выполнения.
// Заметки могут быть категоризированы как рабочие или личные, а также иметь срок выполнения.
type Note struct {
	// ID - уникальный идентификатор заметки (автоинкремент в БД)
	ID int64 `json:"id"`

	// Title - заголовок заметки (обязательное поле, макс. 100 символов)
	Title string `json:"title"`

	// Content - содержание заметки (опциональное поле, макс. 500 символов)
	Content string `json:"content"`

	// Category - категория заметки: "work" или "personal"
	Category string `json:"category"`

	// Completed - флаг выполнения заметки (true/false)
	Completed bool `json:"completed"`

	// DueDate - срок выполнения заметки (опциональное поле).
	// Используется для отображения в календаре и установки напоминаний.
	// Может быть nil, если срок не установлен.
	DueDate *time.Time `json:"dueDate,omitempty"`

	// CreatedAt - дата и время создания заметки
	CreatedAt time.Time `json:"createdAt"`

	// UpdatedAt - дата и время последнего обновления заметки
	UpdatedAt time.Time `json:"updatedAt"`
}

// Category представляет категорию заметки.
// Используется для фильтрации заметок по типу.
type Category string

// Предопределенные категории заметок.
const (
	// CategoryWork - рабочие заметки (задачи, встречи, проекты)
	CategoryWork Category = "work"

	// CategoryPersonal - личные заметки (покупки, дела, идеи)
	CategoryPersonal Category = "personal"
)

// IsValid проверяет, является ли категория допустимой.
// Возвращает true, если категория равна CategoryWork или CategoryPersonal.
//
// Пример:
//
//	category := domain.Category("work")
//	if category.IsValid() {
//	    // категория валидна
//	}
func (c Category) IsValid() bool {
	return c == CategoryWork || c == CategoryPersonal
}

// NoteFilter представляет фильтр для поиска заметок.
// Пустой фильтр возвращает все заметки.
type NoteFilter struct {
	// Category - категория для фильтрации (если пусто, возвращает все категории)
	Category Category

	// DueDateFrom - начальная дата для фильтрации по сроку выполнения
	DueDateFrom *time.Time

	// DueDateTo - конечная дата для фильтрации по сроку выполнения
	DueDateTo *time.Time

	// HasDueDate - фильтр: только заметки с установленным сроком (true) или без срока (false)
	// nil означает, что фильтр не применяется
	HasDueDate *bool
}
