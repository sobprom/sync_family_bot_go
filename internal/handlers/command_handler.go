package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

type CommandHandlerImpl struct {
}

func NewCommandHandler() CommandHandler {
	return &CommandHandlerImpl{}
}

func (h *CommandHandlerImpl) HandleStart(c telebot.Context) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}

func (h *CommandHandlerImpl) HandleStartWithInvite(c telebot.Context, inviteCode string) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}

func (h *CommandHandlerImpl) HandleCreateFamily(c telebot.Context) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}
