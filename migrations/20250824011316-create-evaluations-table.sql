-- +migrate Up
CREATE TABLE evaluations (
    id                CHAR(36) NOT NULL PRIMARY KEY,   -- UUID de la evaluación
    patient_name      VARCHAR(255) NOT NULL,
    patient_age       INT NOT NULL,
    specialist_mail   VARCHAR(255) NOT NULL,
    specialist_id     CHAR(36) NOT NULL,              -- referencia a users.id
    assistant_analysis TEXT,
    storage_url       TEXT,
    storage_key       VARCHAR(512),
    current_status    ENUM('CREATED','IN_PROGRESS','COMPLETED','CANCELLED','FAILED') NOT NULL DEFAULT 'CREATED',
    created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    
    CONSTRAINT fk_evaluations_specialist
        FOREIGN KEY (specialist_id) REFERENCES users(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS evaluations;
