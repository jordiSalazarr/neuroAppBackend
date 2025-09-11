
-- +migrate Up
CREATE TABLE IF NOT EXISTS visual_memory_subtests(
  id              VARCHAR(36)  NOT NULL,
  evaluation_id   VARCHAR(36)  NOT NULL,
  score           int       NOT NULL,
  note         TEXT         NOT NULL,
  image_src          VARCHAR(255)          NULL,
  created_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_eval (evaluation_id),
  CONSTRAINT fk_visualmemory_evaluation FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down

DROP TABLE IF EXISTS visual_memory_subtests;
