package handlers

import (
	"fmt"
	"log"
	"sync_family_bot_go/internal/domain"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"
	"sync_family_bot_go/internal/repository"
	"sync_family_bot_go/internal/service"

	"gopkg.in/telebot.v3"
)

type CallbackHandlerImpl struct {
	familyRepo     repository.FamilyRepository
	productRepo    repository.ProductRepository
	commandHandler CommandHandler
	uiService      *service.UIService
	notification   *service.NotificationService
}

func NewCallbackHandler(
	fr repository.FamilyRepository,
	pr repository.ProductRepository,
	ch CommandHandler,
	ui *service.UIService,
	nt *service.NotificationService) CallbackHandler {
	return &CallbackHandlerImpl{
		familyRepo:     fr,
		productRepo:    pr,
		uiService:      ui,
		commandHandler: ch,
		notification:   nt,
	}
}

// Обработчик нажатий на кнопки (Callback)
func (r *CallbackHandlerImpl) HandleCallback(c telebot.Context) error {
	button, payload := domain.GetCallBack(c.Callback().Data)

	switch button {
	case domain.Refresh:
		log.Printf("🔄 [REFRESH] Обновление списка")
		return r.handleRefresh(c)
	case domain.CreateInvite:
		log.Printf("🔗 [INVITE] Генерация ссылки-приглашения")
		_ = c.Respond()
		return r.commandHandler.HandleCreateFamily(c)

	case domain.ConfirmEditProduct:
		// Здесь можно будет вызвать обработку редактирования, передав payload (ID)
		log.Printf("✏️ [EDIT_INIT] Запрос на правку продукта ID: %s", payload)
		return c.Respond(&telebot.CallbackResponse{Text: "Редактирование..."})
	case domain.ConfirmClear:
		log.Printf("❓ [CLEAR_CONFIRM] Запрос подтверждения очистки")
		return r.handleConfirmClear(c)
	case domain.ClearAll:
		log.Printf("🗑️ [CLEAR_ALL] Полная очистка списка выполнена")
		return r.handleClearAll(c)

	default:
		log.Printf("⚠️ [UNKNOWN] Неизвестный callback: '%s'. Данные: '%s'", button, c.Callback().Data)
		return c.Respond(&telebot.CallbackResponse{Text: "Команда не распознана"})
	}
}

func (r *CallbackHandlerImpl) handleClearAll(c telebot.Context) error {
	chatID := c.Chat().ID
	actorName := c.Sender().FirstName

	// 1. Получаем данные пользователя
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("❌ Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка доступа"})
	}

	familyID := user.FamilyID
	if familyID == nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Вы не состоите в семье"})
	}

	// 2. Удаляем все продукты семьи
	err = r.productRepo.DeleteAllByFamilyId(*familyID)
	if err != nil {
		log.Printf("❌ Ошибка при очистке списка: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Не удалось очистить список"})
	}

	header := fmt.Sprintf("🗑 *%s* очистил(а) список покупок", actorName)
	go r.notification.NotifyFamilyUpdate(c, *user.FamilyID, header)

	return nil

}

func (r *CallbackHandlerImpl) handleConfirmClear(c telebot.Context) error {
	selector := &telebot.ReplyMarkup{}
	btnConfirm := selector.Data("✅ ДА, УДАЛИТЬ", string(domain.ClearAll))
	btnCancel := selector.Data("❌ ОТМЕНА", string(domain.Refresh))
	selector.Inline(
		selector.Row(btnConfirm, btnCancel),
	)
	warningText := "⚠️ *Вы уверены, что хотите полностью очистить список покупок?*"

	err := c.Edit(warningText, selector, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("❌ Ошибка при показе подтверждения удаления: %v", err)
		return err
	}

	// 3. Подтверждаем колбэк (убираем состояние загрузки кнопки)
	return c.Respond()
}

func (r *CallbackHandlerImpl) handleRefresh(c telebot.Context) error {
	chatID := c.Chat().ID
	currentMsg := c.Message()

	// Достаем пользователя и список продуктов
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка: пользователь не найден"})
	}

	// Сбрасываем режим редактирования
	err = r.familyRepo.DropShoppingEditMode(chatID)
	if err != nil {
		return err
	}

	products, err := r.productRepo.GetAllProductsOrdered(*user.FamilyID)
	if err != nil {
		return err
	}

	// 2. Формируем обновление сообщения (аналог EditMessageText)
	// В telebot метод Edit заменяет текст и клавиатуру
	newText := "🛒 *Актуальный список покупок:*"
	newKeyboard := r.uiService.CreateShoppingListKeyboard(products, user.ShoppingListEditMode)

	sent, err := c.Bot().Send(c.Chat(), newText, &telebot.SendOptions{
		ParseMode:   telebot.ModeMarkdown,
		ReplyMarkup: newKeyboard,
	})
	if err != nil {
		return err
	}

	_ = c.Bot().Delete(currentMsg)

	if user.LastMessageID != nil && int(*user.LastMessageID) != currentMsg.ID {
		_ = c.Bot().Delete(&telebot.Message{
			ID:   int(*user.LastMessageID),
			Chat: c.Chat(),
		})
	}

	user.LastMessageID = new(int64(sent.ID))
	_ = r.familyRepo.UpdateLastMessageIds([]model.Users{*user})

	// 3. Подтверждаем callback, чтобы "часики" на кнопке исчезли
	return c.Respond()
}
