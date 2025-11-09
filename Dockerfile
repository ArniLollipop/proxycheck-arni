# --- Этап 1: Сборка фронтенда ---
FROM node:18-alpine AS frontend-builder

WORKDIR /app

# Копируем package.json и package-lock.json
COPY code/client/package*.json ./

# Устанавливаем зависимости
RUN npm install

# Копируем исходники фронтенда
COPY code/client/ ./

# Сборка фронтенда
RUN npm run build


# --- Этап 2: Сборка бэкенда ---
FROM golang:1.24-alpine AS backend-builder

# Устанавливаем C-компилятор для CGO и SQLite
RUN apk add --no-cache build-base

WORKDIR /app

# Копируем зависимости Go
COPY code/go.mod code/go.sum ./
RUN go mod download

# Копируем остальные исходники Go
COPY code/ ./

# Копируем собранный фронтенд
COPY --from=frontend-builder /app/dist ./client/dist

# Копируем необходимые файлы рантайма
COPY code/GeoIP2-ISP.mmdb ./

# Не копируем базу SQLite, она будет примонтирована с хоста

# Сборка Go-приложения
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o /app/main .


# --- Этап 3: Финальный образ ---
FROM alpine:latest

WORKDIR /app

# Устанавливаем runtime зависимости (если нужны)
RUN apk add --no-cache sqlite

# Копируем бинарник и фронтенд
COPY --from=backend-builder /app/main .
COPY --from=backend-builder /app/client/dist ./client/dist
COPY --from=backend-builder /app/GeoIP2-ISP.mmdb ./

# Открываем порт приложения
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
