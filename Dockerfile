# 1. Фронтенд
FROM node:20-alpine AS frontend-builder
WORKDIR /app/frontend

# Копируем package.json
COPY frontend/package*.json ./

# Устанавливаем зависимости
RUN npm ci

# Копируем исходники
COPY frontend/ ./

# Даём права на выполнение
RUN chmod +x node_modules/.bin/vite

# Билдим
RUN npm run build

# Проверяем, что сборка прошла успешно
RUN ls -la dist/ || (echo "ERROR: Frontend build failed!" && exit 1)
RUN test -f dist/index.html || (echo "ERROR: index.html not found in dist!" && exit 1)

# 2. Бэкенд
FROM golang:1.23-alpine AS backend-builder
WORKDIR /app/backend

# Устанавливаем git
RUN apk add --no-cache git

# Проверяем версию Go
RUN go version

# Копируем только go.mod сначала
COPY backend/go.mod ./

# Синхронизируем go.sum (если не синхронизирован)
RUN go mod tidy

# Загружаем зависимости
RUN go mod download

# Копируем код
COPY backend/ ./

# Билдим Go (с подробным выводом ошибок)
# Примечание: статика копируется после сборки, чтобы не блокировать компиляцию
RUN set -e; \
    echo "Starting Go build..."; \
    CGO_ENABLED=0 GOOS=linux go build -v -o server ./cmd/server || { \
        echo "Build failed! Running go build again with full output:"; \
        CGO_ENABLED=0 GOOS=linux go build ./cmd/server; \
        exit 1; \
    }; \
    echo "Build successful!"

# Копируем статику из фронтенда (после сборки)
COPY --from=frontend-builder /app/frontend/dist ./static

# Проверяем, что index.html существует
RUN ls -la ./static/ || echo "Warning: static directory is empty"
RUN test -f ./static/index.html || (echo "ERROR: index.html not found after copy!" && exit 1)

# 3. Финальный образ
FROM alpine:latest
WORKDIR /app
RUN apk --no-cache add ca-certificates
COPY --from=backend-builder /app/backend/server .
COPY --from=backend-builder /app/backend/static ./static
# Копируем дополнительные статические файлы (terms.html, privacy.html) из корня проекта
COPY static/terms.html ./static/
COPY static/privacy.html ./static/
# Устанавливаем права доступа
RUN chmod -R 755 /app/static
# Финальная проверка
RUN ls -la /app/static/ && test -f /app/static/index.html || (echo "ERROR: index.html missing in final image!" && exit 1)
EXPOSE 8080
CMD ["./server"]
