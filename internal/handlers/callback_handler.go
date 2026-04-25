package handlers

import (
	"fmt"
	"log"
	"strconv"
	"sync_family_bot_go/internal/domain"
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
		log.Printf("✏️ [EDIT_INIT] Запрос на выбор действия по редактированию: %s", payload)
		return r.handleConfirmEdit(c, payload)
	case domain.EditProduct:
		log.Printf("✏️ [EDIT_PRODUCT] Запрос на правку продукта ID: %s", payload)
		return r.handleEditProduct(c, payload)
	case domain.DeleteProduct:
		log.Printf("✏️ [DELET_PRODUCT] Запрос на удаления продукта ID: %s", payload)
		return r.handleDeleteProduct(c, payload)
	case domain.ConfirmClear:
		log.Printf("❓ [CLEAR_CONFIRM] Запрос подтверждения очистки")
		return r.handleConfirmClear(c)
	case domain.ClearAll:
		log.Printf("🗑️ [CLEAR_ALL] Полная очистка списка выполнена")
		return r.handleClearAll(c)
	case domain.ToggleModeEdit:
		log.Printf("⚙ [EDIT] Режим редактирования списка")
		return r.handleEdit(c)
	case domain.Buy:
		log.Printf("✅ [BUY] Покупка товара")
		return r.handleBuy(c, payload)

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

func (r *CallbackHandlerImpl) handleConfirmEdit(c telebot.Context, payload string) error {
	chatID := c.Chat().ID

	// Получаем пользователя по chatId
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("❌ Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка доступа"})
	}

	familyID := user.FamilyID
	if familyID == nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Вы не состоите в семье"})
	}

	// Парсим ID продукта из payload
	productID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		log.Printf("❌ Ошибка парсинга ID продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных товара"})
	}

	// Ищем продукт
	product, err := r.productRepo.FindProduct(productID)
	if err != nil {
		log.Printf("❌ Ошибка поиска продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Товар не найден"})
	}

	if product == nil || product.ID == 0 {
		return c.Respond(&telebot.CallbackResponse{Text: "Товар уже удален или не найден"})
	}

	// Создаем клавиатуру с кнопками действий
	selector := &telebot.ReplyMarkup{}

	// Кнопка "Изменить" - формат с тремя параметрами
	btnEdit := selector.Data("📝 Изменить", string(domain.EditProduct), fmt.Sprintf("%d", productID))
	// Кнопка "Удалить" - формат с тремя параметрами
	btnDelete := selector.Data("🗑 Удалить", string(domain.DeleteProduct), fmt.Sprintf("%d", productID))
	// Кнопка "Отмена"
	btnCancel := selector.Data("❌ ОТМЕНА", string(domain.ToggleModeEdit))

	selector.Inline(
		selector.Row(btnEdit, btnDelete),
		selector.Row(btnCancel),
	)

	// Формируем сообщение с вопросом
	productName := product.ProductName
	text := fmt.Sprintf("🧩 *Что делаем с:* %s ?", productName)

	// Редактируем сообщение, показываем кнопки выбора действия
	err = c.Edit(text, selector, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("❌ Ошибка редактирования сообщения: %v", err)
		return err
	}

	// Подтверждаем callback
	return c.Respond()

}

func (r *CallbackHandlerImpl) handleEditProduct(c telebot.Context, payload string) error {
	chatID := c.Chat().ID

	// Парсим ID продукта из payload
	productID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		log.Printf("❌ Ошибка парсинга ID продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных товара"})
	}

	// Получаем пользователя по chatId
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("❌ Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка доступа"})
	}

	familyID := user.FamilyID
	if familyID == nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Вы не состоите в семье"})
	}

	// Ищем продукт
	product, err := r.productRepo.FindProduct(productID)
	if err != nil {
		log.Printf("❌ Ошибка поиска продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Товар не найден"})
	}

	if product == nil || product.ID == 0 {
		// Продукт удален или не найден
		return c.Respond(&telebot.CallbackResponse{Text: "Товар уже удален или не найден"})
	}

	// Проставляем статус редактирования (сохраняем ID продукта, который редактируется)
	err = r.familyRepo.SetEditingProductId(chatID, productID)
	if err != nil {
		log.Printf("❌ Ошибка установки статуса редактирования: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка при начале редактирования"})
	}

	// Формируем сообщение с запросом нового названия
	productName := product.ProductName
	editText := fmt.Sprintf("🧩 *Введите новое название для:* %s", productName)

	// Редактируем сообщение, убираем кнопки, оставляем только текст
	err = c.Edit(editText, &telebot.ReplyMarkup{}, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("❌ Ошибка редактирования сообщения: %v", err)
		return err
	}

	// Подтверждаем callback, чтобы убрать "часики" на кнопке
	return c.Respond()

}

