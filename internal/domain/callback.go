package domain

import "strings"

type CallBack string

const (
	Buy                CallBack = "buy"
	ConfirmClear       CallBack = "confirm_clear"
	ToggleModeEdit     CallBack = "toggle_mode_edit"
	ConfirmEditProduct CallBack = "confirm_edit_product"
	ClearAll           CallBack = "clear_all"
	Refresh            CallBack = "refresh"
	CreateInvite       CallBack = "create_invite"
	EditProduct        CallBack = "edit_product"
	DeleteProduct      CallBack = "delete_product"
	ConfirmLeaveFamily CallBack = "confirm_leave_family"
	Unknown            CallBack = ""
)

// String возвращает строковое значение (аналог getAction() в Java поле)
func (c CallBack) String() string {
	return string(c)
}

// GetCallBack определяет тип колбэка по строке (аналог статического getAction)
func GetCallBack(data string) (CallBack, string) {
	// 1. Убираем префикс телебота \f
	data = strings.TrimPrefix(data, "\f")

	// 2. Разделяем строку по разделителю telebot (обычно |)
	parts := strings.Split(data, "|")
	actionPart := parts[0] // Сама команда (например, confirm_edit_product)

	payload := ""
	if len(parts) > 1 {
		payload = parts[1] // Дополнительные данные (например, "42")
	}

	callbacks := []CallBack{
		ConfirmEditProduct, ConfirmClear, ConfirmLeaveFamily,
		ToggleModeEdit, CreateInvite, EditProduct, DeleteProduct,
		ClearAll, Refresh, Buy,
	}

	// 3. Ищем совпадение по первой части
	for _, c := range callbacks {
		if actionPart == string(c) {
			return c, payload
		}
	}

	return Unknown, ""
}
