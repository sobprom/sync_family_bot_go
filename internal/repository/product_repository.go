package repository

import (
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

func (p *ProductRepositoryImpl) UpdateProductName(productId int64, productName string) (bool, error) {
	stmt := table.ShoppingList.UPDATE(table.ShoppingList.ProductName).
		SET(productName).
		WHERE(table.ShoppingList.ID.EQ(postgres.Int(productId)))

	res, err := stmt.Exec(p.db)
	if err != nil {
		return false, err
	}

	count, _ := res.RowsAffected()
	return count > 0, err

}

func (p *ProductRepositoryImpl) AddProducts(familyId int64, products []string) error {

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

	_, err := stmt.ON_CONFLICT().DO_NOTHING().Exec(p.db)

	return err

}

func (p *ProductRepositoryImpl) GetAllProductsOrdered(familyId int64) ([]model.ShoppingList, error) {
	var products []model.ShoppingList

	stmt := table.ShoppingList.SELECT(table.ShoppingList.AllColumns).
		WHERE(table.ShoppingList.FamilyID.EQ(postgres.Int(familyId))).
		ORDER_BY(
			table.ShoppingList.IsBought.ASC(),
			table.ShoppingList.CreatedAt.DESC(),
		)

	// Выполняем запрос и сканируем результат в слайс структур
	err := stmt.Query(p.db, &products)
	if err != nil {
		return nil, err
	}

	return products, nil
}
