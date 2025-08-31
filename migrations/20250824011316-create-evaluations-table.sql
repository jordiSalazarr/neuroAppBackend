-- 20250826003430-create-verbalMemory-table.sql
-- +migrate Up
CREATE TABLE verbal_memory_subtests (
  pk                          CHAR(36)        NOT NULL,
  evaluation_id               CHAR(36)        NOT NULL,
  seconds_from_start          BIGINT          NOT NULL,
  type                        VARCHAR(32)     NOT NULL,
  given_words                 JSON            NOT NULL,
  recalled_words              JSON            NOT NULL,

  score_score                 INT             NOT NULL,
  score_hits                  INT             NOT NULL,
  score_omissions             INT             NOT NULL,
  score_intrusions            INT             NOT NULL,
  score_perseverations        INT             NOT NULL,
  score_accuracy              DOUBLE          NOT NULL,
  score_intrusion_rate        DOUBLE          NOT NULL,
  score_perseveration_rate    DOUBLE          NOT NULL,

  assistan_analysis           MEDIUMTEXT      NOT NULL,
  created_at                  DATETIME        NOT NULL DEFAULT CURRENT_TIMESTAMP,

  PRIMARY KEY (pk),
  KEY idx_vms_eval (evaluation_id, created_at),
  CONSTRAINT fk_vms_eval
    FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
    ON DELETE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS verbal_memory_subtests;
