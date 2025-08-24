package LCinfra

import (
	"context"
	"database/sql"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"neuro.app.jordi/database/dbmodels"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
)

type LetterCancellationMYSQLRepository struct {
	Exec boil.ContextExecutor
}

func NewInMemoryLetterCancellationRepository(db *sql.DB) *LetterCancellationMYSQLRepository {
	return &LetterCancellationMYSQLRepository{
		Exec: db,
	}
}

func domainToDBLetterCancellation(subtest LCdomain.LettersCancellationSubtest) *dbmodels.LettersCancellationSubtest {
	return &dbmodels.LettersCancellationSubtest{
		ID:                subtest.PK,
		EvaluationID:      subtest.EvaluationID,
		TotalTargets:      subtest.TotalTargets,
		Correct:           subtest.Correct,
		Errors:            subtest.Errors,
		TimeInSecs:        subtest.TimeInSecs,
		AssistantAnalysis: null.StringFrom(subtest.AssistantAnalysis),
		Score:             subtest.CancellationScore.Score,
		CPPerMin:          subtest.CancellationScore.CpPerMin,
		Accuracy:          subtest.CancellationScore.Accuracy,
		Omissions:         subtest.CancellationScore.Omissions,
		OmissionsRate:     subtest.CancellationScore.OmissionsRate,
		CommissionRate:    subtest.CancellationScore.CommissionRate,
		HitsPerMin:        subtest.CancellationScore.HitsPerMin,
		ErrorsPerMin:      subtest.CancellationScore.ErrorsPerMin,
		CreatedAt:         subtest.CreatedAt,
	}
}

func dbToDomainLetterCancellation(model *dbmodels.LettersCancellationSubtest) LCdomain.LettersCancellationSubtest {
	return LCdomain.LettersCancellationSubtest{
		PK:                model.ID,
		EvaluationID:      model.EvaluationID,
		TotalTargets:      model.TotalTargets,
		Correct:           model.Correct,
		Errors:            model.Errors,
		TimeInSecs:        model.TimeInSecs,
		AssistantAnalysis: model.AssistantAnalysis.String,
		CreatedAt:         model.CreatedAt,
		CancellationScore: LCdomain.CancellationScore{
			Score:          model.Score,
			CpPerMin:       model.CPPerMin,
			Accuracy:       model.Accuracy,
			Omissions:      model.Omissions,
			OmissionsRate:  model.OmissionsRate,
			CommissionRate: model.CommissionRate,
			HitsPerMin:     model.HitsPerMin,
			ErrorsPerMin:   model.ErrorsPerMin,
		},
	}
}

func (repo *LetterCancellationMYSQLRepository) Save(ctx context.Context, subtest *LCdomain.LettersCancellationSubtest) error {
	dbLetterCancellation := domainToDBLetterCancellation(*subtest)
	return dbLetterCancellation.Insert(ctx, repo.Exec, boil.Infer())
}

func (repo *LetterCancellationMYSQLRepository) GetByEvaluationID(ctx context.Context, evaluationID string) (LCdomain.LettersCancellationSubtest, error) {
	dbLetterCancellation, err := dbmodels.LettersCancellationSubtests(
		dbmodels.LettersCancellationSubtestWhere.EvaluationID.EQ(evaluationID),
	).One(ctx, repo.Exec)
	if err != nil {
		return LCdomain.LettersCancellationSubtest{}, err
	}

	return dbToDomainLetterCancellation(dbLetterCancellation), nil

}
