package service

import (
	"regexp"
	"strings"
	"unicode"
)

type ListParser struct{}

func NewListParser() *ListParser {
	return &ListParser{}
}

// Parse превращает строку вида "Хлеб, молоко; сыр\n яблоки" в список строк.
func (p *ListParser) Parse(text string) []string {
	if strings.TrimSpace(text) == "" {
		return []string{}
	}

	// Регулярка для деления по , ; или новой строке (один или более раз)
	re := regexp.MustCompile(`[,;\n]+`)
	parts := re.Split(text, -1)

	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, p.capitalize(trimmed))
		}
	}

	return result
}

func (p *ListParser) capitalize(str string) string {
	if str == "" {
		return ""
	}

	// Работаем с рунами (Runes) для корректной поддержки кириллицы
	runes := []rune(strings.ToLower(str))
	runes[0] = unicode.ToUpper(runes[0])

	return string(runes)
}
