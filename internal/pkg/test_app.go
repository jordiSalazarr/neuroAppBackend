package pkg

import (
	"os"

	authD "neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/evaluation/domain"
	EFinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/executive-functions"
	LCinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/letter-cancellation"
	VEMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/verbal-memory"
	VIMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/visual-memory"
	INFRAvisualspatial "neuro.app.jordi/internal/evaluation/infra/sub-tests/visual-spatial"

	"neuro.app.jordi/internal/auth/infra"
	infraE "neuro.app.jordi/internal/evaluation/infra"

	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
	services "neuro.app.jordi/internal/evaluation/services/openAI"
	"neuro.app.jordi/internal/shared/encryption"
	jwtService "neuro.app.jordi/internal/shared/jwt"
	logging "neuro.app.jordi/internal/shared/logger"
	"neuro.app.jordi/internal/shared/mail"
)

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

func getAppMockRepositories() Repositories {
	return Repositories{
		EvaluationsRepository:               infraE.NewMockEvaluationsRepository(),
		LetterCancellationRepository:        LCinfra.NewMockLetterCancellationRepository(),
		VerbalMemorySubtestRepository:       VEMinfra.NewMockVerbalMemoryRepository(),
		LanguageFluencyRepository:           LFdomain.NewLanguageFluencyMock(),
		VisualMemorySubtestRepository:       VIMinfra.NewMockVisualMemoryRepository(),
		ExecutiveFunctionsSubtestRepository: EFinfra.NewMockExecutiveFunctionsRepository(),
		VisualSpatialRepository:             INFRAvisualspatial.NewMockVisualSpatialRepository(),

		UserRepository: infra.NewMockUsersRepository(),
	}
}

func getAppMockServices() Services {
	return Services{
		LLMService:        services.NewMockOpenAIService(),
		MailService:       nil,
		EncryptionService: encryption.NewEncryptionService(),
		JwtService:        jwtService.New(),
	}
}

func NewMockApp() *App {
	appMockRepositories := getAppMockRepositories()
	appMockServices := getAppMockServices()
	return &App{
		// FileFormater:      services.NewFileFormatter(),
		Repositories: appMockRepositories,
		Services:     appMockServices,
		// ImageStorage: VIMinfra.NewLocalImageStorage("./images"),
		MaxMemory: 10 << 20, // 10 MB
		Logger:    logging.NewSlogLogger(os.Getenv("environment")),
	}
}
