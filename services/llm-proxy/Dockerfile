# Этап 1: сборка Go-бинарника в отдельном builder-образе.
FROM golang:1.22-alpine AS builder
# Рабочая директория внутри контейнера сборки.
WORKDIR /app
# Сначала копируем только go.mod/go.sum для лучшего кеширования зависимостей.
COPY go.mod go.sum ./
# Скачиваем зависимости модуля.
RUN go mod download
# Копируем исходный код проекта.
COPY . .
# Собираем статический Linux-бинарник (меньше размер за счет -w -s).
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Этап 2: минимальный runtime-образ без лишних утилит (повышенная безопасность).
FROM gcr.io/distroless/static-debian12:nonroot
# Копируем только готовый бинарник из builder-этапа.
COPY --from=builder /app/server /server
# Документируем, что приложение слушает порт 8080.
EXPOSE 8080
# Запускаем приложение от непривилегированного пользователя.
USER nonroot:nonroot
# Команда запуска контейнера.
ENTRYPOINT ["/server"]
