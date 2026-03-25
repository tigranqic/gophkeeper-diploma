-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS records;
DROP TABLE IF EXISTS users;
-- +goose StatementEnd
