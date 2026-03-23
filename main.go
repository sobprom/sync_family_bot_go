package main

import "sync_family_bot_go/cmd"

func main() {

	// 1. Загружаем настройки
	cfg := cmd.LoadConfig()

	// 2. Создаем приложение (со всеми базами и ботами внутри)
	app := cmd.NewApp(cfg)

	// 3. Поехали!
	app.Start()
}
