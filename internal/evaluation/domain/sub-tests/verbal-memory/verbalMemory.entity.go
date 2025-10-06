package VEMdomain

import (
	"errors"
	"math"
	"strings"
	"time"

	"github.com/google/uuid"
	"neuro.app.jordi/internal/evaluation/utils"
)

const MaxVerbalMemoryWords = 100
const MaxTimeSinceStart = 3600 // segundos (1 hora)
const ImmediateThreshold = 300 // segundos (5 minutos)

type VerbalMemorySubtype string

const (
	VerbalMemorySubtypeImmediate VerbalMemorySubtype = "immediate"
	VerbalMemorySubtypeDelayed   VerbalMemorySubtype = "delayed"
)

type VerbalMemorySubtest struct {
	Pk               string              `json:"pk"`
	SecondsFromStart int64               `json:"seconds_from_start"`
	GivenWords       []string            `json:"given_words"`
	RecalledWords    []string            `json:"recalled_words"`
	Type             VerbalMemorySubtype `json:"type"`
	EvaluationID     string              `json:"evaluation_id"`
	Score            VerbalMemoryScore   `json:"score"`
	AssistanAnalysis string              `json:"assistan_analysis"`
	CreatedAt        time.Time           `json:"created_at"`
}

type VerbalMemoryScore struct {
	Score             int     `json:"score"`          // 0..100
	Hits              int     `json:"hits"`           // # objetivos recordados (únicos)
	Omissions         int     `json:"omissions"`      // objetivos no recordados
	Intrusions        int     `json:"intrusions"`     // palabras no objetivo
	Perseverations    int     `json:"perseverations"` // repeticiones dentro del mismo ensayo (objetivo o no)
	Accuracy          float64 `json:"accuracy"`       // hits / objetivos
	IntrusionRate     float64 `json:"intrusionRate"`  // intrusions / max(1, len(recalled))
	PerseverationRate float64 `json:"perseverationRate"`
}

func NewVerbalMemorySubtest(evaluationID string, startAt time.Time, givenWords, recalledWords []string, subTypeStr string) (VerbalMemorySubtest, error) {
	if evaluationID == "" {
		return VerbalMemorySubtest{}, errors.New("evaluationID es obligatorio")
	}
	timeSinceStart := time.Since(startAt).Seconds()
	if timeSinceStart < 0 || timeSinceStart > MaxTimeSinceStart {
		return VerbalMemorySubtest{}, errors.New("startAt no puede ser en el futuro o más de 1 hora en el pasado")
	}
	if len(givenWords) == 0 || len(recalledWords) == 0 || len(givenWords) > MaxVerbalMemoryWords || len(recalledWords) > MaxVerbalMemoryWords {
		return VerbalMemorySubtest{}, errors.New("givenWords y recalledWords no pueden estar vacíos y como maximo 100 palabras")
	}
	var subType VerbalMemorySubtype
	switch subTypeStr {
	case "immediate":
		subType = VerbalMemorySubtypeImmediate
	case "delayed":
		subType = VerbalMemorySubtypeDelayed
	}
	return VerbalMemorySubtest{
		Pk:               uuid.New().String(),
		SecondsFromStart: int64(timeSinceStart),
		GivenWords:       givenWords,
		RecalledWords:    recalledWords,
		Type:             subType,
		EvaluationID:     evaluationID,
		Score:            VerbalMemoryScore{},
		AssistanAnalysis: "",
		CreatedAt:        time.Now().UTC(),
	}, nil
}

func ScoreVerbalMemory(sub VerbalMemorySubtest) (VerbalMemoryScore, error) {
	if len(sub.GivenWords) == 0 {
		return VerbalMemoryScore{}, errors.New("given_words vacío")
	}
	// cfg por defecto
	ip := 0.5
	pp := 0.25
	norm := true

	// Normaliza
	given := normalizeList(sub.GivenWords, norm)
	recalled := normalizeList(sub.RecalledWords, norm)

	// Set de objetivos (únicos) y tracking de uso
	targetSet := make(map[string]struct{}, len(given))
	for _, w := range given {
		if w == "" {
			continue
		}
		targetSet[w] = struct{}{}
	}
	targetCount := len(targetSet)

	hitUsed := make(map[string]bool, targetCount) // marca si ya contamos ese objetivo
	seen := make(map[string]int, len(recalled))   // para perseveraciones
	var hits, intrusions, perseverations int

	for _, w := range recalled {
		if w == "" {
			continue
		}
		seen[w]++
		if seen[w] > 1 {
			// Repetición de cualquier palabra en el mismo ensayo
			perseverations++
			continue
		}
		if _, ok := targetSet[w]; ok {
			// Objetivo: cuenta solo una vez por palabra objetivo
			if !hitUsed[w] {
				hits++
				hitUsed[w] = true
			} else {
				// si lo vuelve a decir (misma palabra objetivo) caerá en perseveración ya controlada arriba
			}
		} else {
			// No objetivo => intrusión
			intrusions++
		}
	}

	omissions := targetCount - hits
	if omissions < 0 {
		omissions = 0
	}

	recN := len(recalled)
	if recN == 0 {
		recN = 1
	} // evita división por 0

	accuracy := 0.0
	if targetCount > 0 {
		accuracy = float64(hits) / float64(targetCount)
	}
	intrusionRate := float64(intrusions) / float64(recN)
	perseverationRate := float64(perseverations) / float64(recN)

	// Score 0..100
	// Fórmula: precisión menos penalizaciones relativas al nº de objetivos.
	score01 := accuracy - ip*float64(intrusions)/float64(math.Max(1, float64(targetCount))) - pp*float64(perseverations)/float64(math.Max(1, float64(targetCount)))
	if score01 < 0 {
		score01 = 0
	}
	if score01 > 1 {
		score01 = 1
	}
	score := int(score01*100 + 0.5)

	return VerbalMemoryScore{
		Score:             score,
		Hits:              hits,
		Omissions:         omissions,
		Intrusions:        intrusions,
		Perseverations:    perseverations,
		Accuracy:          accuracy,
		IntrusionRate:     intrusionRate,
		PerseverationRate: perseverationRate,
	}, nil
}

func normalizeList(xs []string, do bool) []string {
	out := make([]string, 0, len(xs))
	for _, s := range xs {
		s = strings.TrimSpace(s)
		if do {
			s = strings.ToLower(s)
			s = utils.ReplaceAccentsES(s)
		}
		if s != "" {
			out = append(out, s)
		}
	}
	return out
}
