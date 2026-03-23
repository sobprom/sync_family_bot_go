package handlers

import (
	"fmt"
	"log"
	"net/url"
	"sync_family_bot_go/internal/domain"
	"sync_family_bot_go/internal/repository"

	"gopkg.in/telebot.v3"
)

type CommandHandlerImpl struct {
	familyRepo repository.FamilyRepository
}

func NewCommandHandler(fr repository.FamilyRepository) CommandHandler {
	return &CommandHandlerImpl{familyRepo: fr}
}

func (h *CommandHandlerImpl) HandleStart(c telebot.Context) error {

	// Получаем Chat ID
	chatID := c.Chat().ID
	log.Printf("Chat ID: %d", chatID)
	user, err := h.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		return c.Send("❌ Ошибка доступа к базе данных.")
	}

	// Если пользователь найден (уже в семье)
	if user != nil {
		// Создаем клавиатуру с кнопками
		markup := &telebot.ReplyMarkup{}

		btnShop := markup.Data("🛒 Перейти к покупкам", string(domain.Refresh))
		invite := markup.Data("🔗 Получить ссылку приглашение", string(domain.CreateInvite))
		btnLeave := markup.Data("🚪 Выйти из группы", string(domain.ConfirmLeaveFamily))

		markup.Inline(
			markup.Row(btnShop),
			markup.Row(invite),
			markup.Row(btnLeave),
		)

		return c.Send(
			"🏠 Вы находитесь в семейном чате. Что хотите сделать?",
			markup,
		)
	}

	// Если пользователя нет в базе (новый пользователь)
	return c.Send(`
		👋 Привет! Я помогу синхронизировать список покупок в вашей семье.
		
		🔹 Напиши /create_family, чтобы создать новую группу.
		🔹 Или перейди по ссылке-приглашению от члена семьи.`)
}

func (h *CommandHandlerImpl) HandleStartWithInvite(c telebot.Context, inviteCode string) error {

	chatID, userName := getUserInfo(c)

	log.Printf("Обработка приглашения для chat ID: %d, код приглашения: %s", chatID, inviteCode)

	// Пытаемся присоединить пользователя к семье по инвайт-коду
	user, err := h.familyRepo.JoinFamily(chatID, inviteCode, userName)
	if err != nil {
		log.Printf("Ошибка при подключении к семье: %v", err)
		return c.Send("❌ Ошибка при подключении к семье. Попробуйте позже.")
	}

	// Проверяем результат
	if user != nil {
		// Успешно присоединились
		return c.Send("🤝 Вы успешно вступили в семью по ссылке!")
	}

	// Ссылка недействительна
	return c.Send("❌ Ссылка недействительна или устарела.")
}

func (h *CommandHandlerImpl) HandleCreateFamily(c telebot.Context) error {

	// Получаем данные пользователя
	chatID, userName := getUserInfo(c)

	log.Printf("Создание семьи для chat ID: %d", chatID)

	// Проверяем, есть ли уже пользователь в семье
	user, err := h.familyRepo.GetFamilyMemberByChatId(chatID)
	if err != nil {
		log.Printf("Ошибка при проверке пользователя: %v", err)
		return c.Send("❌ Ошибка доступа к базе данных.")
	}

	var inviteCode string
	var responseText string

	if user != nil {
		// Пользователь уже в семье - получаем существующий код
		if user.FamilyID == nil {
			return c.Send("❌ Ошибка: пользователь не привязан к семье.")
		}
		// Пользователь уже в семье - получаем существующий код
		inviteCode, err = h.familyRepo.GetFamilyCode(*user.FamilyID)
		if err != nil {
			log.Printf("Ошибка при получении кода семьи: %v", err)
			return c.Send("❌ Ошибка при получении кода приглашения.")
		}
		responseText = "📋 Ссылка-приглашение для вашей семьи уже существует! Нажми кнопку ниже, чтобы поделиться:"
		log.Printf("Семья уже существует для chat ID: %d, получен существующий код", chatID)
	} else {
		// Пользователя нет в семье - создаем новую
		newUser, err := h.familyRepo.CreateFamilyMember(chatID, userName)
		if err != nil {
			log.Printf("Ошибка при создании семьи: %v", err)
			return c.Send("❌ Ошибка при создании семьи.")
		}
		inviteCode, err = h.familyRepo.GetFamilyCode(*newUser.FamilyID)
		if err != nil {
			log.Printf("Ошибка при получении кода новой семьи: %v", err)
			return c.Send("❌ Ошибка при получении кода приглашения.")
		}
		responseText = "✅ Новая семья успешно создана! Нажми кнопку ниже, чтобы пригласить участников:"
		log.Printf("Новая семья создана для chat ID: %d", chatID)
	}

	// Формируем ссылки
	botName := "sync_family_bot"
	inviteLink := fmt.Sprintf("https://t.me/%s?start=%s", botName, inviteCode)

	shareURL := fmt.Sprintf("https://t.me/share/url?url=%s&text=%s",
		url.QueryEscape(inviteLink),
		url.QueryEscape("Присоединяйся к моей семье в боте покупок! 🛒"))

	// Создаем клавиатуру с кнопкой приглашения
	markup := &telebot.ReplyMarkup{}

	// Для URL кнопки нужно использовать другой метод
	btnInviteURL := markup.URL("👪 Отправить приглашение", shareURL)
	btnShop := markup.Data("🛒 Перейти к покупкам", string(domain.Refresh))

	markup.Inline(
		markup.Row(btnInviteURL),
		markup.Row(btnShop),
	)

	return c.Send(
		responseText,
		markup,
	)
}

func getUserInfo(c telebot.Context) (int64, string) {
	chatID := c.Chat().ID
	userName := c.Sender().FirstName
	if userName == "" {
		userName = c.Sender().Username // Fallback на username если FirstName пустой
	}
	return chatID, userName
}
