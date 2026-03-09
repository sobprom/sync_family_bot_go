# Загружаем переменные из .env
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# 1. Запуск миграций (Goose)
migrate-up:
	goose -dir migrations postgres "$(DATABASE_URL)" up

migrate-status:
	goose -dir migrations postgres "$(DATABASE_URL)" status

# 2. Генерация метамодели (Jet — аналог jOOQ)
# Подставляем твои данные: localhost:54321 и bots_go
gen-db:
	jet -source=PostgreSQL \
		-dsn="$(DATABASE_URL)" \
		-schema=family_sync \
		-path=./internal/gen

