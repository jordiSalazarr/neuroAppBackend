-- +migrate Up
CREATE TABLE IF NOT EXISTS visual_spatial_subtest (
  id            VARCHAR(36)  NOT NULL PRIMARY KEY,
  evaluation_id VARCHAR(36)  NOT NULL,
  score         TINYINT      NOT NULL CHECK (score BETWEEN 0 AND 5),
  note          VARCHAR(2500) NOT NULL,
  created_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
  updated_at    DATETIME(3)  NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3)
);

-- +migrate Down
DROP TABLE IF EXISTS visual_spatial_subtest;
