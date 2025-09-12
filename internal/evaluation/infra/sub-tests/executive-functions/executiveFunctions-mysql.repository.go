package EFinfra

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"neuro.app.jordi/database/dbmodels"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
)

type ExecutivefunctionsMYSQLRepository struct {
	Exec boil.ContextExecutor
}

func NewExecutiveFunctionsSubtestMYSQLRepository(db *sql.DB) *ExecutivefunctionsMYSQLRepository {
	return &ExecutivefunctionsMYSQLRepository{
		Exec: db,
	}
}

func domainToDBExecutiveFunctions(s EFdomain.ExecutiveFunctionsSubtest) *dbmodels.ExecutiveFunctionsSubtest {
	return &dbmodels.ExecutiveFunctionsSubtest{
		ID:                s.PK,
		EvaluationID:      s.EvauluationId,
		NumberOfItems:     s.NumberOfItems,
		TotalClicks:       s.TotalClicks,
		TotalErrors:       s.TotalErrors,
		TotalCorrect:      s.TotalCorrect,
		TotalTimeSec:      float64(s.Score.Score),
		Type:              string(s.Type),
		Score:             s.Score.Score,
		Accuracy:          s.Score.Accuracy,
		SpeedIndex:        s.Score.SpeedIndex,
		CommissionRate:    s.Score.CommissionRate,
		DurationSec:       s.Score.DurationSec,
		AssistantAnalysis: null.StringFrom(s.AssistanAnalys),
		CreatedAt:         s.CreatedAt,
	}
}

// DB (sqlboiler) -> Domain
func dBToDomainExecutiveFunctions(m *dbmodels.ExecutiveFunctionsSubtest) EFdomain.ExecutiveFunctionsSubtest {
	d := EFdomain.ExecutiveFunctionsSubtest{
		PK:            m.ID,
		EvauluationId: m.EvaluationID,
		NumberOfItems: m.NumberOfItems,
		TotalClicks:   m.TotalClicks,
		TotalErrors:   m.TotalErrors,
		TotalCorrect:  m.TotalCorrect,
		TotalTime:     time.Duration(m.TotalTimeSec * float64(time.Second)), // seconds -> duration
		Type:          EFdomain.ExuctiveFunctionSubtestType(m.Type),
		Score: EFdomain.ExecutiveFunctionsScore{
			Score:          m.Score,
			Accuracy:       m.Accuracy,
			SpeedIndex:     m.SpeedIndex,
			CommissionRate: m.CommissionRate,
			DurationSec:    m.DurationSec,
		},
		AssistanAnalys: m.AssistantAnalysis.String,
		CreatedAt:      m.CreatedAt,
	}
	return d
}

func (m *ExecutivefunctionsMYSQLRepository) Save(ctx context.Context, subtest EFdomain.ExecutiveFunctionsSubtest) error {
	dbExecutiveFunctionSubtest := domainToDBExecutiveFunctions(subtest)
	return dbExecutiveFunctionSubtest.Insert(ctx, m.Exec, boil.Infer())
}
func (m *ExecutivefunctionsMYSQLRepository) GetByID(ctx context.Context, id string) (EFdomain.ExecutiveFunctionsSubtest, error) {
	dbExecutiveFunctionSubtest, err := dbmodels.ExecutiveFunctionsSubtests(
		dbmodels.ExecutiveFunctionsSubtestWhere.ID.EQ(id),
	).One(ctx, m.Exec)
	if err != nil {
		return EFdomain.ExecutiveFunctionsSubtest{}, nil
	}
	return dBToDomainExecutiveFunctions(dbExecutiveFunctionSubtest), nil
}

func (m *ExecutivefunctionsMYSQLRepository) GetByEvaluationID(ctx context.Context, evaluationID string) ([]EFdomain.ExecutiveFunctionsSubtest, error) {
	var out []EFdomain.ExecutiveFunctionsSubtest
	dbExecutiveFunctionSubtest, err := dbmodels.ExecutiveFunctionsSubtests(
		dbmodels.ExecutiveFunctionsSubtestWhere.EvaluationID.EQ(evaluationID),
	).All(ctx, m.Exec)
	if err != nil {
		return []EFdomain.ExecutiveFunctionsSubtest{}, nil
	}
	for _, subtest := range dbExecutiveFunctionSubtest {
		out = append(out, dBToDomainExecutiveFunctions(subtest))
	}
	return out, nil
}
