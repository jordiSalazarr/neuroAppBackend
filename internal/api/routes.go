package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	auth "neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/auth/infra"
	"neuro.app.jordi/internal/evaluation/domain"
	"neuro.app.jordi/internal/evaluation/services"
	"neuro.app.jordi/internal/shared/encryption"
	jwtService "neuro.app.jordi/internal/shared/jwt"
	"neuro.app.jordi/internal/shared/logger"
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
	LLMService            domain.LLMService
	FileFormater          domain.FileFormaterService
	MailService           mail.MailProvider
	EncryptionService     auth.EncryptionService
	EvaluationsRepository domain.EvaluationsRepository //TODO: add this implementation
	UserRepository        auth.UserRepository          //TODO: add this implementation
	Logger                logger.Logger
	JwtService            *jwtService.Service
}

func NewApp() *App {
	return &App{
		LLMService:        services.NewOpenAIService(),
		FileFormater:      services.NewFileFormatter(),
		MailService:       mail.NewMailer(),
		EncryptionService: encryption.NewEncryptionService(),
		Logger:            logger.NewLogger(),
		JwtService:        jwtService.New(),
		UserRepository:    infra.NewUserInMemory(), //TODO: implement this with a real repository (sql)
	}
}

func (app *App) SetupRouter() *gin.Engine {
	router := gin.Default()
	router.Use(rateLimiter, gin.Recovery())

	// Grupo para endpoints relacionados con evaluaciones
	evaluationGroup := router.Group("/v1/evaluations")
	{
		evaluationGroup.POST("/", app.CreateEvaluation)
	}

	// Grupo para otros endpoints (ejemplo)
	userGroup := router.Group("/v1/auth")
	{
		userGroup.POST("/login", app.Login)
		userGroup.POST("/signup", app.SignUp)
		userGroup.PATCH("/verify", app.VerifyUser)

	}

	protectredGroup := router.Group("/v1")
	{
		protectredGroup.Use(midleware.ExtractJWTFromRequest(app.JwtService))
		protectredGroup.GET("/pending-accept-requests", app.GetPendingAcceptRequests)
		protectredGroup.POST("/accept", app.AcceptUser)
	}
	return router
}
