package VIMinfra

import (
	"context"

	"neuro.app.jordi/database/dbmodels"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// ---------- MAPPERS ----------
func transformDomain(d VIMdomain.VisualMemorySubtest) *dbmodels.VisualMemorySubtest {
	return &dbmodels.VisualMemorySubtest{
		ID:           d.PK,
		EvaluationID: d.EvaluationID,
		Note:         d.Note.Val,
		Score:        d.Score.Val,
		ImageSRC:     null.StringFrom(""),
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}
func transformDB(db *dbmodels.VisualMemorySubtest) *VIMdomain.VisualMemorySubtest {
	d, _ := VIMdomain.NewVisualMemorySubtestFromDB(db.ID, db.EvaluationID, &db.ImageSRC.String, int(db.Score), db.Note, db.CreatedAt, db.UpdatedAt)
	return &d
}

// ---------- REPOSITORY ----------
type VisualMemoryMYSQLRepository struct {
	exec boil.ContextExecutor
}

func NewVisualMemoryMYSQLRepository(exec boil.ContextExecutor) *VisualMemoryMYSQLRepository {
	return &VisualMemoryMYSQLRepository{exec: exec}
}

func (r *VisualMemoryMYSQLRepository) Save(ctx context.Context, d *VIMdomain.VisualMemorySubtest) error {
	dbmodel := transformDomain(*d)
	return dbmodel.Insert(ctx, r.exec, boil.Infer())
}

func (r *VisualMemoryMYSQLRepository) GetLastByEvaluationID(ctx context.Context, evaluationID string) (VIMdomain.VisualMemorySubtest, error) {
	m, err := dbmodels.VisualMemorySubtests(
		dbmodels.VisualMemorySubtestWhere.EvaluationID.EQ(evaluationID),
	).One(ctx, r.exec)
	if err != nil {
		return VIMdomain.VisualMemorySubtest{}, err
	}
	return *transformDB(m), nil
}

func (r *VisualMemoryMYSQLRepository) ListByEvaluationID(ctx context.Context, evaluationID string) ([]VIMdomain.VisualMemorySubtest, error) {
	rows, err := dbmodels.VisualMemorySubtests(
		dbmodels.VisualMemorySubtestWhere.EvaluationID.EQ(evaluationID),
		qm.OrderBy(dbmodels.VisualMemorySubtestColumns.CreatedAt+" DESC"),
	).All(ctx, r.exec)
	if err != nil {
		return nil, err
	}
	out := make([]VIMdomain.VisualMemorySubtest, 0, len(rows))
	for _, m := range rows {
		d := transformDB(m)
		out = append(out, *d)
	}
	return out, nil
}
