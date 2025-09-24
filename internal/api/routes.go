package api

import (
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	authD "neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/evaluation/domain"
	services "neuro.app.jordi/internal/evaluation/services/openAI"

	authI "neuro.app.jordi/internal/auth/infra"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
	"neuro.app.jordi/internal/evaluation/infra"
	EFinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/executive-functions"
	LFinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/language-fluency"
	LCinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/letter-cancellation"
	VEMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/verbal-memory"
	VIMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/visual-memory"
	INFRAvisualspatial "neuro.app.jordi/internal/evaluation/infra/sub-tests/visual-spatial"
	"neuro.app.jordi/internal/shared/encryption"
	jwtService "neuro.app.jordi/internal/shared/jwt"
	logging "neuro.app.jordi/internal/shared/logger"
	"neuro.app.jordi/internal/shared/mail"
	"neuro.app.jordi/internal/shared/midleware"
)

var limiter = rate.NewLimiter(100, 5)

func rateLimiter(c *gin.Context) {
	if !limiter.Allow() {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
		c.Abort()
		return
	}
	c.Next()
}

type App struct {
	Repositories Repositories
	Services     Services
	MaxMemory    int64 // MaxMemory for multipart forms, e.g., 8 << 20 is 8 MB
	// ImageStorage VIMdomain.ImageStorage
	Logger logging.Logger
}
type Repositories struct {
	EvaluationsRepository               domain.EvaluationsRepository                 //TODO: add this implementation
	LetterCancellationRepository        LCdomain.LetterCancellationRepository        //TODO: add this implementation
	VerbalMemorySubtestRepository       VEMdomain.VerbalMemoryRepository             //TODO: add this implementation
	LanguageFluencyRepository           LFdomain.LanguageFluencyRepository           //TODO: add this implementation
	VisualMemorySubtestRepository       VIMdomain.VisualMemoryRepository             //TODO: add this implementation
	ExecutiveFunctionsSubtestRepository EFdomain.ExecutiveFunctionsSubtestRepository //TODO: add this implementation
	VisualSpatialRepository             VPdomain.ResultRepository
	UserRepository                      authD.UserRepository
}
type Services struct {
	LLMService        domain.LLMService
	MailService       mail.MailProvider
	JwtService        *jwtService.Service
	EncryptionService authD.EncryptionService
	// TemplateResolver  VIMdomain.TemplateResolver
	FileFormater domain.FileFormaterService
}

func getAppRepositories(db *sql.DB) Repositories {

	return Repositories{
		EvaluationsRepository:               infra.NewEvaluationsMYSQLRepository(db),
		LetterCancellationRepository:        LCinfra.NewMYSQLLetterCancellationRepository(db),
		VerbalMemorySubtestRepository:       VEMinfra.NewVerbalMemoryMYSQLRepository(db),
		ExecutiveFunctionsSubtestRepository: EFinfra.NewExecutiveFunctionsSubtestMYSQLRepository(db),
		LanguageFluencyRepository:           LFinfra.NewLanguageFluencyMYSQLRepository(db),
		VisualSpatialRepository:             INFRAvisualspatial.NewVisualSpatialMYSQLRepo(db),
		VisualMemorySubtestRepository:       VIMinfra.NewVisualMemoryMYSQLRepository(db),
		UserRepository:                      authI.NewUseMYSQLRepository(db),
	}
}

func getAppServices() Services {
	return Services{
		LLMService:  services.NewOpenAIService(),
		MailService: mail.NewMailer(),
		// TemplateResolver:  services.LocalTemplateResolver{},
		EncryptionService: encryption.NewEncryptionService(),
		JwtService:        jwtService.New(),
	}
}
func NewApp(db *sql.DB) *App {
	appRepositories := getAppRepositories(db)
	appServices := getAppServices()
	return &App{
		// FileFormater:      services.NewFileFormatter(),
		Repositories: appRepositories,
		Services:     appServices,
		// ImageStorage: VIMinfra.NewLocalImageStorage("./images"),
		MaxMemory: 10 << 20, // 10 MB
		Logger:    logging.NewSlogLogger(os.Getenv("environment")),
	}
}

func (app *App) SetupRouter() *gin.Engine {
	r := gin.New()
	r.SetTrustedProxies(nil)
	r.Use(gin.Logger(), gin.Recovery())

	corsCfg := cors.Config{
		AllowOrigins: []string{
			"https://neuro-next-web-production.up.railway.app",
			"http://localhost:8400",
		},
		AllowMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		// ¡NO uses "*"! Enumera explícitamente lo que envías desde el front:
		AllowHeaders: []string{
			"Origin",
			"Content-Type",
			"Accept",
			"Authorization",
			"X-Requested-With",
		},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	r.Use(cors.New(corsCfg))

	// Health BEFORE any services
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	// --- your groups below ---
	//TODO: add auth middleware to protect routes
	// auth := midleware.ExtractJWTFromRequest(app.Services.JwtService)
	loggConfig := midleware.NewAccessLogConfig()
	r.Use(midleware.AccessLog(app.Logger, loggConfig))
	r.Use(rateLimiter) // if you want a global rate limiter
	eval := r.Group("/v1/evaluations")
	{
		eval.POST("", app.CreateEvaluation)
		eval.POST("/letter-cancellation", app.CreateLetterCancellationSubtest)
		eval.POST("/verbal-memory", app.VerbalMemorySubtest)
		eval.POST("/executive-functions", app.ExecutiveFunctionsSubtest)
		eval.POST("/language-fluency", app.LanguageFluencySubtest)
		eval.POST("/visual-memory", app.CreateVisualMemorySubtest)
		eval.POST("/visual-spatial", app.CreateVisualSpatialSubtest)
		eval.POST("/finish-evaluation", app.FinnishEvaluation)
		eval.GET("/:id", app.GetEvaluation)
		eval.GET("", app.ListEvaluations)
	}

	user := r.Group("/v1/auth")
	{
		user.POST("/signup", app.SignUp)
		user.POST("/user/:mail/:name", app.RegisterUserInfo)
	}

	return r
}
