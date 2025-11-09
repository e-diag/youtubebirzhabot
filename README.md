# YouTube-Bot - Telegram Mini App для YouTube

Маркетплейс услуг для YouTube-каналов с системой проверки пользователей на мошенничество.

## 🚀 Быстрый старт

### Вариант 1: Docker Compose (рекомендуется)

```bash
# 1. Клонируйте репозиторий
git clone <repository-url>
cd YouTube-Bot

# 2. Создайте .env файл
cat > .env << EOF
DATABASE_URL=postgres://postgres:postgres@postgres:5432/youtube_market?sslmode=disable
PORT=8080
GIN_MODE=release
BOT_TOKEN=your_telegram_bot_token
MANAGER_ID=your_telegram_user_id
EOF

# 3. Запустите проект
docker-compose up -d
```

Приложение будет доступно по адресу: http://localhost:8080

### Вариант 2: Локальная разработка

#### Backend (Go)

```bash
cd backend

# Установите зависимости
go mod download

# Создайте .env файл
cat > .env << EOF
DATABASE_URL=postgres://user:password@localhost:5432/youtube_market?sslmode=disable
PORT=8080
BOT_TOKEN=your_telegram_bot_token
MANAGER_ID=your_telegram_user_id
EOF

# Запустите сервер
go run ./cmd/server
```

#### Frontend (React + Vite)

```bash
cd frontend

# Установите зависимости
npm install

# Запустите dev-сервер
npm run dev
```

Frontend будет доступен по адресу: http://localhost:3000

## 📁 Структура проекта

```
YouTube-Bot/
├── backend/              # Go backend
│   ├── cmd/
│   │   └── server/      # Точка входа приложения
│   └── internal/
│       ├── bot/         # Telegram bot логика
│       ├── db/          # База данных
│       ├── handlers/     # HTTP handlers
│       └── models/       # Модели данных
├── frontend/             # React frontend
│   ├── src/
│   │   ├── components/  # React компоненты
│   │   └── App.tsx      # Главный компонент
│   └── package.json
├── Dockerfile           # Docker образ
├── docker-compose.yml   # Docker Compose конфигурация
└── README.md
```

## 🔧 Переменные окружения

| Переменная | Описание | Обязательно |
|-----------|----------|-------------|
| `DATABASE_URL` | PostgreSQL connection string | Да |
| `PORT` | Порт сервера (по умолчанию: 8080) | Нет |
| `GIN_MODE` | Режим Gin (release/debug) | Нет |
| `BOT_TOKEN` | Telegram Bot Token | Нет |
| `MANAGER_ID` | Telegram User ID менеджера | Нет |

## 📡 API Endpoints

- `GET /api/ads` - Получить все объявления
  - Query params: `cat` (категория), `f1` (фильтр)
- `GET /api/myads?user_id=<id>` - Получить объявления пользователя
- `GET /api/profile/:username` - Получить объявления по username
- `GET /api/scammer/:username` - Проверить пользователя на мошенничество
- `GET /health` - Health check

## 🤖 Telegram Bot

Бот позволяет менеджеру управлять чёрным списком:

- `/addscam @username` - Добавить пользователя в чёрный список
- `/remscam @username` - Удалить пользователя из чёрного списка
- `/start` или `/menu` - Показать меню

## 🛠 Технологии

**Backend:**
- Go 1.24+
- Gin (HTTP framework)
- GORM (ORM)
- PostgreSQL
- Telegram Bot API

**Frontend:**
- React 18
- TypeScript
- Vite
- Tailwind CSS
- Radix UI

## 📦 Сборка для продакшена

```bash
# Backend
cd backend
go build -o server ./cmd/server

# Frontend
cd frontend
npm run build

# Docker
docker-compose -f docker-compose.yml build
docker-compose -f docker-compose.yml up -d
```

## 🔒 Безопасность

- Все SQL запросы используют параметризованные запросы (GORM)
- CORS настроен для работы с frontend
- Валидация входных данных на всех endpoints
- Обработка ошибок на всех уровнях

## 📝 Лицензия

MIT

