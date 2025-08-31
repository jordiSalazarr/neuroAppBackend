package createlanguagefluencysubtest

import (
	"context"

	"neuro.app.jordi/internal/evaluation/domain"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
)

func CreateLanguageFluencySubtestCommandHandler(ctx context.Context, cmd CreateLanguageFluencySubtestCommand, evaluationRepo domain.EvaluationsRepository, llmService domain.LLMService, languageFluencyRepo LFdomain.LanguageFluencyRepository) (LFdomain.LanguageFluency, error) {
	evaluation, err := evaluationRepo.GetByID(ctx, cmd.EvaluationID)

	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}
	languageFluency, err := LFdomain.NewLanguageFluency(cmd.Language, cmd.Proficiency, cmd.Category, cmd.Words, evaluation.PK)
	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}

	score, err := LFdomain.ScoreLanguageFluency(*languageFluency)
	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}
	languageFluency.Score = score

	// res, err := llmService.LanguageFluencyAnalysis(languageFluency, evaluation.PatientAge)
	// if err != nil {
	// 	return LFdomain.LanguageFluency{}, err
	// }
	// languageFluency.AssistantAnalysis = res

	err = languageFluencyRepo.Save(ctx, *languageFluency)
	if err != nil {
		return LFdomain.LanguageFluency{}, err
	}

	return *languageFluency, nil
}
