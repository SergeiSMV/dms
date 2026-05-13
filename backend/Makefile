# Makefile — шорткаты для частых команд разработки.
# Использование: make run, make build, make test, make lint

# Имя выходного бинарника.
APP_NAME=dms-server

# --- Сборка и запуск ---

# Скачать зависимости и обновить go.sum.
.PHONY: tidy
tidy:
	go mod tidy

# Собрать бинарник в папку bin/.
.PHONY: build
build:
	go build -o bin/$(APP_NAME) ./cmd/dms-server

# Запустить сервер без компиляции в файл (удобно при разработке).
.PHONY: run
run:
	go run ./cmd/dms-server

# --- Качество кода ---

# Запустить тесты всех пакетов с подробным выводом.
.PHONY: test
test:
	go test -v ./...

# Статический анализ кода (встроенный в Go).
.PHONY: lint
lint:
	go vet ./...
