package handlers

import "gopkg.in/telebot.v3"

// TextHandlerImpl определяет контракт для обработчика текстовых сообщений
type TextHandler interface {
	HandleText(c telebot.Context) error
}

// CommandHandler определяет контракт для обработчика команд
type CommandHandler interface {
	HandleStart(c telebot.Context) error
	HandleStartWithInvite(c telebot.Context, inviteCode string) error
	HandleCreateFamily(c telebot.Context) error
}

// CallbackHandler определяет контракт для обработчика callback'ов
type CallbackHandler interface {
	HandleCallback(c telebot.Context) error
}

// Можно объединить все в один интерфейс
type MessageHandler interface {
	TextHandler
	CommandHandler
	CallbackHandler
}
