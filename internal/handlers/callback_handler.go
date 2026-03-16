package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

type CallbackHandlerImpl struct{}

func NewCallbackHandler() CallbackHandler {
	return &CallbackHandlerImpl{}
}

// Обработчик нажатий на кнопки (Callback)
func (r *CallbackHandlerImpl) HandleCallback(c telebot.Context) error {
	log.Printf("🔘 Кнопка: %s", c.Callback().Data)
	// В Telegram важно "подтверждать" обработку колбэка
	return c.Respond(&telebot.CallbackResponse{Text: "Кнопка нажата!"})
}
