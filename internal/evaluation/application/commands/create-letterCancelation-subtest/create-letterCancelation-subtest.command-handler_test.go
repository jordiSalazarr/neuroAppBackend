package createlettercancelationsubtest

// handler_letters_cancellation_test.go

import (
	"context"
	"errors"
	"testing"
	"time"

	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	// Ajusta estos imports a tu repo real:
)

// ---------- Fakes (simples, sin frameworks) ----------

type fakeLetterCancellationRepo struct {
	saved   []*LCdomain.LettersCancellationSubtest
	saveErr error
}

func (f *fakeLetterCancellationRepo) GetByEvaluationID(ctx context.Context, evaluationID string) (LCdomain.LettersCancellationSubtest, error) {

	return LCdomain.LettersCancellationSubtest{}, nil
}

func (f *fakeLetterCancellationRepo) Save(ctx context.Context, s *LCdomain.LettersCancellationSubtest) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.saved = append(f.saved, s)
	return nil
}

type fakeEvalRepo struct {
	evals  map[string]domain.Evaluation
	getErr error
}

// Si quieres, también puedes añadir el assert de interfaz:
// var _ domain.EvaluationsRepository = (*fakeEvalRepo)(nil)
func (f *fakeEvalRepo) GetMany(ctx context.Context, fromDate, toDate time.Time, offset, limit int, searchTerm string, specialist_id string, onlyCompleted bool) ([]*domain.Evaluation, error) {
	return nil, nil
}
func (f *fakeEvalRepo) GetByID(ctx context.Context, id string) (domain.Evaluation, error) {
	if f.getErr != nil {
		return domain.Evaluation{}, f.getErr
	}
	e, ok := f.evals[id]
	if !ok {
		return domain.Evaluation{}, errors.New("evaluation not found")
	}
	return e, nil
}

func (f *fakeEvalRepo) Save(ctx context.Context, e domain.Evaluation) error {
	// No lo necesitas en estos tests; implementa si quieres
	return nil
}

func (f *fakeEvalRepo) Update(ctx context.Context, e domain.Evaluation) error {
	// No lo necesitas en estos tests; implementa si quieres
	return nil
}

type fakeLLM struct {
	analysis string
	err      error
	calls    int
	lastAge  int
}

// Asegura en compile-time que implementa la interfaz:
var _ domain.LLMService = (*fakeLLM)(nil)

func (f *fakeLLM) LanguageFluencyAnalysis(subtest *LFdomain.LanguageFluency, patientAge int) (string, error) {
	// No lo usas en estos tests; devuelve algo neutro o propaga f.err si quieres simular fallo global
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) ExecutiveFunctionsAnalysis(subtest *EFdomain.ExecutiveFunctionsSubtest, patientAge int) (string, error) {
	// No lo usas en estos tests; devuelve algo neutro o propaga f.err si quieres simular fallo global
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) VerbalMemoryAnalysis(subtest *VEMdomain.VerbalMemorySubtest, age int) (string, error) {
	// No lo usas en estos tests; devuelve algo neutro o propaga f.err si quieres simular fallo global
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) GenerateAnalysis(evaluation domain.Evaluation) (string, error) {
	// No lo usas en estos tests; devuelve algo neutro o propaga f.err si quieres simular fallo global
	if f.err != nil {
		return "", f.err
	}
	return "ok", nil
}

func (f *fakeLLM) LettersCancellationAnalysis(s *LCdomain.LettersCancellationSubtest, age int) (string, error) {
	f.calls++
	f.lastAge = age
	if f.err != nil {
		return "", f.err
	}
	if f.analysis == "" {
		return "analysis", nil
	}
	return f.analysis, nil
}

// ---------- Helpers ----------

func validCommand() CreateLetterCancellationSubtestCommand {
	return CreateLetterCancellationSubtestCommand{
		TotalTargets: 80,
		Correct:      62,
		Errors:       10,
		TimeInSecs:   150, // 2.5 min
		EvaluationID: "eval-123",
	}
}

func setupSuccessDeps(t *testing.T, patientAge int, llmText string) (*fakeLetterCancellationRepo, *fakeEvalRepo, *fakeLLM) {
	t.Helper()
	lr := &fakeLetterCancellationRepo{}
	er := &fakeEvalRepo{
		evals: map[string]domain.Evaluation{
			"eval-123": { // asume que domain.Evaluation tiene al menos este campo:
				PatientAge: patientAge,
			},
		},
	}
	llm := &fakeLLM{analysis: llmText}
	return lr, er, llm
}

// ---------- Tests ----------

func TestCreateLettersCancellationSubtest_Success(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr, er, llm := setupSuccessDeps(t, 72, "Análisis clínico breve")

	subtest, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if subtest == nil {
		t.Fatalf("expected subtest, got nil")
	}
	// Asersiones básicas de propagación de datos
	if subtest.TotalTargets != cmd.TotalTargets || subtest.Correct != cmd.Correct || subtest.Errors != cmd.Errors {
		t.Errorf("fields not propagated correctly: got %+v", subtest)
	}
	if subtest.EvaluationID != cmd.EvaluationID {
		t.Errorf("expected EvaluationID %q, got %q", cmd.EvaluationID, subtest.EvaluationID)
	}
	// AssistantAnalysis del LLM
	if subtest.AssistantAnalysis != "Análisis clínico breve" {
		t.Errorf("expected AssistantAnalysis to be set")
	}
	// Guardado en repo
	if len(lr.saved) != 1 {
		t.Errorf("expected one saved subtest, got %d", len(lr.saved))
	}
	// PK y CreatedAt razonables
	if subtest.PK == "" {
		t.Errorf("expected PK to be non-empty")
	}
	if subtest.CreatedAt.IsZero() {
		t.Errorf("expected CreatedAt to be set")
	}
	// LLM llamado una vez y con la edad correcta
	if llm.calls != 1 {
		t.Errorf("expected LLM to be called once, got %d", llm.calls)
	}
	if llm.lastAge != 72 {
		t.Errorf("expected patient age 72 passed to LLM, got %d", llm.lastAge)
	}
}

