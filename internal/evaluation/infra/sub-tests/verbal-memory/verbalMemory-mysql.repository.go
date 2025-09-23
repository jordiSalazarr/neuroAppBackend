package VEMinfra

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"neuro.app.jordi/database/dbmodels"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
)

type VerbalMemoryMYSQLRepository struct {
	Exec boil.ContextExecutor
}

type MockVerbalMemoryRepository struct{}

var MockVerbalMemorySubtests []*VEMdomain.VerbalMemorySubtest = []*VEMdomain.VerbalMemorySubtest{
	{
		Pk:               "subtest1",
		EvaluationID:     "eval1",
		SecondsFromStart: 30,
		GivenWords:       []string{"apple", "banana", "cherry"},
		RecalledWords:    []string{"apple", "cherry"},
		Type:             VEMdomain.VerbalMemorySubtypeImmediate,
		Score: VEMdomain.VerbalMemoryScore{
			Score:             66,
			Hits:              2,
			Omissions:         1,
			Intrusions:        0,
			Perseverations:    0,
			Accuracy:          0.67,
			IntrusionRate:     0.0,
			PerseverationRate: 0.0,
		},
		AssistanAnalysis: "Good recall ability.",
	},
	{
		Pk:               "subtest2",
		EvaluationID:     "eval2",
		SecondsFromStart: 30,
		GivenWords:       []string{"dog", "cat", "mouse"},
		RecalledWords:    []string{"dog", "cat", "elephant"},
		Type:             VEMdomain.VerbalMemorySubtypeImmediate,
		Score: VEMdomain.VerbalMemoryScore{
			Score:             66,
			Hits:              2,
			Omissions:         1,
			Intrusions:        1,
			Perseverations:    0,
			Accuracy:          0.67,
			IntrusionRate:     0.33,
			PerseverationRate: 0.0,
		},
		AssistanAnalysis: "Average recall with some intrusions.",
	},
}

func NewVerbalMemoryMYSQLRepository(db *sql.DB) *VerbalMemoryMYSQLRepository {
	return &VerbalMemoryMYSQLRepository{
		Exec: db,
	}
}

func NewMockVerbalMemoryRepository() MockVerbalMemoryRepository {
	return MockVerbalMemoryRepository{}
}

func DBToDomainVerbalMemory(m *dbmodels.VerbalMemorySubtest) (VEMdomain.VerbalMemorySubtest, error) {
	var given []string
	_ = json.Unmarshal(m.GivenWords, &given)

	var recalled []string
	_ = json.Unmarshal(m.GivenWords, &recalled)

	vm := VEMdomain.VerbalMemorySubtest{
		Pk:               m.ID,
		SecondsFromStart: m.SecondsFromStart,
		GivenWords:       given,
		RecalledWords:    recalled,
		Type:             VEMdomain.VerbalMemorySubtype(m.Type),
		EvaluationID:     m.EvaluationID,
		Score: VEMdomain.VerbalMemoryScore{
			Score:             m.ScoreScore,
			Hits:              m.ScoreHits,
			Omissions:         m.ScoreOmissions,
			Intrusions:        m.ScoreIntrusions,
			Perseverations:    m.ScorePerseverations,
			Accuracy:          m.ScoreAccuracy,
			IntrusionRate:     m.ScoreIntrusionRate,
			PerseverationRate: m.ScorePerseverationRate,
		},
		AssistanAnalysis: m.AssistanAnalysis,
		CreatedAt:        m.CreatedAt,
	}
	return vm, nil
}
func strSliceToJSON(sl []string) (types.JSON, error) {
	b, err := json.Marshal(sl)
	if err != nil {
		return nil, fmt.Errorf("marshal []string: %w", err)
	}
	return types.JSON(b), nil
}
func DomainToDBVerbalMemory(d VEMdomain.VerbalMemorySubtest) (*dbmodels.VerbalMemorySubtest, error) {
	given, err := strSliceToJSON(d.GivenWords)
	if err != nil {
		return nil, fmt.Errorf("given_words: %w", err)
	}
	recalled, err := strSliceToJSON(d.RecalledWords)
	if err != nil {
		return nil, fmt.Errorf("recalled_words: %w", err)
	}

	return &dbmodels.VerbalMemorySubtest{
		ID:                     d.Pk,
		EvaluationID:           d.EvaluationID,
		SecondsFromStart:       d.SecondsFromStart,
		Type:                   string(d.Type),
		GivenWords:             given,
		RecalledWords:          recalled,
		ScoreScore:             d.Score.Score,
		ScoreHits:              d.Score.Hits,
		ScoreOmissions:         d.Score.Omissions,
		ScoreIntrusions:        d.Score.Intrusions,
		ScorePerseverations:    d.Score.Perseverations,
		ScoreAccuracy:          d.Score.Accuracy,
		ScoreIntrusionRate:     d.Score.IntrusionRate,
		ScorePerseverationRate: d.Score.PerseverationRate,
		AssistanAnalysis:       d.AssistanAnalysis,
		CreatedAt:              d.CreatedAt,
	}, nil
}

func (r VerbalMemoryMYSQLRepository) Save(ctx context.Context, subtest VEMdomain.VerbalMemorySubtest) error {
	dbVerbalMemorySubtest, err := DomainToDBVerbalMemory(subtest)
	if err != nil {
		return err
	}
	return dbVerbalMemorySubtest.Insert(ctx, r.Exec, boil.Infer())
}

func (r VerbalMemoryMYSQLRepository) GetByID(ctx context.Context, id string) (VEMdomain.VerbalMemorySubtest, error) {
	dbVerbalMemorySubtest, err := dbmodels.VerbalMemorySubtests(
		dbmodels.VerbalMemorySubtestWhere.ID.EQ(id),
	).One(ctx, r.Exec)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, nil
	}

	return DBToDomainVerbalMemory(dbVerbalMemorySubtest)
}

func (r VerbalMemoryMYSQLRepository) GetByEvaluationID(ctx context.Context, id string) (VEMdomain.VerbalMemorySubtest, error) {
	dbVerbalMemorySubtest, err := dbmodels.VerbalMemorySubtests(
		dbmodels.VerbalMemorySubtestWhere.EvaluationID.EQ(id),
	).One(ctx, r.Exec)
	if err != nil {
		return VEMdomain.VerbalMemorySubtest{}, nil
	}

	return DBToDomainVerbalMemory(dbVerbalMemorySubtest)
}

func (r MockVerbalMemoryRepository) Save(ctx context.Context, subtest VEMdomain.VerbalMemorySubtest) error {
	return nil
}

func (r MockVerbalMemoryRepository) GetByID(ctx context.Context, id string) (VEMdomain.VerbalMemorySubtest, error) {
	return *MockVerbalMemorySubtests[0], nil
}

func (r MockVerbalMemoryRepository) GetByEvaluationID(ctx context.Context, id string) (VEMdomain.VerbalMemorySubtest, error) {
	return *MockVerbalMemorySubtests[0], nil
}
