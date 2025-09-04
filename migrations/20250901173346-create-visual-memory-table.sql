
-- +migrate Up
CREATE TABLE IF NOT EXISTS visual_memory_subtest(
  id              VARCHAR(36)  NOT NULL,
  evaluation_id   VARCHAR(36)  NOT NULL,
  shape           ENUM('circle','square','triangle') NOT NULL,
  pass            TINYINT(1)   NOT NULL,
  score           DOUBLE       NOT NULL,
  reasons         TEXT         NOT NULL, -- JSON []
  iou             DOUBLE       NULL,
  circularity     DOUBLE       NULL,
  angle_rmse      DOUBLE       NULL,
  side_cv         DOUBLE       NULL,
  center_x        INT          NULL,
  center_y        INT          NULL,
  bbox_w          INT          NULL,
  bbox_h          INT          NULL,
  created_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6),
  updated_at      DATETIME(6)  NOT NULL DEFAULT CURRENT_TIMESTAMP(6) ON UPDATE CURRENT_TIMESTAMP(6),
  PRIMARY KEY (id),
  KEY idx_eval (evaluation_id),
  CONSTRAINT fk_shape_eval FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
    ON DELETE CASCADE ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down

DROP TABLE IF EXISTS visual_memory_subtest;
