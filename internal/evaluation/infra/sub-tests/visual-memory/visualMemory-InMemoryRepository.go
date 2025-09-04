// file: internal/evaluation/infra/sub-tests/visual-memory/repository_sqlboiler.go
package VIMinfra

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"neuro.app.jordi/database/dbmodels"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"

	"github.com/aarondl/null/v8"
	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/queries/qm"
)

// ---------- MAPPERS ----------

func vmDBToDomain(m *dbmodels.VisualMemorySubtest) (VIMdomain.GeoFigureSubtest, error) {
	var score *VIMdomain.GeoShapeScore

	// reasons (TEXT JSON no nulo en la tabla)
	var reasons []string
	if m.Reasons != "" {
		_ = json.Unmarshal([]byte(m.Reasons), &reasons)
	}

	// Construimos score desde columnas "planas" de la tabla
	score = &VIMdomain.GeoShapeScore{
		FinalScore: int(m.Score),
		Pass:       m.Pass,
		Reasons:    reasons,
	}
	if m.Iou.Valid {
		v := m.Iou.Float64
		score.IoU = &v
	}
	if m.Circularity.Valid {
		v := m.Circularity.Float64
		score.Circularity = &v
	}
	if m.AngleRmse.Valid {
		v := m.AngleRmse.Float64
		score.AngleRMSE = &v
	}
	if m.SideCV.Valid {
		v := m.SideCV.Float64
		score.SideCV = &v
	}

	return VIMdomain.GeoFigureSubtest{
		PK:           m.ID,
		EvaluationID: m.EvaluationID,
		Shape:        VIMdomain.ShapeName(m.Shape),
		// La entidad de dominio tiene metadatos de imagen que NO están en esta tabla (results).
		// Si más adelante los persistes en otra tabla, mápelos allí.
		ImageRef:    "",
		ContentType: "",
		Width:       0,
		Height:      0,
		ImageSHA256: "",
		CapturedAt:  time.Time{},
		Status:      VIMdomain.ShapeStatusScored, // esta tabla representa resultados (ya puntuados)
		Score:       score,
		CreatedAt:   m.CreatedAt,
	}, nil
}

func vmDomainToDB(d *VIMdomain.GeoFigureSubtest, dst *dbmodels.VisualMemorySubtest) error {
	dst.ID = d.PK
	dst.EvaluationID = d.EvaluationID
	dst.Shape = string(d.Shape)

	// Valores por defecto seguros (tabla exige NOT NULL)
	pass := false
	final := float64(0)
	reasonsJSON := "[]"

	if d.Score != nil {
		pass = d.Score.Pass
		final = float64(d.Score.FinalScore)
		if b, err := json.Marshal(d.Score.Reasons); err == nil {
			reasonsJSON = string(b)
		}
	}
	dst.Pass = pass
	dst.Score = final
	dst.Reasons = reasonsJSON

	// métricas opcionales
	dst.Iou = null.Float64{}
	dst.Circularity = null.Float64{}
	dst.AngleRmse = null.Float64{}
	dst.SideCV = null.Float64{}

	if d.Score != nil {
		if d.Score.IoU != nil {
			dst.Iou = null.Float64From(*d.Score.IoU)
		}
		if d.Score.Circularity != nil {
			dst.Circularity = null.Float64From(*d.Score.Circularity)
		}
		if d.Score.AngleRMSE != nil {
			dst.AngleRmse = null.Float64From(*d.Score.AngleRMSE)
		}
		if d.Score.SideCV != nil {
			dst.SideCV = null.Float64From(*d.Score.SideCV)
		}
	}

	// center_x/center_y/bbox_w/bbox_h: no están en la entidad -> NULL
	dst.CenterX = null.Int{}
	dst.CenterY = null.Int{}
	dst.BboxW = null.Int{}
	dst.BboxH = null.Int{}

	// created_at: si no viene del dominio (p.ej. zero), lo pone la DB
	if d.CreatedAt.IsZero() {
		dst.CreatedAt = time.Now().UTC()
	} else {
		dst.CreatedAt = d.CreatedAt
	}
	// updated_at lo gestiona la DB con ON UPDATE
	return nil
}

// ---------- REPOSITORY ----------

