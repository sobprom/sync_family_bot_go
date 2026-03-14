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
func (r *FamilyRepositoryImpl) GetFamilyMemberByChatId(chatID int64) (*model.Users, error) {

	stmt := table.Users.SELECT(table.Users.AllColumns).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID)))

	var user model.Users
	err := stmt.Query(r.db, &user)

	if err != nil {
		// Проверяем, что записи не найдены
		if errors.Is(err, qrm.ErrNoRows) {
			return nil, nil // пользователь не найден
		}
		return nil, err // реальная ошибка БД
	}

	return &user, nil
}

func (r *FamilyRepositoryImpl) GetFamilyMembersByFamilyId(familyId int64) ([]model.Users, error) {
	var users []model.Users

	stmt := table.Users.SELECT(table.Users.AllColumns).
		WHERE(table.Users.FamilyID.EQ(postgres.Int(familyId)))

	// Query сканирует результат напрямую в слайс структур
	err := stmt.Query(r.db, &users)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *FamilyRepositoryImpl) CreateFamilyMember(chatID int64, userName string) (*model.Users, error) {

	inviteCode := uuid.New().String()

	var family model.Families

	insertFamilyStmt := table.Families.INSERT(
		table.Families.InviteCode,
	).VALUES(
		inviteCode,
	).RETURNING(table.Families.ID)

	err := insertFamilyStmt.Query(r.db, &family)
	if err != nil {
		return nil, err
	}

	return r.upsertUserFamily(chatID, family.ID, userName)

}

func (r *FamilyRepositoryImpl) DropEditingProductId(chatID int64) error {
	updateStmt := table.Users.UPDATE(table.Users.EditingProductID).
		SET(nil).
		WHERE(table.Users.ChatID.EQ(postgres.Int(chatID)))

	_, err := updateStmt.Exec(r.db)
	return err
}

func (r *FamilyRepositoryImpl) upsertUserFamily(chatID int64, familyID int64, userName string) (*model.Users, error) {

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
	err := insertStmt.Query(r.db, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (p *FamilyRepositoryImpl) UpdateLastMessageIds(users []model.Users) error {
	if len(users) == 0 {
		return nil
	}

	// Начинаем транзакцию, чтобы пачка UPDATE прошла быстро
	tx, err := p.db.Begin()
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
