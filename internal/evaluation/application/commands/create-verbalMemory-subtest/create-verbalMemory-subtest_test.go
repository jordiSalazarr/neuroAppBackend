package createverbalmemorysubtest

import (
	"context"
	"errors"
	"testing"
	"time"

	// Cambia este import por el de tu handler real:
	// p. ej. "neuro.app.jordi/internal/evaluation/application/createverbalmemorysubtest"

	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
)

/* =========================
   FAKES / MOCKS
========================= */

type fakeEvalRepo struct {
	evals  map[string]domain.Evaluation
	getErr error
	calls  int
}

var _ domain.EvaluationsRepository = (*fakeEvalRepo)(nil)

func (f *fakeEvalRepo) GetMany(ctx context.Context, fromDate, toDate time.Time, offset, limit int, searchTerm string, specialist_id string, onlyCompleted bool) ([]*domain.Evaluation, error) {
	return nil, nil
}
func (f *fakeEvalRepo) Update(ctx context.Context, e domain.Evaluation) error {
	return nil
}

func (f *fakeEvalRepo) Save(ctx context.Context, e domain.Evaluation) error {
	// no usado en estos tests
	return nil
}

func (f *fakeEvalRepo) GetByID(ctx context.Context, id string) (domain.Evaluation, error) {
	f.calls++
	if f.getErr != nil {
		return domain.Evaluation{}, f.getErr
	}
	ev, ok := f.evals[id]
	if !ok {
		return domain.Evaluation{}, errors.New("evaluation not found")
	}
	return ev, nil
}

type fakeLLM struct {
	analysis string
	err      error
	calls    int
	lastAge  int
	lastSub  *VEMdomain.VerbalMemorySubtest
}

var _ domain.LLMService = (*fakeLLM)(nil)

func (f *fakeLLM) LanguageFluencyAnalysis(subtest *LFdomain.LanguageFluency, patientAge int) (string, error) {
	// no usado en estos tests
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) ExecutiveFunctionsAnalysis(subtest *EFdomain.ExecutiveFunctionsSubtest, patientAge int) (string, error) {
	// no usado en estos tests
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) GenerateAnalysis(e domain.Evaluation) (string, error) {
	// no usado en estos tests
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) LettersCancellationAnalysis(_ *LCdomain.LettersCancellationSubtest, _ int) (string, error) {
	// no usado en estos tests
	return "ok", nil
}

func (f *fakeLLM) VerbalMemoryAnalysis(s *VEMdomain.VerbalMemorySubtest, age int) (string, error) {
	f.calls++
	f.lastAge = age
	f.lastSub = s
	if f.err != nil {
		return "", f.err
	}
	if f.analysis == "" {
		return "analysis-ok", nil
	}
	return f.analysis, nil
}

type fakeVerbalRepo struct {
	saveErr error
	calls   int
	last    VEMdomain.VerbalMemorySubtest
}

var _ VEMdomain.VerbalMemoryRepository = (*fakeVerbalRepo)(nil)

func (f *fakeVerbalRepo) GetByID(ctx context.Context, id string) (VEMdomain.VerbalMemorySubtest, error) {
	// no usado en estos tests
	return VEMdomain.VerbalMemorySubtest{}, nil
}
func (f *fakeVerbalRepo) GetByEvaluationID(ctx context.Context, evaluationID string) (VEMdomain.VerbalMemorySubtest, error) {
	// no usado en estos tests
	return VEMdomain.VerbalMemorySubtest{}, nil
}

func (f *fakeVerbalRepo) Save(ctx context.Context, s VEMdomain.VerbalMemorySubtest) error {
	f.calls++
	f.last = s
	if f.saveErr != nil {
		return f.saveErr
	}
	return nil
}

/* =========================
     HELPERS
========================= */

