package model

type Command int

const (
	CommandUnknown Command = iota
	CommandStart
	CommandStartWithInvite
	CommandCreateFamily
)

// String возвращает строковое представление команды
func (c Command) String() string {
	switch c {
	case CommandStart:
		return "/start"
	case CommandStartWithInvite:
		return "/start "
	case CommandCreateFamily:
		return "/create_family"
	default:
		return ""
	}
}

// GetCommand определяет команду по строке
func GetCommand(text string) Command {
	if text == "" {
		return CommandUnknown
	}

	// Проверяем с префиксом /start (с пробелом)
	if len(text) >= 7 && text[:7] == "/start " {
		return CommandStartWithInvite
	}

	// Проверяем точное совпадение
	switch text {
	case "/start":
		return CommandStart
	case "/create_family":
		return CommandCreateFamily
	default:
		return CommandUnknown
	}
}
