package api

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	createevaluation "neuro.app.jordi/internal/evaluation/application/commands/create-evaluation"
	createexecutivefunctionssubtest "neuro.app.jordi/internal/evaluation/application/commands/create-executiveFunctions-subtest"
	createlanguagefluencysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-languageFluency-subtest"
	createlettercancelationsubtest "neuro.app.jordi/internal/evaluation/application/commands/create-letterCancelation-subtest"
	createverbalmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-verbalMemory-subtest"
	createvisualmemorysubtest "neuro.app.jordi/internal/evaluation/application/commands/create-visualMemory-subtest"
	finishevaluation "neuro.app.jordi/internal/evaluation/application/commands/finish-evaluation"
	getevaluation "neuro.app.jordi/internal/evaluation/application/queries/get-evaluation"
	"neuro.app.jordi/internal/evaluation/domain"
	reports "neuro.app.jordi/internal/evaluation/domain/services"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	"neuro.app.jordi/internal/evaluation/utils"
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

func (app *App) GetEvaluation(c *gin.Context) {
	var query getevaluation.GetEvaluationQuery
	id := c.Params.ByName("id")
	query.EvaluationID = id
	evaluation, err := getevaluation.GetEvaluationQueryHandler(c.Request.Context(), query, app.Repositories.EvaluationsRepository)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error getting evaluation", err, nil)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"success":    "Evaluation created",
		"evaluation": domainToAPIEvaluation(evaluation),
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
		app.Repositories.VisualMemorySubtestRepository, app.Repositories.ExecutiveFunctionsSubtestRepository, app.Repositories.LetterCancellationRepository, app.Repositories.LanguageFluencyRepository)
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

func (app *App) VisualMemorySubtest(c *gin.Context) {
	// (Opcional) límite total del cuerpo para evitar abusos
	if app.MaxMemory > 0 {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, app.MaxMemory)
	}

	// Limita cuánto del multipart se guarda en RAM; el resto va a /tmp
	if err := c.Request.ParseMultipartForm(app.MaxMemory); err != nil {
		app.Logger.Error(c.Request.Context(), "error parsing when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid multipart: " + err.Error()})
		return
	}
	defer func() {
		if c.Request.MultipartForm != nil {
			_ = c.Request.MultipartForm.RemoveAll()
		}
	}()

	evalID := c.PostForm("evaluation_id")
	figure := c.PostForm("figure_name")
	capturedStr := c.PostForm("captured_at")

	var capturedAt time.Time
	if capturedStr != "" {
		t, err := time.Parse(time.RFC3339, capturedStr)
		if err != nil {
			app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
			c.JSON(http.StatusBadRequest, gin.H{"error": "captured_at must be RFC3339"})
			return
		}
		capturedAt = t
	}

	// Archivo
	fileHeader, err := c.FormFile("image")
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing image: " + err.Error()})
		return
	}
	file, err := fileHeader.Open()
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot open image: " + err.Error()})
		return
	}
	defer file.Close()

	// Lee con límite (por-archivo)
	raw, err := utils.ReadAllLimited(file, (app.MaxMemory))
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": err.Error()})
		return
	}

	cmd := createvisualmemorysubtest.CreateBVMTSubtestCommand{
		EvaluationID: evalID,
		FigureName:   figure,
		CapturedAt:   capturedAt,
		ImageBytes:   raw,
		ContentType:  fileHeader.Header.Get("Content-Type"), // puede venir vacío
	}

	sub, err := createvisualmemorysubtest.CreateVisualMemoryCommandHandler(
		c.Request.Context(),
		cmd,
		app.Repositories.VisualMemorySubtestRepository,
		app.ImageStorage,
		int(app.MaxMemory),
		app.Scorer,
		app.Services.TemplateResolver,
	)
	if err != nil {
		app.Logger.Error(c.Request.Context(), "error when creating visual memory evaluation", err, c.Keys)
		status := http.StatusBadRequest
		switch {
		case errors.Is(err, VIMdomain.ErrTooLarge):
			status = http.StatusRequestEntityTooLarge
		case errors.Is(err, VIMdomain.ErrEmptyEvaluationID),
			errors.Is(err, VIMdomain.ErrEmptyFigureName),
			errors.Is(err, VIMdomain.ErrNoImageBytes),
			errors.Is(err, VIMdomain.ErrUnsupportedMIME):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, sub)
}
