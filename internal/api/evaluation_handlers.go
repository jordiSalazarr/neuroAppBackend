package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	createevaluation "neuro.app.jordi/internal/evaluation/application/commands/create-evaluation"
	createexecutivefunctionssubtest "neuro.app.jordi/internal/evaluation/application/commands/create-executiveFunctions-subtest"
	createlanguagefluencysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-languageFluency-subtest"
	createlettercancelationsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-letterCancelation-subtest"
	createverbalmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-verbalMemory-subtest"
	createvisualspatialsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visual-spatial-subtest"
	createvisualmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visualMemory-subtest"
	finishevaluation "neuro.app.jordi/internal/evaluation/application/commands/finish-evaluation"
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

	// (Opcional) Si solo vino YYYY-MM-DD en "to", puedes ajustar a fin de día:
	// if !to.IsZero() && dto.ToDateStr != "" && len(dto.ToDateStr) == len("2006-01-02") {
	// 	to = to.Add(24 * time.Hour).Add(-time.Nanosecond) // 23:59:59.999999999
	// }

	// 4) Construir query de dominio
	query := listevaluations.ListEvaluationsQuery{
		SpecialistID:  dto.SpecialistID,
		FromDate:      from, // zero => sin filtro si tu capa de aplicación lo permite
		ToDate:        to,   // zero => sin filtro
		SearchTerm:    dto.SearchTerm,
		Offset:        offset,
		Limit:         limit,
		OnlyCompleted: true,
	}

	// 5) Handler de aplicación
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

	// 6) Map a API
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
	var command createlanguagefluencysubtest.CreateLanguageFluencySubtestCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating language fluency evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	subtest, err := createlanguagefluencysubtest.CreateLanguageFluencySubtestCommandHandler(c.Request.Context(), command, app.Repositories.EvaluationsRepository, app.Services.LLMService, app.Repositories.LanguageFluencyRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error  when creating language fluency evaluation", err, c.Keys)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"subtest": subtest})
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
