-- +migrate Up
CREATE TABLE IF NOT EXISTS clock_draw_subtest_results (
  id CHAR(36) NOT NULL,
  evaluation_id CHAR(36) NOT NULL,
  pass TINYINT(1) NOT NULL,
  reasons JSON NOT NULL,
  center_x INT NOT NULL,
  center_y INT NOT NULL,
  radius DOUBLE NOT NULL,
  dial_circularity DOUBLE NOT NULL,
  minute_angle_deg DOUBLE NOT NULL,
  hour_angle_deg DOUBLE NOT NULL,
  expected_minute_angle DOUBLE NOT NULL,
  expected_hour_angle DOUBLE NOT NULL,
  minute_angular_error_deg DOUBLE NOT NULL,
  hour_angular_error_deg DOUBLE NOT NULL,
  created_at DATETIME(6) NOT NULL,
  updated_at DATETIME(6) NOT NULL,
  PRIMARY KEY (id),
  KEY idx_clock_eval (evaluation_id),
  CONSTRAINT chk_circularity CHECK (dial_circularity >= 0 AND dial_circularity <= 2)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- +migrate Down
DROP TABLE IF EXISTS clock_draw_subtest_results;
