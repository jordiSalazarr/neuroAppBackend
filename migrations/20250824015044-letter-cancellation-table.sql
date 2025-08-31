-- +migrate Up
CREATE TABLE letters_cancellation_subtests (
    id                  CHAR(36) NOT NULL PRIMARY KEY, -- UUID del subtest
    evaluation_id       CHAR(36) NOT NULL,             -- FK a evaluations.id
    total_targets       INT NOT NULL,
    correct             INT NOT NULL,
    errors              INT NOT NULL,
    time_in_secs        INT NOT NULL,
    assistant_analysis  TEXT,
    score               INT NOT NULL,
    cp_per_min          DOUBLE NOT NULL,
    accuracy            DOUBLE NOT NULL,
    omissions           INT NOT NULL,
    omissions_rate      DOUBLE NOT NULL,
    commission_rate     DOUBLE NOT NULL,
    hits_per_min        DOUBLE NOT NULL,
    errors_per_min      DOUBLE NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_letters_cancellation_eval
        FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
        ON DELETE CASCADE
        ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS letters_cancellation_subtests;
