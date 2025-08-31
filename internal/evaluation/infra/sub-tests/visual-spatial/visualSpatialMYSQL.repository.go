package INFRAvisualspatial

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/aarondl/sqlboiler/v4/boil"
	"github.com/aarondl/sqlboiler/v4/types"
	"neuro.app.jordi/database/dbmodels"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

type ClockResultMySQLRepo struct {
	DB *sql.DB
}

func NewClockResultMySQLRepo(db *sql.DB) *ClockResultMySQLRepo {
	return &ClockResultMySQLRepo{DB: db}
}

// ---------------------------
// Mapeos dominio ⇄ BD
// ---------------------------

func toDB(res *VPdomain.ClockDrawResult) (*dbmodels.ClockDrawSubtestResult, error) {
	if res == nil {
		return nil, errors.New("nil domain result")
	}
	// reasons: []string -> JSON
	var reasons types.JSON
	if len(res.Reasons) > 0 {
		b, err := json.Marshal(res.Reasons)
		if err != nil {
			return nil, err
		}
		reasons = types.JSON(b)
	} else {
		reasons = types.JSON([]byte("[]"))
	}

	return &dbmodels.ClockDrawSubtestResult{
		ID:                    res.ID,
		EvaluationID:          res.EvaluationID,
		Pass:                  res.Pass,
		Reasons:               reasons,
		CenterX:               int(res.CenterX),
		CenterY:               int(res.CenterY),
		Radius:                res.Radius,
		DialCircularity:       res.DialCircularity,
		MinuteAngleDeg:        res.MinuteAngleDeg,
		HourAngleDeg:          res.HourAngleDeg,
		ExpectedMinuteAngle:   res.ExpectedMinuteAngle,
		ExpectedHourAngle:     res.ExpectedHourAngle,
		MinuteAngularErrorDeg: res.MinuteAngularErrorDeg,
		HourAngularErrorDeg:   res.HourAngularErrorDeg,
		CreatedAt:             res.CreatedAt,
		UpdatedAt:             res.UpdatedAt,
	}, nil
}

func toDomain(m *dbmodels.ClockDrawSubtestResult) (*VPdomain.ClockDrawResult, error) {
	if m == nil {
		return nil, errors.New("nil db model")
	}
	var reasons []string
	if len(m.Reasons) > 0 {
		if err := json.Unmarshal([]byte(m.Reasons), &reasons); err != nil {
			// si falla, no rompemos: devolvemos vacío
			reasons = nil
		}
	}

	return &VPdomain.ClockDrawResult{
		ID:                    m.ID,
		EvaluationID:          m.EvaluationID,
		Pass:                  m.Pass,
		Reasons:               reasons,
		CenterX:               int(m.CenterX),
		CenterY:               int(m.CenterY),
		Radius:                m.Radius,
		DialCircularity:       m.DialCircularity,
		MinuteAngleDeg:        m.MinuteAngleDeg,
		HourAngleDeg:          m.HourAngleDeg,
		ExpectedMinuteAngle:   m.ExpectedMinuteAngle,
		ExpectedHourAngle:     m.ExpectedHourAngle,
		MinuteAngularErrorDeg: m.MinuteAngularErrorDeg,
		HourAngularErrorDeg:   m.HourAngularErrorDeg,
		CreatedAt:             m.CreatedAt,
		UpdatedAt:             m.UpdatedAt,
	}, nil
}

// ---------------------------
// Persistencia
// ---------------------------

// Save: inserta (o upsert si prefieres idempotencia por ID).
func (r *ClockResultMySQLRepo) Save(ctx context.Context, res *VPdomain.ClockDrawResult) error {
	if res == nil {
		return errors.New("res nil")
	}

	// fallback de timestamps si vienen vacíos
	now := time.Now().UTC()
	if res.CreatedAt.IsZero() {
		res.CreatedAt = now
	}
	if res.UpdatedAt.IsZero() {
		res.UpdatedAt = now
	}

	dbObj, err := toDB(res)
	if err != nil {
		return err
	}

	return dbObj.Insert(ctx, r.DB, boil.Infer())

}

// (opcional) método para ping/healthcheck
func (r *ClockResultMySQLRepo) Ping(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return r.DB.PingContext(ctx)
}

// ---------------------------
// Lectura (ejemplos)
// ---------------------------

func (r *ClockResultMySQLRepo) GetByID(ctx context.Context, id string) (*VPdomain.ClockDrawResult, error) {
	m, err := dbmodels.FindClockDrawSubtestResult(ctx, r.DB, id)
	if err != nil {
		return nil, err
	}
	return toDomain(m)
}

func (r *ClockResultMySQLRepo) GetByEvaluationID(ctx context.Context, id string) (*VPdomain.ClockDrawResult, error) {
	m, err := dbmodels.ClockDrawSubtestResults(dbmodels.ClockDrawSubtestResultWhere.ID.LIKE(id)).One(ctx, r.DB)
	if err != nil {
		return nil, err
	}
	return toDomain(m)
}
