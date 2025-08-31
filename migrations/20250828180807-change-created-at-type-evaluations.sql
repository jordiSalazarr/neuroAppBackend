-- +migrate Up
ALTER TABLE evaluations
MODIFY COLUMN created_at DATE NOT NULL;

-- +migrate Down
ALTER TABLE evaluations
MODIFY COLUMN created_at DATETIME NOT NULL;
