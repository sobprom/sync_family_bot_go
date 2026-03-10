package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

// Обработчик команд (начинаются с /)
func HandleCommand(c telebot.Context) error {
	log.Printf("🤖 Команда: %s", c.Text())
	return c.Send("Вы ввели команду: " + c.Text())
}

// Обработчик обычного текста
func HandleText(c telebot.Context) error {
	log.Printf("✍️ Текст: %s", c.Text())
	return c.Send("Я получил ваше сообщение: " + c.Text())
}

// Обработчик нажатий на кнопки (Callback)
func HandleCallback(c telebot.Context) error {
	log.Printf("🔘 Кнопка: %s", c.Callback().Data)
	// В Telegram важно "подтверждать" обработку колбэка
	return c.Respond(&telebot.CallbackResponse{Text: "Кнопка нажата!"})
}
