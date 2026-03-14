package repository

import "sync_family_bot_go/internal/gen/bots_go/family_sync/model"

type FamilyRepository interface {
	GetFamilyMemberByChatId(chatID int64) (*model.Users, error)
	GetFamilyMembersByFamilyId(familyID int64) ([]model.Users, error)
	CreateFamilyMember(chatID int64, userName string) (*model.Users, error)
	DropEditingProductId(chatID int64) error
	UpdateLastMessageIds(users []model.Users) error
}

type ProductRepository interface {
	UpdateProductName(productID int64, productName string) (bool, error)
	AddProducts(familyID int64, products []string) error
	GetAllProductsOrdered(familyID int64) ([]model.ShoppingList, error)
}
