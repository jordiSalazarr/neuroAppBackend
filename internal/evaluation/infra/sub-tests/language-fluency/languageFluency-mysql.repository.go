package LFinfra

import (
	"context"
	"database/sql"
	"encoding/json"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/types"
	"github.com/aarondl/sqlboiler/v4/boil"
	"neuro.app.jordi/database/dbmodels"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
)

type LanguageFluencyMYSQLRepository struct {
	Exec boil.ContextExecutor
}

func NewLanguageFluencyMYSQLRepository(db *sql.DB) *LanguageFluencyMYSQLRepository {
	return &LanguageFluencyMYSQLRepository{
		Exec: db,
	}
}

func DomainToDBLanguageFluency(s LFdomain.LanguageFluency) *dbmodels.LanguageFluency {
	// encode []string -> JSON
	var aw types.JSON
	if len(s.AnswerWords) > 0 {
		if b, err := json.Marshal(s.AnswerWords); err == nil {
			aw = types.JSON(b)
		} else {
			// if you prefer, handle error upstream; here we leave it nil
			aw = nil
		}
	}

	return &dbmodels.LanguageFluency{
		ID:                s.PK,
		EvaluationID:      s.EvaluationID,
		Language:          s.Language,
		Proficiency:       s.Proficiency,
		Category:          s.Category,
		AnswerWords:       null.JSONFrom(aw), // JSON column
		Score:             s.Score.Score,
		UniqueValid:       s.Score.UniqueValid,
		Intrusions:        s.Score.Intrusions,
		Perseverations:    s.Score.Perseverations,
		TotalProduced:     s.Score.TotalProduced,
		WordsPerMinute:    s.Score.WordsPerMinute,
		IntrusionRate:     s.Score.IntrusionRate,
		PersevRate:        s.Score.PersevRate,
		AssistantAnalysis: null.StringFrom(s.AssistantAnalysis),
		CreatedAt:         s.CreatedAt, // if your column is TIMESTAMP NOT NULL
	}
}

// ------------------------------
// DB (sqlboiler model) -> Domain
// ------------------------------
func DBToDomainLanguageFluency(m *dbmodels.LanguageFluency) LFdomain.LanguageFluency {
	var words []string
	if m.AnswerWords.Valid { // m.AnswerWords es null.JSON
		_ = json.Unmarshal(m.AnswerWords.JSON, &words) // JSON ([]byte) -> []string
	}
	return LFdomain.LanguageFluency{
		PK:           m.ID,
		Language:     m.Language,
		Proficiency:  m.Proficiency,
		Category:     m.Category,
		AnswerWords:  words,
		EvaluationID: m.EvaluationID,
		Score: LFdomain.LanguageFluencyScore{
			Score:          m.Score,
			UniqueValid:    m.UniqueValid,
			Intrusions:     m.Intrusions,
			Perseverations: m.Perseverations,
			TotalProduced:  m.TotalProduced,
			WordsPerMinute: m.WordsPerMinute,
			IntrusionRate:  m.IntrusionRate,
			PersevRate:     m.PersevRate,
		},
		AssistantAnalysis: m.AssistantAnalysis.String,
		CreatedAt:         m.CreatedAt,
	}
}

func (repo *LanguageFluencyMYSQLRepository) Save(ctx context.Context, lf LFdomain.LanguageFluency) error {
	dbLanguageFLuency := DomainToDBLanguageFluency(lf)
	return dbLanguageFLuency.Insert(ctx, repo.Exec, boil.Infer())
}
func (repo *LanguageFluencyMYSQLRepository) GetByID(ctx context.Context, id string) (LFdomain.LanguageFluency, error) {
	dbLanguageFluency, err := dbmodels.LanguageFluencies(
		dbmodels.LanguageFluencyWhere.ID.EQ(id),
	).One(ctx, repo.Exec)
	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}
	return DBToDomainLanguageFluency(dbLanguageFluency), nil
}

func (repo *LanguageFluencyMYSQLRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (LFdomain.LanguageFluency, error) {
	dbLanguageFluency, err := dbmodels.LanguageFluencies(
		dbmodels.LanguageFluencyWhere.EvaluationID.EQ(evaluationID),
	).One(ctx, repo.Exec)
	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}
	return DBToDomainLanguageFluency(dbLanguageFluency), nil
}
