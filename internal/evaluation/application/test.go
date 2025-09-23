package commands_test

import (
	"os"

	"neuro.app.jordi/internal/api"
	"neuro.app.jordi/internal/auth/infra"
	infraE "neuro.app.jordi/internal/evaluation/infra"

	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	EFinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/executive-functions"
	LCinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/letter-cancellation"
	VEMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/verbal-memory"
	VIMinfra "neuro.app.jordi/internal/evaluation/infra/sub-tests/visual-memory"
	services "neuro.app.jordi/internal/evaluation/services/openAI"
	"neuro.app.jordi/internal/shared/encryption"
	jwtService "neuro.app.jordi/internal/shared/jwt"
	logging "neuro.app.jordi/internal/shared/logger"
)

func NewMockApp() *api.App {
	mockedRepositories := getAppMockRepositories()
	mockedServices := getAppMockServices()
	return &api.App{
		Repositories: mockedRepositories,
		Services:     mockedServices,
		MaxMemory:    10 << 20, // 10 MB
		Logger:       logging.NewSlogLogger(os.Getenv("environment")),
	}
}

func getAppMockRepositories() api.Repositories {
	return api.Repositories{
		EvaluationsRepository:               infraE.NewMockEvaluationsRepository(),
		LetterCancellationRepository:        LCinfra.NewMockLetterCancellationRepository(),
		VerbalMemorySubtestRepository:       VEMinfra.NewMockVerbalMemoryRepository(),
		LanguageFluencyRepository:           LFdomain.NewLanguageFluencyMock(),
		VisualMemorySubtestRepository:       VIMinfra.NewMockVisualMemoryRepository(),
		ExecutiveFunctionsSubtestRepository: EFinfra.NewMockExecutiveFunctionsRepository(),
		UserRepository:                      infra.NewMockUsersRepository(),
	}
}

func getAppMockServices() api.Services {
	return api.Services{
		LLMService:        services.NewMockOpenAIService(),
		MailService:       nil,
		EncryptionService: encryption.NewEncryptionService(),
		JwtService:        jwtService.New(),
	}
}
