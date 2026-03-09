package cmd

import (
	"log"
	"sync_family_bot_go/handlers"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"gopkg.in/telebot.v3"
)

// App хранит все зависимости нашего бота
type App struct {
	Config *Config
	Bot    *telebot.Bot
	DB     *sqlx.DB
}

// NewApp — это "конструктор". Он собирает всё воедино.
func NewApp(cfg *Config) *App {

	// 2. Настраиваем бота
	pref := telebot.Settings{
		Token:  cfg.Token,
		Poller: &telebot.LongPoller{Timeout: 10 * time.Second},
	}

	db, err := sqlx.Connect("pgx", cfg.DBUrl)
	if err != nil {
		log.Fatal("❌ Ошибка базы:", err)
	}

	// 2. Настройка Goose (аналог Flyway)
	// Устанавливаем диалект
	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal("❌ Ошибка диалекта goose:", err)
	}

	log.Println("Run migrations...")
	// Запускаем миграции из папки "migrations"
	// db.DB — это извлечение стандартного *sql.DB из sqlx
	if err := goose.Up(db.DB, "migrations"); err != nil {
		log.Fatal("❌ Ошибка миграций:", err)
	}
	log.Println("✅ Миграции успешно применены")

	b, err := telebot.NewBot(pref)
	if err != nil {
		log.Fatal("❌ Ошибка Telegram:", err)
	}

	log.Printf("🤖 Бот авторизован как: %s (ID: %d)", b.Me.Username, b.Me.ID)

	return &App{
		Config: cfg,
		Bot:    b,
		DB:     db,
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
	a.Bot.Handle("/start", handlers.HandleCommand)
	a.Bot.Handle("/help", handlers.HandleCommand)

	// 2. Обычный текст (все, что не команда)
	a.Bot.Handle(telebot.OnText, handlers.HandleText)

	// 3. Кнопки (Inline кнопки)
	a.Bot.Handle(telebot.OnCallback, handlers.HandleCallback)

}

// Start запускает бесконечный цикл бота
func (a *App) Start() {
	log.Println("🚀 Бот на Go успешно запущен!")
	a.Bot.Start()
}