type SQLBoilerVisualMemoryRepository struct {
	exec boil.ContextExecutor // *sql.DB o *sql.Tx
}

func NewSQLBoilerVisualMemoryRepository(exec boil.ContextExecutor) *SQLBoilerVisualMemoryRepository {
	return &SQLBoilerVisualMemoryRepository{exec: exec}
}

// Save: upsert lógico por ID (PK del dominio).
// Esta tabla es de "resultados", así que se espera d.Score != nil al guardar.
// Si te llega "uploaded" (sin score), considera tener una tabla de ingesta aparte.
func (r *SQLBoilerVisualMemoryRepository) Save(ctx context.Context, d *VIMdomain.GeoFigureSubtest) error {
	if d == nil {
		return errors.New("nil GeoFigureSubtest")
	}
	if strings.TrimSpace(d.PK) == "" {
		return errors.New("empty PK")
	}
	// Encuentra por PK
	existing, err := dbmodels.FindVisualMemorySubtest(ctx, r.exec, d.PK)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	if existing == nil {
		// INSERT
		m := &dbmodels.VisualMemorySubtest{}
		if err := vmDomainToDB(d, m); err != nil {
			return err
		}
		return m.Insert(ctx, r.exec, boil.Infer())
	}

	// UPDATE
	if err := vmDomainToDB(d, existing); err != nil {
		return err
	}
	cols := boil.Whitelist(
		dbmodels.VisualMemorySubtestColumns.EvaluationID,
		dbmodels.VisualMemorySubtestColumns.Shape,
		dbmodels.VisualMemorySubtestColumns.Pass,
		dbmodels.VisualMemorySubtestColumns.Score,
		dbmodels.VisualMemorySubtestColumns.Reasons,
		dbmodels.VisualMemorySubtestColumns.Iou,
		dbmodels.VisualMemorySubtestColumns.Circularity,
		dbmodels.VisualMemorySubtestColumns.AngleRmse,
		dbmodels.VisualMemorySubtestColumns.SideCV,
		dbmodels.VisualMemorySubtestColumns.CenterX,
		dbmodels.VisualMemorySubtestColumns.CenterY,
		dbmodels.VisualMemorySubtestColumns.BboxW,
		dbmodels.VisualMemorySubtestColumns.BboxH,
		// updated_at lo gestiona la DB
	)
	_, err = existing.Update(ctx, r.exec, cols)
	return err
}

// GetByID (PK)
func (r *SQLBoilerVisualMemoryRepository) GetByID(ctx context.Context, id string) (VIMdomain.GeoFigureSubtest, error) {
	m, err := dbmodels.FindVisualMemorySubtest(ctx, r.exec, id)
	if err != nil {
		return VIMdomain.GeoFigureSubtest{}, err
	}
	return vmDBToDomain(m)
}

// GetLastByEvaluationID: último resultado por created_at DESC
func (r *SQLBoilerVisualMemoryRepository) GetLastByEvaluationID(ctx context.Context, evaluationID string) (VIMdomain.GeoFigureSubtest, error) {
	m, err := dbmodels.VisualMemorySubtests(
		dbmodels.VisualMemorySubtestWhere.EvaluationID.EQ(evaluationID),
	).One(ctx, r.exec)
	if err != nil {
		return VIMdomain.GeoFigureSubtest{}, err
	}
	return vmDBToDomain(m)
}

// ListByEvaluationID: todos los resultados de una evaluación
func (r *SQLBoilerVisualMemoryRepository) ListByEvaluationID(ctx context.Context, evaluationID string) ([]VIMdomain.GeoFigureSubtest, error) {
	rows, err := dbmodels.VisualMemorySubtests(
		dbmodels.VisualMemorySubtestWhere.EvaluationID.EQ(evaluationID),
		qm.OrderBy(dbmodels.VisualMemorySubtestColumns.CreatedAt+" DESC"),
	).All(ctx, r.exec)
	if err != nil {
		return nil, err
	}
	out := make([]VIMdomain.GeoFigureSubtest, 0, len(rows))
	for _, m := range rows {
		d, e := vmDBToDomain(m)
		if e != nil {
			return nil, e
		}
		out = append(out, d)
	}
	return out, nil
}
