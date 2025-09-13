package infra

import (
	"context"
	"database/sql"
	"time"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
	"neuro.app.jordi/database/dbmodels"
	"neuro.app.jordi/internal/evaluation/domain"
)

type EvaluationsMYSQLRepository struct {
	Exec boil.ContextExecutor
}

// NewUsersMYSQLRepository crea una nueva instancia de UsersMYSQLRepository
func NewEvaluationsMYSQLRepository(db *sql.DB) *EvaluationsMYSQLRepository {
	return &EvaluationsMYSQLRepository{Exec: db}
}

func domainEvaluationToDB(evaluation domain.Evaluation) *dbmodels.Evaluation {
	return &dbmodels.Evaluation{
		ID:                evaluation.PK,
		AssistantAnalysis: null.StringFrom(evaluation.AssistantAnalysis),
		PatientName:       evaluation.PatientName,
		PatientAge:        evaluation.PatientAge,
		SpecialistMail:    evaluation.SpecialistMail,
		SpecialistID:      evaluation.SpecialistID,
		StorageURL:        null.StringFrom(evaluation.StorageURL),
		StorageKey:        null.StringFrom(evaluation.StorageKey),
		CreatedAt:         evaluation.CreatedAt,
		CurrentStatus:     string(evaluation.CurrentStatus),
	}
}

func dbEvaluationToDomain(evaluation *dbmodels.Evaluation) domain.Evaluation {
	return domain.Evaluation{
		PK:                evaluation.ID,
		PatientName:       evaluation.PatientName,
		PatientAge:        evaluation.PatientAge,
		SpecialistMail:    evaluation.SpecialistMail,
		SpecialistID:      evaluation.SpecialistID,
		CurrentStatus:     domain.EvaluationCurrentStatus(evaluation.CurrentStatus),
		AssistantAnalysis: evaluation.AssistantAnalysis.String,
		StorageURL:        evaluation.StorageKey.String,
		StorageKey:        evaluation.StorageKey.String,
		CreatedAt:         evaluation.CreatedAt,
	}
}

func (m *EvaluationsMYSQLRepository) Save(ctx context.Context, evaluation domain.Evaluation) error {
	dbEvaluation := domainEvaluationToDB(evaluation)
	return dbEvaluation.Insert(ctx, m.Exec, boil.Infer())
}

func (m *EvaluationsMYSQLRepository) Update(ctx context.Context, evaluation domain.Evaluation) error {
	dbEvaluation, err := dbmodels.Evaluations(dbmodels.EvaluationWhere.ID.EQ(evaluation.PK)).One(ctx, m.Exec)
	if err != nil {
		return err
	}
	//TODO: here we should add the fields we want to update
	dbEvaluation.CurrentStatus = string(evaluation.CurrentStatus)
	dbEvaluation.AssistantAnalysis = null.StringFrom(evaluation.AssistantAnalysis)
	_, err = dbEvaluation.Update(ctx, m.Exec, boil.Infer())
	return err
}
func (m *EvaluationsMYSQLRepository) GetByID(ctx context.Context, id string) (domain.Evaluation, error) {
	dbEvaluation, err := dbmodels.Evaluations(dbmodels.EvaluationWhere.ID.EQ(id)).One(ctx, m.Exec)
	if err != nil {
		return domain.Evaluation{}, err
	}
	domainEvaluation := dbEvaluationToDomain(dbEvaluation)
	return domainEvaluation, nil

}

func (f *EvaluationsMYSQLRepository) GetMany(ctx context.Context, fromDate, toDate time.Time, offset, limit int, searchTerm string, specialist_id string, onlyCompleted bool) ([]*domain.Evaluation, error) {

	var query []qm.QueryMod
	defaultLimit := 30
	if limit > defaultLimit {
		limit = defaultLimit
	}
	if specialist_id != "" {
	}
	query = append(query,
		dbmodels.EvaluationWhere.CreatedAt.GTE(fromDate),
		dbmodels.EvaluationWhere.CreatedAt.LTE(toDate),
		qm.Offset(offset),
		qm.Limit(limit),
	)
	if onlyCompleted {
		query = append(query, dbmodels.EvaluationWhere.CurrentStatus.EQ(dbmodels.EvaluationsCurrentStatusCOMPLETED))
	}
	if searchTerm != "" {
		query = append(
			query, dbmodels.EvaluationWhere.PatientName.LIKE("%"+searchTerm+"%"),
		)
	}
	if specialist_id != "" {
		query = append(query, dbmodels.EvaluationWhere.SpecialistID.EQ(specialist_id))
	}

	evaluations, err := dbmodels.Evaluations(query...).All(ctx, f.Exec)
	if err != nil {
		return nil, err
	}
	var domainEvaluations []*domain.Evaluation
	for _, evaluation := range evaluations {
		domainEval := dbEvaluationToDomain(evaluation)
		domainEvaluations = append(domainEvaluations, &domainEval)
	}
	return domainEvaluations, nil
}
