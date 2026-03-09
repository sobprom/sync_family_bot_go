package cmd

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Token string
	DBUrl string
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		// Не фатально, может быть используем системные переменные
		log.Println("⚠️ .env файл не найден, используем системные переменные окружения")
	}
	token := os.Getenv("BOT_TOKEN")
	dbUrl := os.Getenv("DATABASE_URL")

	if token == "" {
		log.Fatal("❌ Переменная BOT_TOKEN не установлена!")
	}

	return &Config{
		Token: token,
		DBUrl: dbUrl,
	}
}
