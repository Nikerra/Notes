# Деплой на Linux сервер

## Вариант 1: Автоматическая установка (рекомендуется)

```bash
curl -fsSL https://raw.githubusercontent.com/Nikerra/Notes/main/notes-app/deploy/install.sh | bash
```

Эта команда:
- Установит Go (если не установлен)
- Склонирует репозиторий
- Соберет приложение
- Настроит systemd сервис
- Запустит приложение

Приложение будет доступно на `http://YOUR_SERVER_IP:8080`

## Вариант 2: Docker (самый простой)

```bash
# Клонируем репозиторий
git clone https://github.com/Nikerra/Notes.git
cd Notes/notes-app

# Запускаем через Docker Compose
docker-compose up -d

# Проверяем статус
docker-compose ps
```

Приложение будет доступно на `http://YOUR_SERVER_IP:8080`

## Вариант 3: Быстрый запуск (для тестирования)

```bash
curl -fsSL https://raw.githubusercontent.com/Nikerra/Notes/main/notes-app/deploy/quick-start.sh | bash
```

Запускает приложение без установки как сервис.

## Управление сервисом

После установки через Вариант 1:

```bash
# Статус
sudo systemctl status notes-app

# Перезапуск
sudo systemctl restart notes-app

# Остановка
sudo systemctl stop notes-app

# Логи
sudo journalctl -u notes-app -f
```

## Обновление

```bash
cd ~/notes-app
git pull
cd server
go build -o ../bin/notes-server cmd/server/main.go
sudo systemctl restart notes-app
```

## Настройка домена (опционально)

### С Nginx

```bash
sudo apt install nginx

sudo tee /etc/nginx/sites-available/notes << EOF
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade \$http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host \$host;
        proxy_cache_bypass \$http_upgrade;
    }
}
EOF

sudo ln -s /etc/nginx/sites-available/notes /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

### SSL с Certbot

```bash
sudo apt install certbot python3-certbot-nginx
sudo certbot --nginx -d your-domain.com
```

## Переменные окружения

Создайте файл `/etc/notes-app.env`:

```bash
PORT=8080
DATA_DIR=/var/lib/notes
GIN_MODE=release
```

Обновите systemd сервис:

```ini
[Service]
EnvironmentFile=/etc/notes-app.env
```

## Минимальные требования

- OS: Linux (Ubuntu/Debian/CentOS)
- RAM: 256 MB
- Disk: 100 MB
- Go: 1.21+ (для Варианта 1 и 3)
- Docker: любой версии (для Варианта 2)

## Безопасность

1. **Firewall**:
```bash
sudo ufw allow 8080/tcp
sudo ufw enable
```

2. **Ограничение доступа** (опционально):
```bash
# Только с определенных IP
sudo ufw allow from 192.168.1.0/24 to any port 8080
```

3. **HTTPS** через Nginx + Certbot (рекомендуется)
