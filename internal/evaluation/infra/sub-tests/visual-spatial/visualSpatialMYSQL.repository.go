package VPdomain

import (
	"context"
	"database/sql"
	"errors"
	"time"

	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

// =========================
// Repo
// =========================

type VisualSpatialMYSQLRepo struct {
	DB *sql.DB
}

func NewVisualSpatialMYSQLRepo(db *sql.DB) *VisualSpatialMYSQLRepo {
	return &VisualSpatialMYSQLRepo{DB: db}
}

// Save hace un upsert "manual":
// - Intenta UPDATE por id.
// - Si no actualiza filas, hace INSERT.
// Nota: actualiza UpdatedAt a time.Now().
func (r *VisualSpatialMYSQLRepo) Save(ctx context.Context, res *VPdomain.VisualSpatialSubtest) error {
	if r == nil || r.DB == nil {
		return errors.New("nil repo or DB")
	}
	if res == nil {
		return errors.New("nil VPdomain.VisualSpatialSubtest")
	}

	// Aseguramos precisión a milisegundos (DATETIME(3))
	now := time.Now().Truncate(time.Millisecond)
	if res.CreatedAt.IsZero() {
		res.CreatedAt = now
	}
	res.UpdatedAt = now

	row := toRow(res)

	// UPDATE primero (optimista)
	const updateSQL = `
		UPDATE visual_spatial_subtest
		   SET evaluation_id = ?,
		       score         = ?,
		       note          = ?,
		       updated_at    = ?
		 WHERE id = ?
	`
	ur, err := r.DB.ExecContext(ctx, updateSQL,
		row.EvaluationID, row.Score, row.Note, row.UpdatedAt, row.ID,
	)
	if err != nil {
		return err
	}
	affected, err := ur.RowsAffected()
	if err != nil {
		return err
	}
	if affected > 0 {
		return nil
	}

	// Si no existe, INSERT
	const insertSQL = `
		INSERT INTO visual_spatial_subtest
		    (id, evaluation_id, score, note, created_at, updated_at)
		VALUES (?,  ?,            ?,     ?,    ?,          ?)
	`
	_, err = r.DB.ExecContext(ctx, insertSQL,
		row.ID, row.EvaluationID, row.Score, row.Note, row.CreatedAt, row.UpdatedAt,
	)
	return err
}

func (r *VisualSpatialMYSQLRepo) GetByID(ctx context.Context, id string) (*VPdomain.VisualSpatialSubtest, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("nil repo or DB")
	}
	const q = `
		SELECT id, evaluation_id, score, note, created_at, updated_at
		  FROM visual_spatial_subtest
		 WHERE id = ?
		 LIMIT 1
	`
	var row visualSpatialRow
	err := r.DB.QueryRowContext(ctx, q, id).Scan(
		&row.ID, &row.EvaluationID, &row.Score, &row.Note, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err // sql.ErrNoRows si no existe
	}
	return row.toDomain()
}

func (r *VisualSpatialMYSQLRepo) GetByEvaluationID(ctx context.Context, evaluationID string) (*VPdomain.VisualSpatialSubtest, error) {
	if r == nil || r.DB == nil {
		return nil, errors.New("nil repo or DB")
	}
	// Si hubiera más de un registro por evaluation_id, devolvemos el más reciente.
	const q = `
		SELECT id, evaluation_id, score, note, created_at, updated_at
		  FROM visual_spatial_subtest
		 WHERE evaluation_id = ?
		 ORDER BY created_at DESC
		 LIMIT 1
	`
	var row visualSpatialRow
	err := r.DB.QueryRowContext(ctx, q, evaluationID).Scan(
		&row.ID, &row.EvaluationID, &row.Score, &row.Note, &row.CreatedAt, &row.UpdatedAt,
	)
	if err != nil {
		return nil, err // sql.ErrNoRows si no existe
	}
	return row.toDomain()
}

// =========================
// Mappers dominio ⇄ DB
// =========================

type visualSpatialRow struct {
	ID           string
	EvaluationID string
	Score        int
	Note         string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func toRow(d *VPdomain.VisualSpatialSubtest) visualSpatialRow {
	// Mapea el typo del dominio: EvalautionId -> evaluation_id
	return visualSpatialRow{
		ID:           d.Id,
		EvaluationID: d.EvalautionId,
		Score:        d.Score.Val,
		Note:         d.Note.Val,
		CreatedAt:    d.CreatedAt.Truncate(time.Millisecond),
		UpdatedAt:    d.UpdatedAt.Truncate(time.Millisecond),
	}
}

func (r visualSpatialRow) toDomain() (*VPdomain.VisualSpatialSubtest, error) {
	return VPdomain.NewVisualSpatialSubtestFromExisting(
		r.ID,
		r.EvaluationID,
		r.Note,
		r.Score,
		r.CreatedAt,
		r.UpdatedAt,
	)
}
