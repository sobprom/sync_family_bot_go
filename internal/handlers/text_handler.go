package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

// Обработчик обычного текста
func HandleText(c telebot.Context) error {
	log.Printf("✍️ Текст: %s", c.Text())
	return c.Send("Я получил ваше сообщение: " + c.Text())
}
