package handlers

import (
	"log"

	"gopkg.in/telebot.v3"
)

func HandleStart(c telebot.Context) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}

func HandleStartWithInvite(c telebot.Context, inviteCode string) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}

func HandleCreateFamily(c telebot.Context) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)

	return c.Send("Вы ввели команду: " + c.Text())
}
