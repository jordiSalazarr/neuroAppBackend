-- +migrate Up
CREATE TABLE language_fluencies (
  id                 CHAR(36) NOT NULL PRIMARY KEY,          -- PK (UUID from app)
  evaluation_id      CHAR(36) NOT NULL,                      -- FK to evaluations.id
  language           VARCHAR(64)  NOT NULL,
  proficiency        VARCHAR(64)  NOT NULL,
  category           VARCHAR(128) NOT NULL,
  answer_words       JSON NULL,                               -- []string as JSON
  score              INT NOT NULL,                            -- LanguageFluencyScore.Score
  unique_valid       INT NOT NULL,
  intrusions         INT NOT NULL,
  perseverations     INT NOT NULL,
  total_produced     INT NOT NULL,
  words_per_minute   DOUBLE NOT NULL,
  intrusion_rate     DOUBLE NOT NULL,
  persev_rate        DOUBLE NOT NULL,
  assistant_analysis TEXT NULL,
  created_at         TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

  CONSTRAINT fk_lang_fluency_eval
    FOREIGN KEY (evaluation_id) REFERENCES evaluations(id)
    ON DELETE CASCADE
    ON UPDATE CASCADE
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

-- +migrate Down
DROP TABLE IF EXISTS language_fluencies;
