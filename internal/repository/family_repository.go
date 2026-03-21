package repository

import (
	"database/sql"
	"errors"
	"log"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/model"
	"sync_family_bot_go/internal/gen/bots_go/family_sync/table"

	"github.com/go-jet/jet/v2/postgres"
	"github.com/go-jet/jet/v2/qrm"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type FamilyRepositoryImpl struct {
	db *sqlx.DB
}

func NewFamilyRepository(db *sqlx.DB) FamilyRepository {
	return &FamilyRepositoryImpl{db: db}
}

// GetFamilyMemberByChatId возвращает пользователя по chat_id.
// Если пользователь не найден, возвращает nil, nil.
// В случае ошибки БД возвращает ошибку.
func (repo *FamilyRepositoryImpl) GetFamilyMemberByChatId(chatID int64) (*model.Users, error) {

	stmt := table.Users.SELECT(table.Users.AllColumns).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID)))

	var user model.Users
	err := stmt.Query(repo.db, &user)

	if err != nil {
		// Проверяем, что записи не найдены
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil // пользователь не найден
		}
		return nil, err // реальная ошибка БД
	}

	return &user, nil
}

func (repo *FamilyRepositoryImpl) GetFamilyMembersByFamilyId(familyId int64) ([]model.Users, error) {
	var users []model.Users

	stmt := table.Users.SELECT(table.Users.AllColumns).
		WHERE(table.Users.FamilyID.EQ(postgres.Int(familyId)))

	// Query сканирует результат напрямую в слайс структур
	err := stmt.Query(repo.db, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (repo *FamilyRepositoryImpl) CreateFamilyMember(chatID int64, userName string) (*model.Users, error) {

	inviteCode := uuid.New().String()

	var family model.Families

	insertFamilyStmt := table.Families.INSERT(
		table.Families.InviteCode,
	).VALUES(
		inviteCode,
	).RETURNING(table.Families.ID)

	err := insertFamilyStmt.Query(repo.db, &family)
	if err != nil {
		return nil, err
	}

	return repo.upsertUserFamily(chatID, family.ID, userName)

}

func (repo *FamilyRepositoryImpl) DropEditingProductId(chatID int64) error {
	updateStmt := table.Users.UPDATE(table.Users.EditingProductID).
		SET(nil).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID)))

	_, err := updateStmt.Exec(repo.db)
	return err
}

func (repo *FamilyRepositoryImpl) UpdateLastMessageIds(users []model.Users) error {
	if len(users) == 0 {
		return nil
	}

	// Начинаем транзакцию, чтобы пачка UPDATE прошла быстро
	tx, err := repo.db.Begin()
	if err != nil {
		return err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			// Логируем ошибку, но не прерываем выполнение
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	for _, user := range users {

		if user.LastMessageID == nil {
			// Пропускаем пользователей с nil значением или обрабатываем по-другому
			continue
		}
		// Формируем запрос для каждого пользователя
		stmt := table.Users.UPDATE(table.Users.LastMessageID).
			SET(postgres.Int(*user.LastMessageID)).
			WHERE(table.Users.ChatID.EQ(postgres.Int(user.ChatID)))

		_, err := stmt.Exec(tx)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (repo *FamilyRepositoryImpl) JoinFamily(chatID int64,
	code string,
	userName string) (*model.Users, error) {

	// Начинаем транзакцию, чтобы обеспечить атомарность операции
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := tx.Rollback(); err != nil && !errors.Is(err, sql.ErrTxDone) {
			log.Printf("failed to rollback transaction: %v", err)
		}
	}()

	// 1. Ищем семью по invite_code
	var family model.Families
	selectFamilyStmt := table.Families.SELECT(table.Families.ID).
		WHERE(table.Families.InviteCode.EQ(postgres.String(code)))

	err = selectFamilyStmt.Query(tx, &family)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			// Семья с таким кодом не найдена
			return nil, nil
		}
		return nil, err
	}

	// 2. Добавляем пользователя в семью
	user, err := repo.upsertUserFamily(chatID, family.ID, userName)
	if err != nil {
		return nil, err
	}

	// 3. Фиксируем транзакцию
	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return user, nil

}

func (repo *FamilyRepositoryImpl) GetFamilyCode(familyID int64) (string, error) {
	var family model.Families

	stmt := table.Families.SELECT(table.Families.InviteCode).
		WHERE(table.Families.ID.EQ(postgres.Int(familyID)))

	err := stmt.Query(repo.db, &family)
	if err != nil {
		if errors.Is(err, qrm.ErrNoRows) {
			return "", nil // семья не найдена
		}
		return "", err // реальная ошибка БД
	}

	return family.InviteCode, nil
}

func (repo *FamilyRepositoryImpl) DropShoppingEditMode(chatID int64) error {
	updateStmt := table.Users.UPDATE(table.Users.ShoppingListEditMode).
		SET(postgres.Bool(false)).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID)))

	// Выполняем запрос
	_, err := updateStmt.Exec(repo.db)
	if err != nil {
		log.Printf("failed to drop shopping edit mode for chat %d: %v", chatID, err)
		return err
	}

	return nil
}

func (repo *FamilyRepositoryImpl) ToggleShoppingEditMode(chatID int64) (*model.Users, error) {

	stmt := table.Users.UPDATE(table.Users.ShoppingListEditMode).
		SET(
			postgres.Bool(true),
		).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID))).
		RETURNING(table.Users.AllColumns)

	var updatedUser model.Users
	err := stmt.Query(repo.db, &updatedUser)
	if err != nil {
		return nil, err
	}

	return &updatedUser, nil
}

func (repo *FamilyRepositoryImpl) SetEditingProductId(chatID int64, productID int64) error {
	updateStmt := table.Users.UPDATE(
		table.Users.EditingProductID,
	).SET(
		postgres.Int(productID),
	).WHERE(
		table.Users.ChatID.EQ(postgres.Int(chatID)),
	)

	_, err := updateStmt.Exec(repo.db)
	if err != nil {
		log.Printf("❌ Ошибка установки editing_product_id для chat %d: %v", chatID, err)
		return err
	}

	return nil
}

func (repo *FamilyRepositoryImpl) upsertUserFamily(chatID int64,
	familyID int64,
	userName string) (*model.Users, error) {

	insertStmt := table.Users.INSERT(
		table.Users.ChatID,
		table.Users.FamilyID,
		table.Users.Username,
	).VALUES(
		chatID,
		familyID,
		userName,
	).ON_CONFLICT(table.Users.ChatID).DO_UPDATE(
		postgres.SET(
			table.Users.FamilyID.SET(postgres.Int(familyID)),
			table.Users.Username.SET(postgres.String(userName)),
		),
	).RETURNING(table.Users.AllColumns)

	var user model.Users
	err := insertStmt.Query(repo.db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}
