package service

import (
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

func (s *NotificationService) NotifyFamilyUpdate(ctx telebot.Context, familyID int64, header string) {

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
