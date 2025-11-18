# GitHub Actions CI/CD

Этот проект использует GitHub Actions для автоматической сборки Docker образов и деплоя на сервер.

## Workflows

### 1. CI/CD Pipeline (`.github/workflows/ci-cd.yml`)

Автоматически запускается при:
- Push в ветки `main`, `master`, `develop`
- Pull Request в `main` или `master`
- Ручной запуск через GitHub Actions UI

**Что делает:**
- Собирает Docker образ
- Публикует в GitHub Packages
- Автоматически деплоит на production (только для main/master)

### 2. Manual Deploy (`.github/workflows/deploy.yml`)

Ручной деплой с выбором окружения и тега образа.

**Использование:**
1. Перейдите в Actions → Manual Deploy
2. Нажмите "Run workflow"
3. Выберите окружение (production/staging)
4. Укажите тег образа (или оставьте пустым для latest)

## Настройка Secrets

В настройках репозитория (Settings → Secrets and variables → Actions) добавьте:

### Обязательные:
- `SSH_PRIVATE_KEY` - приватный SSH ключ для доступа к серверу
- `SERVER_HOST` - IP или домен сервера
- `SERVER_USER` - пользователь для SSH (обычно `root`)
- `SERVER_PATH` - путь к проекту на сервере (например, `/root/youtube-market`)

### Опциональные:
- `SERVER_URL` - URL сервера для health check (например, `https://your-domain.com`)

## Настройка GitHub Packages

1. Перейдите в Settings → Actions → General
2. В разделе "Workflow permissions" выберите "Read and write permissions"
3. Включите "Allow GitHub Actions to create and approve pull requests"

## Использование образа из GitHub Packages

### На сервере:

1. Создайте Personal Access Token (PAT) с правами `read:packages`
2. Войдите в GitHub Container Registry:
   ```bash
   echo "YOUR_GITHUB_TOKEN" | docker login ghcr.io -u YOUR_USERNAME --password-stdin
   ```

3. Обновите `docker-compose.yml`:
   ```yaml
   app:
     image: ghcr.io/OWNER/REPO:latest
     # вместо build: .
   ```

4. Запустите:
   ```bash
   docker compose pull
   docker compose up -d
   ```

## Локальный деплой через скрипт

Используйте скрипт `scripts/deploy.sh` для быстрого деплоя:

```bash
# Установите переменные окружения
export SERVER_HOST=your-server.com
export SERVER_USER=root
export SERVER_PATH=/root/youtube-market
export GITHUB_TOKEN=your_github_token

# Деплой latest версии
./scripts/deploy.sh

# Деплой конкретного тега
./scripts/deploy.sh v1.0.0
```

## Теги образов

Образы автоматически тегируются:
- `latest` - для ветки по умолчанию (main/master)
- `branch-name` - для других веток
- `branch-name-sha` - с SHA коммита
- `v1.0.0` - для тегов (semver)

## Troubleshooting

### Ошибка доступа к GitHub Packages
- Убедитесь, что PAT имеет права `read:packages`
- Проверьте, что workflow имеет права на запись в packages

### Ошибка SSH подключения
- Проверьте, что SSH ключ добавлен в `authorized_keys` на сервере
- Убедитесь, что `SERVER_HOST` и `SERVER_USER` указаны правильно

### Образ не найден
- Проверьте, что образ был успешно собран и опубликован
- Убедитесь, что используете правильный тег образа

