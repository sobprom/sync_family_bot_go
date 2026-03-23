package cmd

import (
	"log"
	"strings"
	"sync_family_bot_go/internal/domain"
	"sync_family_bot_go/internal/handlers"
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
	Config         *Config
	Bot            *telebot.Bot
	DB             *sqlx.DB
	messageHandler handlers.MessageHandler
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

	messageHandler := handlers.NewMessageHandler(db)

	app := &App{
		Config:         cfg,
		Bot:            b,
		DB:             db,
		messageHandler: messageHandler,
	}

	app.registerHandlers()

	return app
}

// Start запускает бесконечный цикл бота
func (a *App) Start() {
	log.Println("🚀 Бот на Go успешно запущен!")
	a.Bot.Start()
}

func (a *App) registerHandlers() {

	a.Bot.Use(func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(c telebot.Context) error {
			log.Printf("👤 [%d] %s: %s", c.Sender().ID, c.Sender().FirstName, c.Text())
			return next(c)
		}
	})

	// ЕДИНСТВЕННЫЙ обработчик для всего текста
	a.Bot.Handle(telebot.OnText, func(ctx telebot.Context) error {
		text := ctx.Text()
		chatID := ctx.Chat().ID

		// Получаем команду через твою модель
		command := domain.GetCommand(text)

		log.Printf("📨 Chat %d: команда %v, текст: %s", chatID, command, text)

		// Твой switch из Java
		switch command {
		case domain.CommandStart:
			// Обычный /start
			log.Printf("🤖 Команда: %s", command)

			return a.messageHandler.HandleStart(ctx)

		case domain.CommandStartWithInvite:
			// /start с инвайтом - извлекаем код
			inviteCode := text[7:] // после "/start "
			log.Printf("🤖 Команда: %s", command)
			return a.messageHandler.HandleStartWithInvite(ctx, inviteCode)

		case domain.CommandCreateFamily:
			// /create_family
			log.Printf("🤖 Команда: %s", command)
			return a.messageHandler.HandleCreateFamily(ctx)

		case domain.CommandUnknown:
			// Если начинается с /, но неизвестная команда
			if strings.HasPrefix(text, "/") {
				return ctx.Send("❌ Неизвестная команда. Доступные: /start, /create_family")
			}
			// Обычный текст
			return a.messageHandler.HandleText(ctx)
		}

		log.Println("✅ Обработчики бота зарегистрированы")

		return nil
	})

	// 3. Кнопки (Inline кнопки)
	a.Bot.Handle(telebot.OnCallback, a.messageHandler.HandleCallback)

}
