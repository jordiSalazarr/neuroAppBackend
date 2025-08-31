-- +migrate Up
CREATE TABLE users (
    id          CHAR(36) NOT NULL PRIMARY KEY,      -- UUID generado en la app
    cognito_id  VARCHAR(255) NOT NULL UNIQUE,       -- sub de Cognito
    name        VARCHAR(255) NOT NULL,
    email       VARCHAR(320) NOT NULL UNIQUE,       -- RFC permite hasta 320 chars
    phone       VARCHAR(32),                        -- opcional
    is_active boolean not null default true,
    has_accepted_latest_terms boolean not null default false,
    created_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
                 ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS users;
