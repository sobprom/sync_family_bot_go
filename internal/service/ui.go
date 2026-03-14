package service

import (
	"fmt"
	"sync_family_bot_go/internal/domain"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"

	"gopkg.in/telebot.v3"
)

type UIService struct{}

// NewUIService создает новый экземпляр сервиса интерфейса
func NewUIService() *UIService {
	return &UIService{}
}

func (s *UIService) CreateShoppingListKeyboard(products []model.ShoppingList, edit bool) *telebot.ReplyMarkup {
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	if edit {
		return s.createEditListKeyboard(products)
	}

	for _, p := range products {
		label := p.ProductName
		if p.IsBought {
			label = "✅ " + label
		}

		btn := selector.Data(label, string(domain.Buy), fmt.Sprintf("%d", p.ID))
		rows = append(rows, selector.Row(btn))
	}

	if len(products) > 0 {
		btnEdit := selector.Data("⚙ Редактировать список", string(domain.ToggleModeEdit))
		btnFinish := selector.Data("🏁 Завершить покупки", string(domain.ConfirmClear))

		rows = append(rows, selector.Row(btnEdit))
		rows = append(rows, selector.Row(btnFinish))
	}

	selector.Inline(rows...)
	return selector
}

func (s *UIService) createEditListKeyboard(products []model.ShoppingList) *telebot.ReplyMarkup {
	selector := &telebot.ReplyMarkup{}
	var rows []telebot.Row

	for _, p := range products {
		btn := selector.Data("✏️ "+p.ProductName, string(domain.ConfirmEditProduct), fmt.Sprintf("%d", p.ID))
		rows = append(rows, selector.Row(btn))
	}

	if len(products) > 0 {
		btnBack := selector.Data("⬅️ Назад к покупкам", string(domain.Refresh))
		rows = append(rows, selector.Row(btnBack))
	}

	selector.Inline(rows...)
	return selector
}
