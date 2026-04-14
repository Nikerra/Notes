FROM golang:1.21-alpine AS builder

WORKDIR /app

# Копируем go.mod для кэширования зависимостей
COPY server/go.mod server/go.sum ./
RUN go mod download

# Копируем исходный код
COPY server/ ./
RUN CGO_ENABLED=1 GOOS=linux go build -o /bin/notes-server cmd/server/main.go

# Финальный образ
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite

WORKDIR /app

# Копируем бинарник
COPY --from=builder /bin/notes-server /app/
COPY web/ /app/web/

# Создаем директорию для данных
RUN mkdir -p /app/data

EXPOSE 8080

ENV PORT=8080
ENV DATA_DIR=/app/data
ENV GIN_MODE=release

CMD ["/app/notes-server"]
