package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	createevaluation "neuro.app.jordi/internal/evaluation/application/commands/create-evaluation"
	createexecutivefunctionssubtest "neuro.app.jordi/internal/evaluation/application/commands/create-executiveFunctions-subtest"
	createlanguagefluencysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-languageFluency-subtest"
	createlettercancelationsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-letterCancelation-subtest"
	createverbalmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-verbalMemory-subtest"
	createvisualspatialsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visual-spatial-subtest"
	createvisualmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visualMemory-subtest"
	finishevaluation "neuro.app.jordi/internal/evaluation/application/commands/finish-evaluation"
	canfinishevaluation "neuro.app.jordi/internal/evaluation/application/queries/can-finish-evaluation"
	getevaluation "neuro.app.jordi/internal/evaluation/application/queries/get-evaluation"
	listevaluations "neuro.app.jordi/internal/evaluation/application/queries/get-evaluations"
	"neuro.app.jordi/internal/evaluation/domain"
	reports "neuro.app.jordi/internal/evaluation/domain/services"
)

type EvaluationAPI struct {
	PK                string    `json:"pk"`
	PatientName       string    `json:"patientName"`
	PatientAge        int       `json:"patientAge"`
	SpecialistMail    string    `json:"specialistMail"`
	SpecialistID      string    `json:"specialistId"`
	AssistantAnalysis string    `json:"assistantAnalysis"`
	StorageURL        string    `json:"storage_url"`
	CreatedAt         time.Time `json:"createdAt"`
	CurrentStatus     string    `json:"currentStatus"`
}

func domainToAPIEvaluation(eval domain.Evaluation) EvaluationAPI {
	return EvaluationAPI{
		PK:                eval.PK,
		PatientName:       eval.PatientName,
		PatientAge:        eval.PatientAge,
		SpecialistMail:    eval.SpecialistMail,
		SpecialistID:      eval.SpecialistID,
		AssistantAnalysis: eval.AssistantAnalysis,
		StorageURL:        eval.StorageURL,
		CreatedAt:         eval.CreatedAt,
		CurrentStatus:     string(eval.CurrentStatus),
	}
}

func parseTimeFlexible(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil // "no proporcionado"
	}
	// Intentos: RFC3339 y fecha simple
	layouts := []string{
		time.RFC3339,       // 2025-08-23T11:17:30Z o con offset
		"2006-01-02",       // 2025-08-23 (asumimos 00:00:00 UTC)
		"2006-01-02 15:04", // opcional: 2025-08-23 11:17
		"2006-01-02 15:04:05",
	}
	var t time.Time
	var err error
	for _, l := range layouts {
		t, err = time.Parse(l, s)
		if err == nil {
			// Si venía sin zona, normalizamos a UTC
			if t.Location() == time.Local {
				t = time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
			}
			return t.UTC(), nil
		}
	}
	return time.Time{}, fmt.Errorf("invalid time format: %q (use RFC3339 or YYYY-MM-DD)", s)
}

type listEvaluationsQueryDTO struct {
	SpecialistID string `form:"specialist_id"`
	FromDateStr  string `form:"from_date"` // ej: 2025-06-01 o 2025-06-01T00:00:00Z
	ToDateStr    string `form:"to_date"`   // ej: 2025-08-01 o 2025-08-01T23:59:59Z
	SearchTerm   string `form:"search_term"`
	Offset       int    `form:"offset"` // default 0
	Limit        int    `form:"limit"`  // default/cap abajo
}

func (app *App) ListEvaluations(c *gin.Context) {
	// 1) Bind de query params
	var dto listEvaluationsQueryDTO
	if err := c.ShouldBindQuery(&dto); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query params", "details": err.Error()})
		return
	}

	// 2) Defaults y saneo de paginación
	offset := dto.Offset
	if offset < 0 {
		offset = 0
	}
	limit := dto.Limit
	if limit <= 0 {
		limit = 50
	}
	const maxLimit = 200
	if limit > maxLimit {
		limit = maxLimit
	}

	// 3) Parse de fechas (opcionales)
	from, err := parseTimeFlexible(dto.FromDateStr)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing from_date", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "param": "from_date"})
		return
	}
	to, err := parseTimeFlexible(dto.ToDateStr)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing to_date", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "param": "to_date"})
		return
	}

	query := listevaluations.ListEvaluationsQuery{
		SpecialistID:  dto.SpecialistID,
		FromDate:      from, // zero => sin filtro si tu capa de aplicación lo permite
		ToDate:        to,   // zero => sin filtro
		SearchTerm:    dto.SearchTerm,
		Offset:        offset,
		Limit:         limit,
		OnlyCompleted: true,
	}
	//TODO: this is a query handler...
	evaluations, err := listevaluations.GetEvaluationsCommandHandler(
		c.Request.Context(),
		query,
		app.Repositories.EvaluationsRepository,
	)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error listing evaluations", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	returnEvals := make([]EvaluationAPI, 0, len(evaluations))
	for _, eval := range evaluations {
		apiEval := domainToAPIEvaluation(*eval)
		returnEvals = append(returnEvals, apiEval)
	}

	c.JSON(http.StatusOK, gin.H{
		"evaluations": returnEvals,
		"meta": gin.H{
			"offset": offset,
			"limit":  limit,
			"count":  len(returnEvals),
		},
	})
}