func TestCreateLettersCancellationSubtest_InvalidInputs(t *testing.T) {
	ctx := context.Background()
	lr, er, llm := setupSuccessDeps(t, 68, "ok")

	// 1) totalTargets <= 0
	_, err := CreateLetterCancellationSubtestCommandHandler(ctx, CreateLetterCancellationSubtestCommand{
		TotalTargets: 0, Correct: 10, Errors: 2, TimeInSecs: 120, EvaluationID: "eval-123",
	}, lr, er, llm)
	if err == nil {
		t.Errorf("expected error for TotalTargets=0")
	}

	// 2) correct > totalTargets
	_, err = CreateLetterCancellationSubtestCommandHandler(ctx, CreateLetterCancellationSubtestCommand{
		TotalTargets: 50, Correct: 60, Errors: 2, TimeInSecs: 120, EvaluationID: "eval-123",
	}, lr, er, llm)
	if err == nil {
		t.Errorf("expected error for Correct > TotalTargets")
	}

	// 3) errors < 0 (no puede suceder en JSON normal, pero validamos por si acaso)
	_, err = CreateLetterCancellationSubtestCommandHandler(ctx, CreateLetterCancellationSubtestCommand{
		TotalTargets: 50, Correct: 10, Errors: -1, TimeInSecs: 120, EvaluationID: "eval-123",
	}, lr, er, llm)
	if err == nil {
		t.Errorf("expected error for Errors < 0")
	}

	// 4) timeInSecs <= 0
	_, err = CreateLetterCancellationSubtestCommandHandler(ctx, CreateLetterCancellationSubtestCommand{
		TotalTargets: 50, Correct: 10, Errors: 1, TimeInSecs: 0, EvaluationID: "eval-123",
	}, lr, er, llm)
	if err == nil {
		t.Errorf("expected error for TimeInSecs <= 0")
	}
}

func TestCreateLettersCancellationSubtest_EvaluationNotFound(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr := &fakeLetterCancellationRepo{}
	er := &fakeEvalRepo{evals: map[string]domain.Evaluation{}} // vacío
	llm := &fakeLLM{analysis: "no importa"}

	_, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err == nil {
		t.Fatalf("expected error when evaluation is not found")
	}
	if len(lr.saved) != 0 {
		t.Errorf("should not save on evaluation lookup error")
	}
	if llm.calls != 0 {
		t.Errorf("LLM should not be called if evaluation lookup fails first")
	}
}

func TestCreateLettersCancellationSubtest_EvaluationRepoError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr := &fakeLetterCancellationRepo{}
	er := &fakeEvalRepo{getErr: errors.New("db down")}
	llm := &fakeLLM{}

	_, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err == nil {
		t.Fatalf("expected error from evaluationsRepo")
	}
	if len(lr.saved) != 0 {
		t.Errorf("should not save on evaluation repo error")
	}
	if llm.calls != 0 {
		t.Errorf("LLM should not be called when eval repo fails")
	}
}

func TestCreateLettersCancellationSubtest_LLMError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr, er, llm := setupSuccessDeps(t, 70, "")
	llm.err = errors.New("llm timeout")

	_, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err == nil {
		t.Fatalf("expected error from LLM service")
	}
	if len(lr.saved) != 0 {
		t.Errorf("should not save when LLM fails")
	}
}

func TestCreateLettersCancellationSubtest_SaveError(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr, er, llm := setupSuccessDeps(t, 70, "ok")
	lr.saveErr = errors.New("write failed")

	_, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err == nil {
		t.Fatalf("expected error from repository Save")
	}
	// LLM se llamó antes del Save
	if llm.calls != 1 {
		t.Errorf("expected LLM to be called once before save, got %d", llm.calls)
	}
}

func TestCreateLettersCancellationSubtest_AssistantAnalysisIsPersisted(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr, er, llm := setupSuccessDeps(t, 65, "Perfil atencional con enlentecimiento moderado.")
	subtest, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if subtest.AssistantAnalysis == "" {
		t.Errorf("expected AssistantAnalysis to be set")
	}
	if len(lr.saved) != 1 {
		t.Fatalf("expected one saved record")
	}
	if lr.saved[0].AssistantAnalysis != subtest.AssistantAnalysis {
		t.Errorf("saved analysis mismatch")
	}
}

func TestCreateLettersCancellationSubtest_TimestampsAndIDs(t *testing.T) {
	ctx := context.Background()
	cmd := validCommand()
	lr, er, llm := setupSuccessDeps(t, 60, "ok")

	start := time.Now().Add(-1 * time.Minute)
	subtest, err := CreateLetterCancellationSubtestCommandHandler(ctx, cmd, lr, er, llm)
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if subtest.PK == "" {
		t.Errorf("expected PK not empty")
	}
	if subtest.CreatedAt.Before(start) {
		t.Errorf("CreatedAt looks too old; expected near 'now'")
	}
}
