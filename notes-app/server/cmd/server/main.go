// Package main является точкой входа в серверное приложение заметок.
// Инициализирует все компоненты приложения и запускает HTTP сервер.
//
// Запуск:
//
//	go run cmd/server/main.go
//
// Переменные окружения:
//   - PORT: порт сервера (по умолчанию 8080)
//   - DATA_DIR: директория для хранения данных (по умолчанию ./data)
//   - GIN_MODE: режим работы Gin - "debug" или "release" (по умолчанию debug)
//
// Пример запуска с кастомными параметрами:
//
//	PORT=3000 DATA_DIR=/var/lib/notes GIN_MODE=release go run cmd/server/main.go
package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"

	"notes-app/internal/api"
	"notes-app/internal/repository"
	"notes-app/internal/service"
)

// main является точкой входа приложения.
// Выполняет следующие шаги:
//  1. Загружает конфигурацию из переменных окружения
//  2. Создает директорию для данных
//  3. Подключается к SQLite базе данных
//  4. Запускает миграции БД
//  5. Инициализирует слои приложения (repository -> service -> api)
//  6. Настраивает маршруты HTTP
//  7. Запускает HTTP сервер
func main() {
	// Загрузка конфигурации из переменных окружения
	dataDir := getEnv("DATA_DIR", "./data")
	port := getEnv("PORT", "8080")
	ginMode := getEnv("GIN_MODE", "debug")

	// Автоматическое определение пути к web директории
	webDir := getEnv("WEB_DIR", "")
	if webDir == "" {
		// Проверяем несколько возможных путей
		possiblePaths := []string{"./web", "../web", "../../web"}
		for _, path := range possiblePaths {
			if _, err := os.Stat(path); err == nil {
				webDir = path
				break
			}
		}
		if webDir == "" {
			log.Fatal("Web directory not found. Set WEB_DIR environment variable")
		}
	}

	// Установка режима Gin (debug или release)
	gin.SetMode(ginMode)

	// Создание директории для хранения данных
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		log.Fatalf("Failed to create data directory: %v", err)
	}

	// Подключение к SQLite базе данных
	db, err := sql.Open("sqlite3", dataDir+"/notes.db")
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	// Инициализация репозитория и выполнение миграций
	repo := repository.NewSQLiteRepository(db)
	if err := repo.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	// Инициализация слоев приложения
	// Зависимости: API -> Service -> Repository
	noteService := service.NewNoteService(repo)
	noteHandler := api.NewNoteHandler(noteService)

	// Настройка HTTP роутера
	router := gin.Default()

	// Настройка CORS для кросс-доменных запросов
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},                                                // Разрешить все источники
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"}, // Разрешенные методы
		AllowHeaders:     []string{"Origin", "Content-Type"},                           // Разрешенные заголовки
		AllowCredentials: true,                                                         // Разрешить credentials
	}))

	// Регистрация API маршрутов
	apiGroup := router.Group("/api")
	noteHandler.RegisterRoutes(apiGroup)

	// Настройка статических файлов для PWA
	router.Static("/static", webDir)
	router.StaticFile("/", webDir+"/index.html")

	// Запуск HTTP сервера
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// getEnv возвращает значение переменной окружения или значение по умолчанию.
//
// Параметры:
//   - key: имя переменной окружения
//   - defaultValue: значение по умолчанию, если переменная не установлена
//
// Возвращает:
//   - string: значение переменной окружения или defaultValue
//
// Пример:
//
//	port := getEnv("PORT", "8080") // вернет "8080" если PORT не установлен
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
