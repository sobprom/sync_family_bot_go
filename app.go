package main

import (
	"log"
	"time"

	"gopkg.in/telebot.v3"
)

// App хранит все зависимости нашего бота
type App struct {
	Config *Config
	Bot    *telebot.Bot
}

// NewApp — это "конструктор". Он собирает всё воедино.
func NewApp(cfg *Config) *App {

	// 2. Настраиваем бота
	pref := telebot.Settings{
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal("❌ Ошибка Telegram:", err)
	}

	log.Printf("🤖 Бот авторизован как: %s (ID: %d)", b.Me.Username, b.Me.ID)

	return &App{
		Config: cfg,
		Bot:    b,
	}
}

func (a *App) RegisterHandlers() {

	a.Bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			log.Printf("👤 [%d] %s: %s", c.Sender().ID, c.Sender().FirstName, c.Text())
			return next(c)
		}
	})

	// 1. Команды (явная регистрация)
	a.Bot.Handle("/start", a.handleCommand)
	a.Bot.Handle("/help", a.handleCommand)

	// 2. Обычный текст (все, что не команда)
	a.Bot.Handle(telebot.OnText, a.handleText)

	// 3. Кнопки (Inline кнопки)
	a.Bot.Handle(telebot.OnCallback, a.handleCallback)

}

// Start запускает бесконечный цикл бота
func (a *App) Start() {
	log.Println("🚀 Бот на Go успешно запущен!")
	a.Bot.Start()
}
