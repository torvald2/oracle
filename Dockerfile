
# --- Этап сборки ---
FROM golang:1.24-alpine AS builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем go.mod и go.sum для кэширования зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копируем остальной исходный код
COPY . .

# Собираем приложение.
# -o /app/main создает бинарный файл с именем 'main'
# CGO_ENABLED=0 и -ldflags "-s -w" делают бинарник статическим и уменьшают его размер
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/main .

# --- Финальный этап ---
FROM alpine:latest

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем бинарный файл из этапа сборки
COPY --from=builder /app/main .

# Копируем SQL-скрипт, хотя он используется только сервисом db,
# это может быть полезно для отладки внутри контейнера.
# Этот шаг опционален.
COPY ./db/init/init.sql .

# Открываем порт, который слушает наше приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