func validCommandImmediate() CreateVerbalMemorySubtestCommand {
	return CreateVerbalMemorySubtestCommand{
		EvaluationID: "eval-1",
		StartAt:      time.Now().UTC(), // ~0s => immediate
		GivenWords:   []string{"pera", "manzana", "platano", "uva", "camisa", "pantalon", "chaqueta", "falda", "perro", "gato", "vaca", "caballo"},
		RecalledWords: []string{
			"manzana", "pera", "platano", "uva", "camisa",
			"pantalon", "chaqueta", "falda", "perro", "gato", "vaca", "caballo",
		},
	}
}

func validCommandDelayed() CreateVerbalMemorySubtestCommand {
	return CreateVerbalMemorySubtestCommand{
		EvaluationID: "eval-1",
		StartAt:      time.Now().UTC().Add(-30 * time.Minute), // claramente diferido
		GivenWords:   []string{"pera", "manzana", "platano", "uva", "camisa", "pantalon", "chaqueta", "falda", "perro", "gato", "vaca", "caballo"},
		RecalledWords: []string{
			"manzana", "pera", "gato", "vaca", "camisa", "falda", "caballo",
		},
	}
}

func setupSuccess(t *testing.T, patientAge int, llmText string) (*fakeEvalRepo, *fakeLLM, *fakeVerbalRepo) {
	t.Helper()
	er := &fakeEvalRepo{
		evals: map[string]domain.Evaluation{
			"eval-1": {PatientAge: patientAge},
		},
	}
	llm := &fakeLLM{analysis: llmText}
	vr := &fakeVerbalRepo{}
	return er, llm, vr
}

/* =========================
       TESTS
========================= */

func TestCreateVerbalMemorySubtest_Success_Immediate(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er, llm, vr := setupSuccess(t, 72, "Análisis HVLT inmediato")

	got, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verificaciones básicas
	if got.EvaluationID != cmd.EvaluationID {
		t.Errorf("EvaluationID mismatch: want %s got %s", cmd.EvaluationID, got.EvaluationID)
	}
	if got.Type != VEMdomain.VerbalMemorySubtypeImmediate {
		t.Errorf("expected Type=immediate, got %s", got.Type)
	}
	if got.CreatedAt.IsZero() {
		t.Errorf("CreatedAt should be set")
	}
	// Score en rango
	if got.Score.Score < 0 || got.Score.Score > 100 {
		t.Errorf("score out of range: %d", got.Score.Score)
	}
	// LLM llamado y con la edad correcta
	if llm.calls != 1 {
		t.Errorf("LLM should be called once, got %d", llm.calls)
	}
	if llm.lastAge != 72 {
		t.Errorf("age not propagated to LLM: want 72 got %d", llm.lastAge)
	}
	// Se guardó el subtest
	if vr.calls != 1 {
		t.Errorf("Save should be called once, got %d", vr.calls)
	}
	// AssistantAnalysis persistido
	if got.AssistanAnalysis == "" {
		t.Errorf("AssistantAnalysis should be set")
	}
}

func TestCreateVerbalMemorySubtest_Success_Delayed(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandDelayed()
	er, llm, vr := setupSuccess(t, 60, "Análisis HVLT diferido")

	got, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got.Type != VEMdomain.VerbalMemorySubtypeDelayed {
		t.Errorf("expected Type=delayed, got %s", got.Type)
	}
	if llm.calls != 1 {
		t.Errorf("LLM should be called once")
	}
	if vr.calls != 1 {
		t.Errorf("Save should be called once")
	}
}

func TestCreateVerbalMemorySubtest_EvaluationNotFound(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er := &fakeEvalRepo{evals: map[string]domain.Evaluation{}} // vacío
	llm := &fakeLLM{analysis: "no importa"}
	vr := &fakeVerbalRepo{}

	_, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err == nil {
		t.Fatalf("expected error when evaluation not found")
	}
	if llm.calls != 0 {
		t.Errorf("LLM should not be called when eval lookup fails")
	}
	if vr.calls != 0 {
		t.Errorf("Save should not be called when eval lookup fails")
	}
}

