#!/bin/bash

# Быстрое развертывание одной командой
# Запуск: curl -fsSL https://raw.githubusercontent.com/Nikerra/Notes/main/notes-app/deploy/quick-start.sh | bash

set -e

echo "🚀 Быстрый запуск Notes App..."

# Клонирование и запуск
if [ ! -d "$HOME/notes-app" ]; then
    git clone https://github.com/Nikerra/Notes.git "$HOME/notes-app"
fi

cd "$HOME/notes-app/server"

# Проверка Go
if ! command -v go &> /dev/null; then
    echo "❌ Go не установлен. Установите Go: https://go.dev/dl/"
    exit 1
fi

# Установка зависимостей и запуск
go mod download
go run cmd/server/main.go
