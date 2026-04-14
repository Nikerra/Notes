#!/bin/bash

# Notes App - Автоматическая установка для Linux
# Запуск: curl -fsSL https://raw.githubusercontent.com/Nikerra/Notes/main/notes-app/deploy/install.sh | bash

set -e

echo "========================================="
echo "Notes App - Установка на Linux"
echo "========================================="

# Проверка Go
if ! command -v go &> /dev/null; then
    echo "Go не установлен. Установка Go..."
    wget -q https://go.dev/dl/go1.21.6.linux-amd64.tar.gz
    sudo tar -C /usr/local -xzf go1.21.6.linux-amd64.tar.gz
    rm go1.21.6.linux-amd64.tar.gz
    export PATH=$PATH:/usr/local/go/bin
    echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
    echo "Go установлен"
else
    echo "✓ Go уже установлен: $(go version)"
fi

# Создание директории
INSTALL_DIR="${HOME}/notes-app"
mkdir -p "$INSTALL_DIR"
cd "$INSTALL_DIR"

# Клонирование репозитория
if [ ! -d ".git" ]; then
    echo "Клонирование репозитория..."
    git clone https://github.com/Nikerra/Notes.git .
else
    echo "✓ Репозиторий уже клонирован"
fi

# Сборка сервера
echo "Сборка сервера..."
cd server
go mod download
go build -o ../bin/notes-server cmd/server/main.go

# Создание systemd сервиса
SERVICE_FILE="/tmp/notes-app.service"
cat > "$SERVICE_FILE" << 'EOL'
[Unit]
Description=Notes App Server
After=network.target

[Service]
Type=simple
User=%USER%
WorkingDirectory=%INSTALL_DIR%/notes-app
ExecStart=%INSTALL_DIR%/bin/notes-server
Restart=on-failure
RestartSec=10

Environment=PORT=8080
Environment=DATA_DIR=%INSTALL_DIR%/notes-app/data
Environment=GIN_MODE=release

[Install]
WantedBy=multi-user.target
EOL

# Подстановка переменных
sed -i "s|%USER%|$USER|g" "$SERVICE_FILE"
sed -i "s|%INSTALL_DIR%|$HOME|g" "$SERVICE_FILE"

echo "Установка systemd сервиса..."
sudo mv "$SERVICE_FILE" /etc/systemd/system/notes-app.service
sudo systemctl daemon-reload
sudo systemctl enable notes-app
sudo systemctl start notes-app

# Проверка статуса
sleep 2
if sudo systemctl is-active --quiet notes-app; then
    echo ""
    echo "========================================="
    echo "✓ Установка завершена успешно!"
    echo "========================================="
    echo ""
    echo "Приложение доступно: http://$(hostname -I | awk '{print $1}'):8080"
    echo ""
    echo "Команды управления:"
    echo "  sudo systemctl status notes-app   - статус"
    echo "  sudo systemctl restart notes-app  - перезапуск"
    echo "  sudo systemctl stop notes-app     - остановка"
    echo "  sudo journalctl -u notes-app -f   - логи"
    echo ""
else
    echo "❌ Ошибка запуска сервиса. Проверьте логи:"
    echo "  sudo journalctl -u notes-app -n 50"
fi
