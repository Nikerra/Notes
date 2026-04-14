# Notes App - Кроссплатформенное приложение заметок

Легковесное приложение для заметок с трехслойной архитектурой, unit-тестами и поддержкой Android, iOS и веб-браузера.

## Быстрый старт

### Локальный запуск (macOS/Linux)

```bash
cd notes-app/server
WEB_DIR=../web go run cmd/server/main.go
```

Приложение доступно: http://localhost:8080

### Деплой на Linux сервер (одна команда)

```bash
curl -fsSL https://raw.githubusercontent.com/Nikerra/Notes/main/notes-app/deploy/install.sh | bash
```

Автоматически установит Go, соберет и запустит приложение как systemd сервис.

### Docker (рекомендуется для продакшена)

```bash
git clone https://github.com/Nikerra/Notes.git
cd Notes/notes-app
docker-compose up -d
```

Приложение доступно: http://YOUR_SERVER_IP:8080

Подробная инструкция: [DEPLOY.md](DEPLOY.md)

## Архитектура

Проект использует чистую трехслойную архитектуру:

```
server/
├── cmd/
│   └── server/
│       └── main.go              # Точка входа
├── internal/
│   ├── domain/                  # Доменный слой
│   │   ├── note.go             # Сущности и модели
│   │   └── repository.go       # Интерфейсы репозиториев
│   ├── service/                 # Бизнес-логика
│   │   └── note.go             # Сервис заметок
│   ├── api/                     # HTTP слой
│   │   └── handler.go          # Обработчики запросов
│   └── repository/              # Слой данных
│       └── sqlite.go           # Реализация SQLite
├── tests/
│   ├── repository_test.go      # Тесты репозитория
│   └── service_test.go         # Тесты сервиса
└── go.mod
```

### Слои

**Domain** - ядро приложения:
- Сущности (`Note`, `Category`)
- Интерфейсы репозиториев
- Не имеет внешних зависимостей

**Service** - бизнес-логика:
- Создание, обновление, удаление заметок
- Валидация данных
- Возвращает domain-типы и ошибки

**API** - HTTP интерфейс:
- REST endpoints
- Преобразование JSON
- Обработка HTTP статусов

**Repository** - хранение данных:
- SQLite реализация
- SQL запросы
- Миграции БД

## Технологии

**Backend**:
- Go 1.21+
- Gin (HTTP framework)
- SQLite (база данных)

**Frontend**:
- Vanilla JavaScript
- Progressive Web App (PWA)
- Service Worker для офлайн

**Тестирование**:
- Go `testing` package
- In-memory SQLite для тестов

## API Endpoints

| Method | Endpoint | Описание |
|--------|----------|----------|
| GET | /api/notes | Получить все заметки |
| GET | /api/notes?category=work | Фильтр по категории |
| POST | /api/notes | Создать заметку |
| PUT | /api/notes/:id | Обновить заметку |
| PATCH | /api/notes/:id/toggle | Переключить статус |
| DELETE | /api/notes/:id | Удалить заметку |

### Примеры запросов

**Создать заметку**:
```bash
curl -X POST http://localhost:8080/api/notes \
  -H "Content-Type: application/json" \
  -d '{"title":"Купить продукты","content":"Молоко, хлеб","category":"personal"}'
```

**Получить заметки работы**:
```bash
curl http://localhost:8080/api/notes?category=work
```

## Запуск

### Разработка

```bash
cd notes-app
go run server/cmd/server/main.go
```

Сервер запустится на `http://localhost:8080`

### Переменные окружения

| Переменная | По умолчанию | Описание |
|------------|--------------|----------|
| PORT | 8080 | Порт сервера |
| DATA_DIR | ./data | Директория для БД |
| GIN_MODE | debug | Режим (debug/release) |

### Production

```bash
cd notes-app/server
go build -o bin/server cmd/server/main.go
DATA_DIR=/var/lib/notes GIN_MODE=release ./bin/server
```

## Тестирование

### Запуск всех тестов

```bash
cd notes-app/server
go test ./tests/... -v
```

### Запуск с покрытием

```bash
go test ./tests/... -cover
```

### Структура тестов

- `repository_test.go` - тесты слоя данных (SQLite)
- `service_test.go` - тесты бизнес-логики (с mock)

## Установка на мобильные устройства

### Android

1. Откройте `http://localhost:8080` в Chrome
2. Меню → "Добавить на главный экран"

### iOS

1. Откройте `http://localhost:8080` в Safari
2. Поделиться → "На экран Домой"

## PWA Features

- Устанавливается как приложение
- Работает офлайн (localStorage + Service Worker)
- Автоматическая синхронизация
- Нативный вид и поведение

## Нативные мобильные приложения

Для полноценных нативных приложений рекомендуется:

**iOS (Swift + SwiftUI)**:
- URLSession для API
- SwiftData/CoreData для офлайн
- SwiftUI для UI

**Android (Kotlin + Compose)**:
- Retrofit для API
- Room для локальной БД
- Jetpack Compose для UI

Backend API остается тем же самым - единая точка синхронизации.

## Разработка

### Добавление новой функциональности

1. **Domain**: Добавить сущность/интерфейс в `internal/domain/`
2. **Repository**: Реализовать методы в `internal/repository/`
3. **Service**: Добавить бизнес-логику в `internal/service/`
4. **API**: Создать handler в `internal/api/`
5. **Tests**: Написать тесты для каждого слоя

### Code Style

- GoDoc комментарии для всех экспортируемых сущностей
- Контекст первым параметром в функциях
- Возврат ошибок через `domain.ErrXXX`
- Валидация в service слое

## Структура проекта

```
notes-app/
├── server/                      # Go backend
│   ├── cmd/server/             # Точка входа
│   ├── internal/               # Внутренние пакеты
│   │   ├── domain/            # Доменные модели
│   │   ├── service/           # Бизнес-логика
│   │   ├── api/               # HTTP handlers
│   │   └── repository/        # Реализация БД
│   ├── tests/                  # Unit тесты
│   ├── go.mod                  # Зависимости
│   └── go.sum
├── web/                        # PWA frontend
│   ├── index.html             # Главная страница
│   ├── app.js                 # Логика приложения
│   ├── styles.css             # Стили
│   ├── manifest.json          # PWA манифест
│   ├── sw.js                  # Service Worker
│   └── icon-*.png             # Иконки
├── kilo.json                   # Kilo конфигурация
├── RULES.md                    # Правила разработки
└── README.md                   # Документация
```

## Лицензия

MIT

## Документация кода

### Просмотр документации

Все пакеты имеют полную GoDoc документацию:

```bash
# Документация домена
go doc ./internal/domain

# Документация сервиса
go doc ./internal/service

# Документация конкретного метода
go doc ./internal/service NoteService.Create

# Документация API
go doc ./internal/api NoteHandler

# Документация репозитория
go doc ./internal/repository SQLiteRepository
```

### Пример документации метода

```go
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
func (s *NoteService) Create(ctx context.Context, title, content string, category domain.Category) (*domain.Note, error)
```

### Структура документации

Каждый экспортируемый элемент должен иметь:

1. **Пакеты**: Описание назначения пакета
2. **Типы**: Описание структуры и всех полей
3. **Методы**: 
   - Краткое описание
   - Параметры с типами и описанием
   - Возвращаемые значения
   - Возможные ошибки
   - Пример использования

