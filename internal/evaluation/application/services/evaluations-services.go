package services

import (
	"context"
	"errors"

	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	VIMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-memory"
	VPdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/visual-spatial"
)

func PopulateEvaluationWithSubtests(ctx context.Context,
	evaluation *domain.Evaluation,
	verbalMemoryRepository VEMdomain.VerbalMemoryRepository,
	visualMemoryRepository VIMdomain.VisualMemoryRepository,
	executiveFunctionsRepository EFdomain.ExecutiveFunctionsSubtestRepository,
	letterCancellationRepository LCdomain.LetterCancellationRepository,
	languageFluencyRepository LFdomain.LanguageFluencyRepository,
	visualSpatialMemotry VPdomain.ResultRepository,
) error {
	if evaluation == nil {
		return errors.New("populateEvaluationWithSubtests: evaluation is nil")
	}

	var merr error

	// 1) Verbal Memory
	vm, err := verbalMemoryRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.VerbalmemorySubTest = vm
	//TODO: ML needing tests
	// 2) Visual Memory //this is geometric figures
	vim, err := visualMemoryRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.VisualMemorySubTest = vim

	// 3) Executive Functions (TMT, etc.)
	ef, err := executiveFunctionsRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.ExecutiveFunctionSubTest = ef

	// 4) Letter Cancellation (Atención sostenida)
	lc, err := letterCancellationRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.LetterCancellationSubTest = lc

	// 5) Language Fluency
	lf, err := languageFluencyRepository.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.LanguageFluencySubTest = lf
	// 5) Visual Spatial
	vp, err := visualSpatialMemotry.GetByEvaluationID(ctx, evaluation.PK)
	if err != nil {
		return err
	}
	evaluation.VisualSpatialSubTest = *vp

	return merr
}
