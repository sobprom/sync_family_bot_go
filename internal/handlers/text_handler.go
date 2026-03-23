package handlers

import (
	"fmt"
	"log"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"
	"sync_family_bot_go/internal/repository"
	"sync_family_bot_go/internal/service"

	"gopkg.in/telebot.v3"
)

type TextHandlerImpl struct {
	familyRepo   repository.FamilyRepository
	productRepo  repository.ProductRepository
	listParser   *service.ListParser
	uiService    *service.UIService
	notification *service.NotificationService
}

func NewTextHandler(fr repository.FamilyRepository,
	pr repository.ProductRepository,
	lp *service.ListParser,
	ui *service.UIService,
	nt *service.NotificationService) TextHandler {
	return &TextHandlerImpl{familyRepo: fr, productRepo: pr, listParser: lp, uiService: ui, notification: nt}
}

func (h *TextHandlerImpl) HandleText(c telebot.Context) error {
	senderChatID := c.Chat().ID
	text := c.Text()

	// 1. Получаем/Создаем юзера
	user, err := h.getOrCreateUser(senderChatID, c.Sender().FirstName)
	if err != nil {
		return c.Send("❌ Ошибка доступа к базе данных.")
	}

	if user.FamilyID == nil {
		return c.Send("⚠️ Вы не состоите в семье.")
	}
	familyID := *user.FamilyID

	// 2. Обрабатываем ввод (Редактирование или Добавление)
	if err := h.processInput(user, text); err != nil {
		log.Printf("❌ Ошибка бизнес-логики: %v", err)
		return c.Send("❌ Не удалось обновить список.")
	}

	// 3. Рассылаем обновления всей семье
	header := fmt.Sprintf("🛒 %s обновил(а) список покупок:", user.Username)
	go h.notification.NotifyFamilyUpdate(c, familyID, header)

	return nil
}

// getOrCreateUser выносит логику инициализации пользователя
func (h *TextHandlerImpl) getOrCreateUser(chatID int64, firstName string) (*model.Users, error) {
	user, err := h.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return h.familyRepo.CreateFamilyMember(chatID, firstName)
	}
	return user, nil
}

// processInput решает: обновить существующий продукт или распарсить список новых
func (h *TextHandlerImpl) processInput(user *model.Users, text string) error {
	if user.EditingProductID != nil {
		updated, err := h.productRepo.UpdateProductName(*user.EditingProductID, text)
		if err == nil && updated {
			return h.familyRepo.DropEditingProductId(user.ChatID)
		}
		return err
	}

	products := h.listParser.Parse(text)
	if len(products) > 0 {
		return h.productRepo.AddProducts(*user.FamilyID, products)
	}
	return nil
}
