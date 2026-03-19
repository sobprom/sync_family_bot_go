package repository

import (
	"log"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/jmoiron/sqlx"
)

type ProductRepositoryImpl struct {
	db *sqlx.DB
}

func NewProductRepository(db *sqlx.DB) ProductRepository {
	return &ProductRepositoryImpl{db: db}
}

func (repo *ProductRepositoryImpl) UpdateProductName(productId int64, productName string) (bool, error) {
	stmt := table.ShoppingList.UPDATE(table.ShoppingList.ProductName).
		SET(productName).
		WHERE(table.ShoppingList.ID.EQ(postgres.Int(productId)))

	res, err := stmt.Exec(repo.db)
	if err != nil {
		return false, err
	}

	count, _ := res.RowsAffected()
	return count > 0, err

}

func (repo *ProductRepositoryImpl) AddProducts(familyId int64, products []string) error {

	if len(products) == 0 {
		return nil
	}

	stmt := table.ShoppingList.INSERT(
		table.ShoppingList.FamilyID,
		table.ShoppingList.ProductName,
	)

	for _, name := range products {
		stmt = stmt.VALUES(postgres.Int(familyId), postgres.String(name))
	}

	_, err := stmt.ON_CONFLICT().DO_NOTHING().Exec(repo.db)

	return err

}

func (repo *ProductRepositoryImpl) GetAllProductsOrdered(familyId int64) ([]model.ShoppingList, error) {
	var products []model.ShoppingList

	stmt := table.ShoppingList.SELECT(table.ShoppingList.AllColumns).
		WHERE(table.ShoppingList.FamilyID.EQ(postgres.Int(familyId))).
		ORDER_BY(
			table.ShoppingList.IsBought.ASC(),
			table.ShoppingList.CreatedAt.DESC(),
		)

	// Выполняем запрос и сканируем результат в слайс структур
	err := stmt.Query(repo.db, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}

func (repo *ProductRepositoryImpl) DeleteAllByFamilyId(familyID int64) error {

	stmt := table.ShoppingList.DELETE().
		WHERE(table.ShoppingList.FamilyID.EQ(postgres.Int(familyID)))

	_, err := stmt.Exec(repo.db)
	if err != nil {
		log.Printf("❌ Ошибка при удалении продуктов семьи %d: %v", familyID, err)
		return err
	}

	return nil
}
