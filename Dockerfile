
# --- Этап 1: Сборка фронтенда ---
FROM node:18-alpine AS frontend-builder

# Устанавливаем рабочую директорию
WORKDIR /app

# Копируем package.json и package-lock.json для установки зависимостей
COPY code/client/package*.json ./

# Устанавливаем зависимости npm
RUN npm install

# Копируем остальные исходники фронтенда
COPY code/client/ ./

# Собираем фронтенд. Результат будет в папке /app/dist
RUN npm run build


# --- Этап 2: Сборка бэкенда ---
FROM golang:1.24-alpine AS backend-builder

# Устанавливаем C-компилятор, необходимый для CGO и SQLite
RUN apk add --no-cache build-base

WORKDIR /app

# Копируем файлы зависимостей Go и загружаем их
COPY code/go.mod code/go.sum ./
RUN go mod download

# Копируем все остальные исходники бэкенда
COPY code/ ./

# Копируем собранный фронтенд из предыдущего этапа
COPY --from=frontend-builder /app/dist ./client/dist

# Копируем базу данных GeoIP
COPY code/GeoIP2-ISP.mmdb ./

# Копируем базу данных SQLite
COPY code/database ./database

# Собираем Go-приложение с включенным CGO
RUN CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-w -s' -o /app/main .


# --- Этап 3: Финальный образ ---
FROM alpine:latest

WORKDIR /app

# Копируем скомпилированный бинарник из стадии сборки бэкенда
COPY --from=backend-builder /app/main .

# Копируем необходимые для работы рантайм-файлы
COPY --from=backend-builder /app/client/dist ./client/dist
COPY --from=backend-builder /app/GeoIP2-ISP.mmdb ./
COPY --from=backend-builder /app/database ./database

# Открываем порт, который слушает наше приложение
EXPOSE 8080

# Команда для запуска приложения
CMD ["./main"]