func (app *App) GetEvaluation(c *gin.Context) {
	var query getevaluation.GetEvaluationQuery
	id := c.Params.ByName("id")
	query.EvaluationID = id
	evaluation, err := getevaluation.GetEvaluationQueryHandler(c.Request.Context(), query, app.Repositories.EvaluationsRepository, app.Repositories.VerbalMemorySubtestRepository, app.Repositories.VisualMemorySubtestRepository, app.Repositories.ExecutiveFunctionsSubtestRepository, app.Repositories.LetterCancellationRepository, app.Repositories.LanguageFluencyRepository, app.Repositories.VisualSpatialRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error getting evaluation", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    "Evaluation created",
		"evaluation": (evaluation),
	})
}
func (app *App) CreateEvaluation(c *gin.Context) {
	var command createevaluation.CreateEvaluationCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	evaluation, err := createevaluation.CreateEvaluationCommandHandler(command, c, app.Repositories.EvaluationsRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error creating evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    "Evaluation created",
		"evaluation": evaluation,
	})
}
func (app *App) FinnishEvaluation(c *gin.Context) {
	var command finishevaluation.FinisEvaluationCommannd
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when finishiing evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "internal error"})
		return
	}

	reportsPublisher := reports.Publisher{
		Bucket: domain.NewMockBucket(),
	}

	evaluation, err := finishevaluation.FinisEvaluationCommanndHandler(c.Request.Context(),
		command, app.Repositories.EvaluationsRepository,
		app.Services.LLMService, app.Services.FileFormater, reportsPublisher, app.Repositories.VerbalMemorySubtestRepository,
		app.Repositories.VisualMemorySubtestRepository, app.Repositories.ExecutiveFunctionsSubtestRepository, app.Repositories.LetterCancellationRepository, app.Repositories.LanguageFluencyRepository, app.Repositories.VisualSpatialRepository,
		app.Services.MailService)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when finishiing evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":    "Evaluation finnished",
		"evaluation": evaluation,
	})
}

func (app *App) CreateLetterCancellationSubtest(c *gin.Context) {
	var command createlettercancelationsubtest.CreateLetterCancellationSubtestCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating letter cancellation evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtest, err := createlettercancelationsubtest.CreateLetterCancellationSubtestCommandHandler(c.Request.Context(), command, app.Repositories.LetterCancellationRepository, app.Repositories.EvaluationsRepository, app.Services.LLMService)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error  when creating letter cancellation evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtest": subtest})
}

func (app *App) VerbalMemorySubtest(c *gin.Context) {
	var command createverbalmemorysubtest.CreateVerbalMemorySubtestCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating verbal memory evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtest, err := createverbalmemorysubtest.CreateVerbalMemorySubtestCommandhandler(c.Request.Context(), command, app.Repositories.EvaluationsRepository, app.Services.LLMService, app.Repositories.VerbalMemorySubtestRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating letter cancellation evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtest": subtest})
}

func (app *App) ExecutiveFunctionsSubtest(c *gin.Context) {
	var command createexecutivefunctionssubtest.CreateExecutiveFunctionsSubtestCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating executive function evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtest, err := createexecutivefunctionssubtest.CreateExecutiveFunctionsSubtestCommandHandler(c.Request.Context(), command, app.Repositories.EvaluationsRepository, app.Services.LLMService, app.Repositories.ExecutiveFunctionsSubtestRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error  when creating executive function evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtest": subtest})
}

func (app *App) LanguageFluencySubtest(c *gin.Context) {
	ct := c.ContentType()

	switch ct {
	case "multipart/form-data":
		app.languageFluencyFromMultipart(c)
	default: // JSON u otros -> intentamos JSON como hasta ahora
		var command createlanguagefluencysubtest.CreateLanguageFluencySubtestCommand
		if err := c.ShouldBindJSON(&command); err != nil {
			app.Logger.Error(c.Request.Context(), "error parsing when creating language fluency evaluation (json path)", err, c.Keys)
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		app.execLanguageFluencyCommand(c, command, "json")
	}
}

func (app *App) languageFluencyFromMultipart(c *gin.Context) {
	// 1) Límite de tamaño razonable
	const maxBytes = 20 << 20 // 20 MiB
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)

	// 2) Leer archivo de audio
	fileHeader, err := c.FormFile("audio")
	if err != nil {
		app.Logger.Error(c.Request.Context(), "missing audio file", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'audio' file"})
		return
	}
	if fileHeader.Size == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty 'audio' file"})
		return
	}
	f, err := fileHeader.Open()
	if err != nil {
		app.Logger.Error(c.Request.Context(), "cannot open audio", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open audio file"})
		return
	}
	defer f.Close()

	audioBytes, err := io.ReadAll(f)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "cannot read audio", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot read audio file"})
		return
	}

	// 3) Leer payload JSON
	var cmd createlanguagefluencysubtest.CreateLanguageFluencySubtestCommand
	jsonStr := c.PostForm("payload")
	if jsonStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing 'payload' JSON"})
		return
	}
	if err := json.Unmarshal([]byte(jsonStr), &cmd); err != nil {
		app.Logger.Error(c.Request.Context(), "invalid payload JSON", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid 'payload' JSON"})
		return
	}

	// 4) STT
	if app.Services.SpeechToText == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "speech-to-text service not configured"})
		return
	}
	transcript, err := app.Services.SpeechToText.GetTextFromSpeech(audioBytes)

	// limpiar buffer (no persistimos)
	for i := range audioBytes {
		audioBytes[i] = 0
	}
	if err != nil {
		app.Logger.Error(c.Request.Context(), "speech-to-text error", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to transcribe audio"})
		return
	}

	// 5) Extraer palabras
	cmd.Words = extractWords(transcript)
	// Defaults si no llegan (tu dominio los exige no vacíos)
	if cmd.Duration == 0 {
		cmd.Duration = 60
	}
	cmd.Language = "es"

	cmd.Proficiency = "nativo"

	cmd.Category = "animales"

	app.execLanguageFluencyCommand(c, cmd, "audio+json")
}

