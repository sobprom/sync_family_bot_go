package main

import (
	"log"
	"os"
)

type Config struct {
	Token string
	DBUrl string
}

// LoadConfig читает данные из системы (Railway Settings)
func LoadConfig() *Config {
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
