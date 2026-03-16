package main

import "sync_family_bot_go/cmd"

//go:generate jet -source=PostgreSQL -dsn=postgres://bot:bot@localhost:5433/bots_go?sslmode=disable -schema=family_sync -path=./internal/gen

func main() {

	// 1. Загружаем настройки
	cfg := cmd.LoadConfig()

	// 2. Создаем приложение (со всеми базами и ботами внутри)
	app := cmd.NewApp(cfg)

	// 3. Поехали!
	app.Start()
}
