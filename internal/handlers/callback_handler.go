package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

// Обработчик нажатий на кнопки (Callback)
func HandleCallback(c telebot.Context) error {
	log.Printf("🔘 Кнопка: %s", c.Callback().Data)
	// В Telegram важно "подтверждать" обработку колбэка
	return c.Respond(&telebot.CallbackResponse{Text: "Кнопка нажата!"})
}
