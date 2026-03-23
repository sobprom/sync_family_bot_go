package handlers

import (
	"sync_family_bot_go/internal/repository"
	"sync_family_bot_go/internal/service"

	"github.com/jmoiron/sqlx"
	"gopkg.in/telebot.v3"
)

// MessageHandlerImpl объединяет все обработчики
type MessageHandlerImpl struct {
	textHandler     TextHandler
	commandHandler  CommandHandler
	callbackHandler CallbackHandler
}

// NewMessageHandler создает новый экземпляр MessageHandler
func NewMessageHandler(db *sqlx.DB) MessageHandler {
	// Создаем репозитории
	familyRepo := repository.NewFamilyRepository(db)
	productRepo := repository.NewProductRepository(db)

	// Создаем сервисы
	listParser := service.NewListParser()
	uiService := service.NewUIService()
	notification := service.NewNotificationService(familyRepo, productRepo, uiService)

	// Создаем обработчики
	textHandler := NewTextHandler(familyRepo, productRepo, listParser, uiService, notification)
	commandHandler := NewCommandHandler(familyRepo)
	callbackHandler := NewCallbackHandler(familyRepo, productRepo, commandHandler, uiService, notification)

	return &MessageHandlerImpl{
		textHandler:     textHandler,
		commandHandler:  commandHandler,
		callbackHandler: callbackHandler,
	}
}

// HandleText делегирует обработку текстовому обработчику
func (h *MessageHandlerImpl) HandleText(c telebot.Context) error {
	return h.textHandler.HandleText(c)
}

// HandleStart делегирует обработку команды /start
func (h *MessageHandlerImpl) HandleStart(c telebot.Context) error {
	return h.commandHandler.HandleStart(c)
}

// HandleStartWithInvite делегирует обработку команды /start с инвайтом
func (h *MessageHandlerImpl) HandleStartWithInvite(c telebot.Context, inviteCode string) error {
	return h.commandHandler.HandleStartWithInvite(c, inviteCode)
}

// HandleCreateFamily делегирует обработку команды /create_family
func (h *MessageHandlerImpl) HandleCreateFamily(c telebot.Context) error {
	return h.commandHandler.HandleCreateFamily(c)
}

// HandleCallback делегирует обработку callback'ов
func (h *MessageHandlerImpl) HandleCallback(c telebot.Context) error {
	return h.callbackHandler.HandleCallback(c)
}
