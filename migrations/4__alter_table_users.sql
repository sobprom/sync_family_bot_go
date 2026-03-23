-- +goose Up
-- +goose StatementBegin
-- 1. Меняем тип существующей колонки на int8
ALTER TABLE family_sync.users
    ALTER COLUMN last_message_id TYPE int8;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
-- 1. Возвращаем тип int4 (может не сработать, если данные уже превысили лимит int4)
ALTER TABLE family_sync.users
    ALTER COLUMN last_message_id TYPE int4;
-- +goose StatementEnd