func (app *App) execLanguageFluencyCommand(
	c *gin.Context,
	command createlanguagefluencysubtest.CreateLanguageFluencySubtestCommand,
	inputSource string,
) {
	if strings.TrimSpace(command.EvaluationID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "evaluationId is required"})
		return
	}
	// Si quisieras forzar que haya palabras en el modo JSON:
	// if len(command.Words) == 0 && inputSource == "json" {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "words are required in JSON path"})
	// 	return
	// }

	subtest, err := createlanguagefluencysubtest.CreateLanguageFluencySubtestCommandHandler(
		c.Request.Context(),
		command,
		app.Repositories.EvaluationsRepository,
		app.Services.LLMService,
		app.Repositories.LanguageFluencyRepository,
	)
	if err != nil {
		// Envuelve errores de dominio comunes para devolver 400 en vez de 500 si aplica
		if errors.Is(err, context.DeadlineExceeded) {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "timeout"})
			return
		}
		app.Logger.Error(c.Request.Context(), "error when creating language fluency evaluation ("+inputSource+")", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"subtest":     subtest,
		"inputSource": inputSource,
	})
}

// --- util de tokenización sencilla orientada a ES ---
// - minúsculas
// - elimina todo lo no letra/apóstrofo
// - separa por espacios colapsados
func extractWords(s string) []string {
	if s == "" {
		return nil
	}
	s = strings.ToLower(s)
	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if unicode.IsLetter(r) || r == '\'' {
			b.WriteRune(r)
		} else {
			b.WriteByte(' ')
		}
	}
	out := strings.Fields(b.String())
	if len(out) == 0 {
		return nil
	}
	return out
}

func (app *App) CreateVisualMemorySubtest(c *gin.Context) {
	var cmd createvisualmemorysubtest.CreateVisualMemorySubtestCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := createvisualmemorysubtest.CreateVisualMemoryCommandHandler(c.Request.Context(), cmd, app.Repositories.VisualMemorySubtestRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (app *App) CreateVisualSpatialSubtest(c *gin.Context) {
	var cmd createvisualspatialsubtest.CreateVisualSpatialSubtestCommand
	if err := c.ShouldBindJSON(&cmd); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating visual spatial evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	sub, err := createvisualspatialsubtest.CreateViusualSpatialCommandHandler(c.Request.Context(), cmd, app.Repositories.VisualSpatialRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual spatial evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, sub)
}

func (app *App) CanFinishEvaluation(c *gin.Context) {
	evalID := c.Param("evaluation_id")
	if evalID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "evaluation_id is required"})
		return
	}
	specialistID := c.Param("specialist_id")
	if specialistID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "specialist_id is required"})
		return
	}

	query := canfinishevaluation.CanFinishEvaluationQuery{
		EvaluationID: evalID,
		SpecialistID: specialistID,
	}
	canFinish, err := canfinishevaluation.CanFinishEvaluationQueryHandler(c.Request.Context(), query, app.Repositories.EvaluationsRepository, app.Repositories.VerbalMemorySubtestRepository, app.Repositories.VisualMemorySubtestRepository, app.Repositories.ExecutiveFunctionsSubtestRepository, app.Repositories.LetterCancellationRepository, app.Repositories.LanguageFluencyRepository, app.Repositories.VisualSpatialRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error checking if can finish evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"can_finish": canFinish})
}