func (r *CallbackHandlerImpl) handleDeleteProduct(c telebot.Context, payload string) error {
	chatID := c.Chat().ID
	actorName := c.Sender().FirstName

	// Парсим ID продукта из payload
	productID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		log.Printf("❌ Ошибка парсинга ID продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных товара"})
	}

	log.Printf("🗑️ Удаление продукта ID: %d", productID)

	// Получаем пользователя по chatId
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("❌ Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка доступа"})
	}

	familyID := user.FamilyID
	if familyID == nil {
		return c.Respond(&telebot.CallbackResponse{Text: "Вы не состоите в семье"})
	}

	// Находим продукт перед удалением, чтобы получить его название для уведомления
	product, err := r.productRepo.FindProduct(productID)
	if err != nil {
		log.Printf("❌ Ошибка поиска продукта: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Товар не найден"})
	}

	if product == nil || product.ID == 0 {
		return c.Respond(&telebot.CallbackResponse{Text: "Товар уже удален или не найден"})
	}

	// Удаляем продукт
	err = r.productRepo.DeleteByProductId(productID)
	if err != nil {
		log.Printf("❌ Ошибка при удалении продукта %d: %v", productID, err)
		return c.Respond(&telebot.CallbackResponse{Text: "Не удалось удалить товар"})
	}

	header := fmt.Sprintf("🗑 *%s* удалил(а) товар: *%s*", actorName, product.ProductName)

	go r.notification.NotifyFamilyUpdate(c, *familyID, header)

	// Подтверждаем callback
	return c.Respond()

}

func (r *CallbackHandlerImpl) handleConfirmClear(c telebot.Context) error {
	selector := &telebot.ReplyMarkup{}
	btnConfirm := selector.Data("✅ ДА, УДАЛИТЬ", string(domain.ClearAll))
	btnCancel := selector.Data("❌ ОТМЕНА", string(domain.Refresh))
	selector.Inline(
		selector.Row(btnConfirm, btnCancel),
	)
	warningText := "⚠️ *Удалить все купленные товары?*"

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

	// Достаем пользователя и список продуктов
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("Ошибка поиска пользователя: %v", err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка: пользователь не найден"})
	}

	// Сбрасываем режим редактирования
	user, err = r.familyRepo.DropShoppingEditMode(chatID)
	if err != nil {
		return err
	}

	go r.notification.NotifyUserUpdate(c, user, "🛒 *Актуальный список покупок:*")

	// 3. Подтверждаем callback, чтобы "часики" на кнопке исчезли
	return c.Respond()
}

func (r *CallbackHandlerImpl) handleEdit(c telebot.Context) error {

	chatID := c.Chat().ID

	updatedUser, err := r.familyRepo.ToggleShoppingEditMode(chatID)
	if err != nil {
		log.Printf("❌ Ошибка переключения режима: %v", err)
		return err
	}

	// 2. Получаем актуальные продукты для отрисовки новой клавиатуры
	products, err := r.productRepo.GetAllProductsOrdered(*updatedUser.FamilyID)
	if err != nil {
		return err
	}

	// 3. Обновляем сообщение (текст и клавиатуру)
	text := "🛒 *Режим редактирования:*"

	kb := r.uiService.CreateShoppingListKeyboard(products, updatedUser.ShoppingListEditMode)

	err = c.Edit(text, kb, telebot.ModeMarkdown)
	if err != nil {
		log.Printf("❌ Ошибка обновления сообщения: %v", err)
		return err
	}

	return c.Respond()
}

func (r *CallbackHandlerImpl) handleBuy(c telebot.Context, payload string) error {
	chatID := c.Chat().ID
	user, err := r.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		return err
	}

	productID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil {
		// Если в payload не число, отвечаем пользователю или логируем
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка данных товара"})
	}

	product, err := r.productRepo.InverseBought(productID)

	if err != nil {
		// Ошибка базы данных
		log.Printf("Ошибка при покупке товара %d: %v", productID, err)
		return c.Respond(&telebot.CallbackResponse{Text: "Ошибка базы данных"})
	}

	if product.ID == 0 {
		return c.Respond(&telebot.CallbackResponse{Text: "Товар уже удален или не найден"})
	}

	action := "отменил(а) покупку"

	if product.IsBought {
		action = "купил(а)"
	}

	header := fmt.Sprintf("🛒 *%s* %s: *%s*", c.Sender().FirstName, action, product.ProductName)

	go r.notification.NotifyFamilyUpdate(c, *user.FamilyID, header)
	_ = c.Respond()

	return nil

}
