#!/usr/bin/env bash
set -euo pipefail

# ── настройки ─────────────────────────────────────────────────────────────────
REPO_URL="git@github.com:SergeiSMV/local_models_api.git"
APP_DIR="$HOME/local_models_api"
OLLAMA_URL="http://192.168.0.109:11434"
PORT=8080

# ── вывод ─────────────────────────────────────────────────────────────────────
GREEN='\033[0;32m'; YELLOW='\033[1;33m'; RED='\033[0;31m'; NC='\033[0m'
ok()   { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[!!]${NC} $*"; }
die()  { echo -e "${RED}[ERR]${NC} $*"; exit 1; }
step() { echo -e "\n${YELLOW}==> $*${NC}"; }

# ── 1. Docker ─────────────────────────────────────────────────────────────────
step "Проверка Docker"
if command -v docker &>/dev/null; then
    ok "Docker уже установлен: $(docker --version)"
else
    warn "Docker не найден, устанавливаем..."
    curl -fsSL https://get.docker.com | sudo sh
    sudo usermod -aG docker "$USER"
    warn "Docker установлен. Выполни 'newgrp docker' и запусти скрипт снова."
    exit 0
fi

if ! docker ps &>/dev/null; then
    die "Нет доступа к Docker без sudo. Выполни 'newgrp docker' и попробуй снова."
fi

# ── 2. docker compose ─────────────────────────────────────────────────────────
step "Проверка docker compose"
if docker compose version &>/dev/null; then
    ok "$(docker compose version)"
else
    die "docker compose plugin не найден. Нужен Docker Engine >= 20.10"
fi

# ── 3. Репозиторий ────────────────────────────────────────────────────────────
step "Репозиторий"
if [ -d "$APP_DIR/.git" ]; then
    ok "Репозиторий найден, обновляем..."
    git -C "$APP_DIR" pull
else
    git clone "$REPO_URL" "$APP_DIR"
    ok "Клонировано в $APP_DIR"
fi
cd "$APP_DIR"

# ── 4. .env ───────────────────────────────────────────────────────────────────
step "Конфигурация .env"
if [ -f .env ]; then
    ok ".env уже существует, не трогаем"
else
    echo ""
    echo "  Введи API-ключ или нажми Enter для автогенерации:"
    read -r -p "  API_KEY: " INPUT_KEY
    API_KEY=${INPUT_KEY:-$(openssl rand -hex 32)}

    cat > .env <<EOF
PORT=$PORT
OLLAMA_URL=$OLLAMA_URL
API_KEYS=$API_KEY
EOF
    ok ".env создан"
    echo ""
    echo -e "  ${GREEN}API_KEY: $API_KEY${NC}"
    echo "  Добавь этот ключ в заголовок X-API-Key всех запросов к API."
    echo "  Пример: curl -H \"X-API-Key: $API_KEY\" http://127.0.0.1:$PORT/v1/models"
    echo ""
fi

# ── 5. Запуск ─────────────────────────────────────────────────────────────────
step "Сборка и запуск контейнера"
docker compose up -d --build
ok "Контейнер запущен"

# ── 6. Smoke-тест ─────────────────────────────────────────────────────────────
step "Smoke-тест /health"
sleep 2
HEALTH=$(curl -s "http://127.0.0.1:$PORT/health")
if echo "$HEALTH" | grep -q '"ok"'; then
    ok "Сервис отвечает: $HEALTH"
else
    die "Сервис не отвечает. Логи: docker compose -f $APP_DIR/docker-compose.yml logs"
fi

# ── готово ────────────────────────────────────────────────────────────────────
echo ""
ok "Готово!"
echo "  Endpoint : http://127.0.0.1:$PORT"
echo "  API ключ : $(grep API_KEYS .env | cut -d= -f2)"
echo "  Логи     : docker compose logs -f"
echo ""
