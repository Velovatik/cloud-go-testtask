#!/bin/bash

set -e

log() {
  echo "[$(date +"%Y-%m-%d %H:%M:%S")] $1"
}

log "Запуск базы данных..."
docker-compose up -d db

log "Ожидание готовности базы данных..."
until docker exec exec -T db pg_isready -U postgres; do
  log "База данных ещё не готова. Ожидание..."
  sleep 2
done

log "База данных готова."

log "Применение миграций..."
docker-compose run --rm migrations
log "Миграции применены успешно."

log "Запуск приложения..."
docker-compose up -d app --build

log "Все сервисы запущены."

log "Логи приложения:"

docker-compose logs -f app
