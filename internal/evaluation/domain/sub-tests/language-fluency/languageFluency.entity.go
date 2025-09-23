package LFdomain

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"neuro.app.jordi/internal/evaluation/utils"
)

/* ====== Tu modelo (tal cual) ====== */

type LanguageFluency struct {
	PK                string               `json:"pk"`
	Language          string               `json:"language"`
	Proficiency       string               `json:"proficiency"`
	Category          string               `json:"category"`
	AnswerWords       []string             `json:"answer_words"`
	EvaluationID      string               `json:"evaluation_id"`
	Score             LanguageFluencyScore `json:"score"`
	AssistantAnalysis string               `json:"assistant_analysis"`
	CreatedAt         time.Time            `json:"created_at"`
}

func NewLanguageFluency(language, proficiency, category string, answerWords []string, evaluationID string) (*LanguageFluency, error) {
	if language == "" || proficiency == "" || category == "" || len(answerWords) == 0 || evaluationID == "" {
		return nil, errors.New("invalid input for creating LanguageFluency")
	}
	return &LanguageFluency{
		PK:                uuid.New().String(),
		Language:          language,
		Proficiency:       proficiency,
		Category:          category,
		AnswerWords:       answerWords,
		EvaluationID:      evaluationID,
		CreatedAt:         time.Now().UTC(),
		AssistantAnalysis: "",
	}, nil
}

/* ====== Scoring sencillo y reproducible ====== */

type LanguageFluencyScore struct {
	Score          int     `json:"score"`          // 0..100
	UniqueValid    int     `json:"uniqueValid"`    // válidas únicas en la categoría
	Intrusions     int     `json:"intrusions"`     // no pertenecen a la categoría
	Perseverations int     `json:"perseverations"` // repeticiones
	TotalProduced  int     `json:"totalProduced"`  // len(answer_words)
	WordsPerMinute float64 `json:"wordsPerMinute"` // uniqueValid / (duration/60)
	IntrusionRate  float64 `json:"intrusionRate"`  // intrusions / max(1,totalProduced)
	PersevRate     float64 `json:"persevRate"`     // perseverations / max(1,totalProduced)
}

type LanguageFluencyScoreConfig struct {
	DurationSec          int     // por defecto 60
	MaxExpectedPerMinute int     // cap superior esperado (default 30)
	NormalizeWords       bool    // minúsculas + quitar tildes/diéresis, mapear ñ→n (default true)
	IntrusionPenalty     float64 // default 0.5
	PersevPenalty        float64 // default 0.25
	// Conjunto de términos válidos para la categoría (opcional).
	ValidSet map[string]struct{}
}

func ScoreLanguageFluency(sub LanguageFluency) (LanguageFluencyScore, error) {
	// Defaults
	c := LanguageFluencyScoreConfig{
		DurationSec:          60,
		MaxExpectedPerMinute: 30,
		NormalizeWords:       true,
		IntrusionPenalty:     0.5,
		PersevPenalty:        0.25,
		ValidSet:             nil,
	}

	words := sanitizeList(sub.AnswerWords, c.NormalizeWords)
	totalProduced := len(words)

	seen := make(map[string]int, totalProduced)
	uniqueValid := 0
	intrusions := 0
	persevs := 0

	isValid := func(w string) bool {
		if c.ValidSet == nil {
			// Si no pasas diccionario, todo cuenta como “válido”
			return true
		}
		_, ok := c.ValidSet[w]
		return ok
	}

	for _, w := range words {
		if w == "" {
			continue
		}
		seen[w]++
		if seen[w] > 1 {
			persevs++
			continue
		}
		if isValid(w) {
			uniqueValid++
		} else {
			intrusions++
		}
	}

	minutes := math.Max(1.0/60.0, float64(c.DurationSec)/60.0)
	wpm := float64(uniqueValid) / minutes
	intrRate := float64(intrusions) / math.Max(1, float64(totalProduced))
	persevRate := float64(persevs) / math.Max(1, float64(totalProduced))

	// Normalización 0..100
	maxCap := float64(c.MaxExpectedPerMinute)
	rateIdx := utils.Clamp01(wpm / maxCap)
	penalty := c.IntrusionPenalty*float64(intrusions)/maxCap + c.PersevPenalty*float64(persevs)/maxCap
	score01 := utils.Clamp01(rateIdx - penalty)
	score := int(math.Round(100 * score01))

	return LanguageFluencyScore{
		Score:          score,
		UniqueValid:    uniqueValid,
		Intrusions:     intrusions,
		Perseverations: persevs,
		TotalProduced:  totalProduced,
		WordsPerMinute: wpm,
		IntrusionRate:  intrRate,
		PersevRate:     persevRate,
	}, nil
}

/* ====== utils ====== */

func sanitizeList(xs []string, normalize bool) []string {
	out := make([]string, 0, len(xs))
	for _, s := range xs {
		s = strings.TrimSpace(s)
		if normalize {
			s = strings.ToLower(s)
			s = utils.ReplaceAccentsES(s) // quita tildes/diéresis y mapea ñ→n
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
