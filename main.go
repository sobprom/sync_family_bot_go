package main

func main() {

	// 1. Загружаем настройки
	cfg := LoadConfig()

	// 2. Создаем приложение (со всеми базами и ботами внутри)
	app := NewApp(cfg)

	// 3. Регистрируем наши хендлеры (методы из handlers.go)
	app.RegisterHandlers()

	// 4. Поехали!
	app.Start()
}
