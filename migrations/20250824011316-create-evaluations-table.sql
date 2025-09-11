-- +migrate Up
CREATE TABLE evaluations (
  id                CHAR(36) NOT NULL PRIMARY KEY, -- assuming UUID/char id from sqlboiler string
  patient_name      VARCHAR(255) NOT NULL,
  patient_age       INT NOT NULL,
  specialist_mail   VARCHAR(255) NOT NULL,
  specialist_id     CHAR(36) NOT NULL,
  assistant_analysis TEXT NULL,
  storage_url       TEXT NULL,
  storage_key       TEXT NULL,
  current_status    VARCHAR(50) NOT NULL,
  created_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at        DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS evaluations;
