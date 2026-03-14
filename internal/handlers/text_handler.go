package handlers

import (
	"fmt"
	"log"
	"sync_family_bot_go/internal/repository"
	"sync_family_bot_go/internal/service"

	"gopkg.in/telebot.v3"
)

type TextHandler struct {
	familyRepo  *repository.FamilyRepository
	productRepo *repository.ProductRepository
	listParser  *service.ListParser
	uiService   *service.UIService
}

func NewTextHandler(
	familyRepo *repository.FamilyRepository,
	productRepo *repository.ProductRepository,
	listParser *service.ListParser,
	uiService *service.UIService) *TextHandler {

	return &TextHandler{
		familyRepo:  familyRepo,
		productRepo: productRepo,
		listParser:  listParser,
		uiService:   uiService,
	}
}

// Обработчик обычного текста
func (h *TextHandler) HandleText(c telebot.Context) error {
	log.Printf("✍️ Текст: %s", c.Text())
	senderChatID := c.Chat().ID
	text := c.Text()

	// 1. Получаем или создаем пользователя
	currentUser, err := h.familyRepo.GetFamilyMemberByChatId(senderChatID)
	if err != nil {
		log.Printf("❌ Ошибка БД (GetMember): %v", err)
		return c.Send("❌ Ошибка доступа к базе данных.")
	}
	if currentUser == nil {
		currentUser, err = h.familyRepo.CreateFamilyMember(senderChatID, c.Sender().FirstName)
		if err != nil {
			log.Printf("❌ Ошибка БД (Create): %v", err)
			return c.Send("❌ Не удалось создать профиль.")
		}
	}

	familyID := *currentUser.FamilyID

	// 2. Обновление названия продукта, если пользователь в режиме редактирования
	if currentUser.EditingProductID != nil {
		updated, updateErr := h.productRepo.UpdateProductName(*currentUser.EditingProductID, text)
		err = updateErr // сохраняем для общей проверки
		if updated && err == nil {
			err = h.familyRepo.DropEditingProductId(senderChatID)
		}

		// добавление продуктов, если пользователь не в режиме редактирования
	} else {
		products := h.listParser.Parse(text)
		if len(products) > 0 {
			err = h.productRepo.AddProducts(familyID, products)
		}
	}

	// 3. Единый обработчик ошибок для шага №2
	if err != nil {
		log.Printf("❌ Ошибка бизнес-логики: %v", err)
		return c.Send("❌ Не удалось обновить список. Попробуйте позже.")
	}

	// 4. Получаем данные для отображения
	productsOrdered, err := h.productRepo.GetAllProductsOrdered(familyID)
	if err != nil {
		log.Printf("❌ Ошибка получения списка: %v", err)
		return c.Send("❌ Не удалось обновить список. Попробуйте позже.")
	}

	allUsers, err := h.familyRepo.GetFamilyMembersByFamilyId(familyID)

	if err != nil {
		log.Printf("❌ Ошибка получения пользователей семьи: %v", err)
		return c.Send("❌ Не удалось обновить список. Попробуйте позже.")
	}

	textMsg := fmt.Sprintf("🛒 %s обновил(а) список покупок:", currentUser.Username)

	// 1. Проходим по всем участникам семьи
	for i := range allUsers {
		user := &allUsers[i] // Берем по указателю, чтобы обновить LastMessageID внутри структуры

		// 2. Удаляем предыдущее сообщение, если оно было (аналог DeleteMessage)
		if user.LastMessageID != nil && *user.LastMessageID != 0 {
			// Игнорируем ошибку удаления (сообщение могло быть удалено вручную или истек срок)
			_ = c.Bot().Delete(&telebot.Message{
				ID:   int(*user.LastMessageID),
				Chat: &telebot.Chat{ID: user.ChatID},
			})
		}

		// 3. Формируем клавиатуру и текст
		keyboard := h.uiService.CreateShoppingListKeyboard(productsOrdered, user.ShoppingListEditMode)

		// 4. Отправляем новое сообщение
		sent, err := c.Bot().Send(&telebot.Chat{ID: user.ChatID}, textMsg, keyboard)

		if err != nil {

			log.Printf("⚠️ Не удалось отправить сообщение пользователю %d: %v", user.ChatID, err)
			continue
		}

		if sent != nil {
			user.LastMessageID = new(int64(sent.ID))
		}
	}

	// 5. Массово обновляем LastMessageID в базе данных (аналог familyRepository.updateLastMessageId)
	err = h.familyRepo.UpdateLastMessageIds(allUsers)
	if err != nil {
		log.Printf("❌ Ошибка при сохранении LastMessageIds: %v", err)
	}

	return nil
}