func TestCreateVerbalMemorySubtest_EvaluationRepoError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er := &fakeEvalRepo{getErr: errors.New("db down")}
	llm := &fakeLLM{}
	vr := &fakeVerbalRepo{}

	_, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err == nil {
		t.Fatalf("expected error from evaluationsRepo")
	}
	if llm.calls != 0 {
		t.Errorf("LLM should not be called")
	}
	if vr.calls != 0 {
		t.Errorf("Save should not be called")
	}
}

func TestCreateVerbalMemorySubtest_LLMError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er, llm, vr := setupSuccess(t, 50, "")
	llm.err = errors.New("llm timeout")

	_, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err == nil {
		t.Fatalf("expected error from LLM")
	}
	if vr.calls != 0 {
		t.Errorf("Save should not be called when LLM fails")
	}
}

func TestCreateVerbalMemorySubtest_SaveError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er, llm, vr := setupSuccess(t, 50, "ok")
	vr.saveErr = errors.New("write failed")

	_, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err == nil {
		t.Fatalf("expected error from Save")
	}
	// LLM fue llamado antes de fallar al guardar
	if llm.calls != 1 {
		t.Errorf("LLM should be called once before save error")
	}
}

func TestCreateVerbalMemorySubtest_ScoringError_WhenGivenWordsEmpty(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	cmd.GivenWords = nil // provoca error en ScoreVerbalMemory si valida lista vacía
	er, llm, vr := setupSuccess(t, 67, "no importa")

	_, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err == nil {
		t.Fatalf("expected error from ScoreVerbalMemory (given_words empty)")
	}
	if llm.calls != 0 {
		t.Errorf("LLM should not be called when scoring fails")
	}
	if vr.calls != 0 {
		t.Errorf("Save should not be called when scoring fails")
	}
}

func TestCreateVerbalMemorySubtest_AnalysisPropagatedAndSaved(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er, llm, vr := setupSuccess(t, 45, "perfil atencional: retención adecuada")
	got, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AssistanAnalysis != "perfil atencional: retención adecuada" {
		t.Errorf("assistant analysis mismatch: %q", got.AssistanAnalysis)
	}
	if vr.last.AssistanAnalysis != got.AssistanAnalysis {
		t.Errorf("saved analysis mismatch with returned entity")
	}
}

func TestCreateVerbalMemorySubtest_LLMReceivesSubtestPointerAndAge(t *testing.T) {
	ctx := context.Background()
	cmd := validCommandImmediate()
	er, llm, vr := setupSuccess(t, 33, "ok")

	got, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if llm.lastSub == nil {
		t.Fatalf("expected LLM to receive subtest pointer")
	}
	// sanity checks sobre el subtest pasado al LLM
	if llm.lastSub.EvaluationID != cmd.EvaluationID {
		t.Errorf("LLM subtest evalID mismatch")
	}
	if llm.lastAge != 33 {
		t.Errorf("LLM age mismatch: want 33 got %d", llm.lastAge)
	}
	// coherencia con lo guardado
	if vr.last.Pk == "" || got.Pk == "" {
		t.Errorf("PK should be non-empty")
	}
}

func TestCreateVerbalMemorySubtest_TypeBoundary_Exactly5Minutes(t *testing.T) {
	ctx := context.Background()
	// Si tu NewVerbalMemorySubtest clasifica <=300s como "immediate"
	start := time.Now().UTC().Add(-300 * time.Second)
	cmd := CreateVerbalMemorySubtestCommand{
		EvaluationID:  "eval-1",
		StartAt:       start,
		GivenWords:    []string{"a", "b", "c"},
		RecalledWords: []string{"a"},
	}
	er, llm, vr := setupSuccess(t, 20, "ok")

	got, err := CreateVerbalMemorySubtestCommandhandler(ctx, cmd, er, llm, vr)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Type != VEMdomain.VerbalMemorySubtypeImmediate {
		t.Errorf("expected immediate at 300s boundary, got %s", got.Type)
	}
}
