-- +migrate Up
CREATE TABLE if not exists executive_functions_subtests (
  id                CHAR(36) NOT NULL PRIMARY KEY,         -- PK (UUID from app)
  evaluation_id     CHAR(36) NOT NULL,                      -- FK to evaluations.id
  number_of_items   INT NOT NULL,
  total_clicks      INT NOT NULL,
  total_errors      INT NOT NULL,
  total_correct     INT NOT NULL,
  total_time_sec    DOUBLE NOT NULL,                         -- time.Duration stored in seconds
  type              VARCHAR(32) NOT NULL,                    -- ExuctiveFunctionSubtestType
  score             INT NOT NULL,                            -- ExecutiveFunctionsScore.Score
  accuracy          DOUBLE NOT NULL,
  speed_index       DOUBLE NOT NULL,
  commission_rate   DOUBLE NOT NULL,
  duration_sec      DOUBLE NOT NULL,                         -- duplicate of duration in score
  assistant_analysis TEXT NULL,
  created_at        TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_exec_fn_eval
    FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS executive_functions_subtests;
