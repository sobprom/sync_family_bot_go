package main

import (
	"log"
	"strings"
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

	return &App{
		Config: cfg,
		Bot:    b,
	}
}

func (a *App) RegisterHandlers() {
	// Middleware: выполняется для каждого входящего апдейта
	a.Bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			// 1. Проверяем нажатие кнопки
			if c.Callback() != nil {
				return a.handleCallback(c)
			}

			// 2. Если это сообщение
			if c.Message() != nil {
				text := c.Text()
				if strings.HasPrefix(text, "/") {
					return a.handleCommand(c)
				}
				return a.handleText(c)
			}

			// Обязательно вызываем next, если хочешь,
			// чтобы работали другие хендлеры (если они есть)
			return next(c)
		}
	})
}

// Start запускает бесконечный цикл бота
func (a *App) Start() {
	log.Println("🚀 Бот на Go успешно запущен!")
	a.Bot.Start()
}
