package service

import (
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"
	"sync_family_bot_go/internal/repository"

	"gopkg.in/telebot.v3"
)

type NotificationService struct {
	familyRepo  repository.FamilyRepository
	productRepo repository.ProductRepository
	uiService   *UIService
}

func NewNotificationService(
	fr repository.FamilyRepository,
	pr repository.ProductRepository,
	ui *UIService) *NotificationService {
	return &NotificationService{familyRepo: fr, productRepo: pr, uiService: ui}
}

func (s *NotificationService) NotifyFamilyNewMessage(ctx telebot.Context, familyID int64, header string) {

	products, _ := s.productRepo.GetAllProductsOrdered(familyID)
	members, _ := s.familyRepo.GetFamilyMembersByFamilyId(familyID)
	bot := ctx.Bot()

	for i := range members {
		member := &members[i]
		kb := s.uiService.CreateShoppingListKeyboard(products, member.ShoppingListEditMode)

		// Отправляем новое сообщение
		sent, err := bot.Send(&telebot.Chat{ID: member.ChatID}, header, kb, telebot.ModeMarkdown)
		if err == nil {
			// Удаляем старое
			if member.LastMessageID != nil {
				_ = bot.Delete(&telebot.Message{ID: int(*member.LastMessageID), Chat: &telebot.Chat{ID: member.ChatID}})
			}
			// Обновляем ID в локальной структуре
			member.LastMessageID = new(int64(sent.ID))
		}
	}
	// Массово обновляем ID в базе
	_ = s.familyRepo.UpdateLastMessageIds(members)
}

func (s *NotificationService) NotifyFamilyUpdate(ctx telebot.Context, familyID int64, header string) {

	products, _ := s.productRepo.GetAllProductsOrdered(familyID)
	members, _ := s.familyRepo.GetFamilyMembersByFamilyId(familyID)
	bot := ctx.Bot()

	for i := range members {
		member := &members[i]
		kb := s.uiService.CreateShoppingListKeyboard(products, member.ShoppingListEditMode)

		chat := &telebot.Chat{ID: member.ChatID}
		var sent *telebot.Message
		var errEdit error
		var errSend error

		// 1. Попытка обновить существующее сообщение
		if member.LastMessageID != nil {
			oldMsg := &telebot.Message{ID: int(*member.LastMessageID), Chat: chat}
			sent, errEdit = bot.Edit(oldMsg, header, kb, telebot.ModeMarkdown)
		}

		// 2. Если ID нет ИЛИ Edit вернул ошибку (сообщение удалено/устарело/не изменилось)
		if member.LastMessageID == nil || errEdit != nil {

			// Отправляем новое
			sent, errSend = bot.Send(chat, header, kb, telebot.ModeMarkdown)

			// Если старое сообщение было, пробуем его удалить (на всякий случай)
			if member.LastMessageID != nil {
				_ = bot.Delete(&telebot.Message{ID: int(*member.LastMessageID), Chat: chat})
			}

		}

		// 3. Обновляем ID в структуре, если отправка/редактирование успешны
		if errSend == nil && sent != nil {
			member.LastMessageID = new(int64(sent.ID))
		}
	}

	// Массово обновляем ID в базе
	_ = s.familyRepo.UpdateLastMessageIds(members)
}

func (s *NotificationService) NotifyUserUpdate(ctx telebot.Context, user *model.Users, header string) {

	products, _ := s.productRepo.GetAllProductsOrdered(*user.FamilyID)
	bot := ctx.Bot()
	chat := &telebot.Chat{ID: user.ChatID}

	kb := s.uiService.CreateShoppingListKeyboard(products, user.ShoppingListEditMode)

	var sent *telebot.Message
	var errEdit error
	var errSend error

	// 2. Пробуем отредактировать старый список
	if user.LastMessageID != nil {
		oldMsg := &telebot.Message{ID: int(*user.LastMessageID), Chat: chat}
		sent, errEdit = bot.Edit(oldMsg, header, kb, telebot.ModeMarkdown)
	}

	// 3. Если редактирование не вышло — удаляем старое (если было) и шлем новое
	if user.LastMessageID == nil || errEdit != nil {
		if user.LastMessageID != nil {
			_ = bot.Delete(&telebot.Message{ID: int(*user.LastMessageID), Chat: chat})
		}

		sent, errSend = bot.Send(chat, header, kb, telebot.ModeMarkdown)
	}

	// 4. Сохраняем новый ID и обновляем в БД
	if errSend == nil && sent != nil {
		newID := int64(sent.ID)
		user.LastMessageID = &newID
		_ = s.familyRepo.UpdateLastMessageIds([]model.Users{*user})
	}
}
