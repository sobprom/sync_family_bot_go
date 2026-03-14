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
func GetCallBack(data string) CallBack {
	callbacks := []CallBack{
		Buy, ConfirmClear, ToggleModeEdit, ConfirmEditProduct,
		ClearAll, Refresh, EditProduct, DeleteProduct, ConfirmLeaveFamily,
	}

	for _, c := range callbacks {
		if strings.HasPrefix(data, string(c)) {
			return c
		}
	}
	return Unknown
}
