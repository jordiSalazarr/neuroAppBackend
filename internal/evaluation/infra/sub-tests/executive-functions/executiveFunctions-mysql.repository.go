package EFinfra

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"neuro.app.jordi/database/dbmodels"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
)

type ExecutivefunctionsMYSQLRepository struct {
	Exec boil.ContextExecutor
}
type MockExecutiveFunctionsRepository struct{}

var MockExecutiveFunctionsSubtests []EFdomain.ExecutiveFunctionsSubtest = []EFdomain.ExecutiveFunctionsSubtest{
	{
		PK:            "subtest1",
		EvauluationId: "eval1",
		NumberOfItems: 20,
		TotalClicks:   50,
		TotalErrors:   5,
		TotalCorrect:  15,
		TotalTime:     time.Duration(120 * time.Second),
		Type:          EFdomain.ExuctiveFunctionSubtestType("a"),
		Score: EFdomain.ExecutiveFunctionsScore{
			Score:          85,
			Accuracy:       0.75,
			SpeedIndex:     0.7,
			CommissionRate: 0.1,
			DurationSec:    120,
		},
	},
}

func NewExecutiveFunctionsSubtestMYSQLRepository(db *sql.DB) *ExecutivefunctionsMYSQLRepository {
	return &ExecutivefunctionsMYSQLRepository{
		Exec: db,
	}
}

func NewMockExecutiveFunctionsRepository() *MockExecutiveFunctionsRepository {
	return &MockExecutiveFunctionsRepository{}
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
		return []EFdomain.ExecutiveFunctionsSubtest{}, err
	}
	if len(dbExecutiveFunctionSubtest) < 2 {
		return []EFdomain.ExecutiveFunctionsSubtest{}, errors.New("no executive functions subtest found for evaluation")
	}
	for _, subtest := range dbExecutiveFunctionSubtest {
		out = append(out, dBToDomainExecutiveFunctions(subtest))
	}
	return out, nil
}

func (m *MockExecutiveFunctionsRepository) Save(ctx context.Context, subtest EFdomain.ExecutiveFunctionsSubtest) error {
	return nil
}
func (m *MockExecutiveFunctionsRepository) GetByID(ctx context.Context, id string) (EFdomain.ExecutiveFunctionsSubtest, error) {
	return MockExecutiveFunctionsSubtests[0], nil
}

func (m *MockExecutiveFunctionsRepository) GetByEvaluationID(ctx context.Context, evaluationID string) ([]EFdomain.ExecutiveFunctionsSubtest, error) {
	return MockExecutiveFunctionsSubtests, nil
}
