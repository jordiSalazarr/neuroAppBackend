package services

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
)

// =============== PUBLIC API =================

func formatEvaluationForLLM(ev domain.Evaluation) string {
	safe := sanitizeForLLM(ev)
	summary := buildLLMSummary(safe)

	raw, _ := json.Marshal(safe)

	payload := map[string]any{
		"LLM_INPUT_SUMMARY":         summary, // <- prioriza esto en tu prompt
		"INPUT_COMPLETO_SANITIZADO": json.RawMessage(raw),
	}

	out, _ := json.Marshal(payload)
	return string(out)
}

// =============== SANITIZE ===================

func sanitizeForLLM(ev domain.Evaluation) domain.Evaluation {
	ev.PatientName = ""
	ev.SpecialistMail = ""
	ev.SpecialistID = ""
	ev.StorageKey = ""
	ev.StorageURL = ""
	return ev
}

// =============== SUMMARY DTOs ===============

type LLMVisualMemorySummary struct {
	Present      bool   `json:"present"`
	Score0to2    int    `json:"score_0_2"`     // de ev.VisualMemorySubTest.Score.Val
	VMNorm0to100 int    `json:"vm_norm_0_100"` // score_0_2 normalizado a 0..100
	Note         string `json:"note"`          // ev.VisualMemorySubTest.Note.Val
	Comment      string `json:"comment,omitempty"`
}

type LLMLettersSummary struct {
	Present        bool    `json:"present"`
	Accuracy       float64 `json:"accuracy"`
	Omissions      int     `json:"omissions"`
	OmissionsRate  float64 `json:"omissionsRate"`
	CommissionRate float64 `json:"commissionRate"`
	HitsPerMin     float64 `json:"hitsPerMin"`
	ErrorsPerMin   float64 `json:"errorsPerMin"`
	CpPerMin       float64 `json:"cpPerMin"`
	TimeSec        int     `json:"time_sec"`
}

type LLMVerbalMemorySummary struct {
	Present           bool    `json:"present"`
	Score0to100       int     `json:"score_0_100"`
	Hits              int     `json:"hits"`
	Omissions         int     `json:"omissions"`
	Intrusions        int     `json:"intrusions"`
	Perseverations    int     `json:"perseverations"`
	Accuracy          float64 `json:"accuracy"`
	IntrusionRate     float64 `json:"intrusionRate"`
	PerseverationRate float64 `json:"perseverationRate"`
}

type LLMExecOnePart struct {
	Present     bool    `json:"present"`
	NumberItems int     `json:"numberOfItems"`
	Errors      int     `json:"totalErrors"`
	Correct     int     `json:"totalCorrect"`
	Clicks      int     `json:"totalClicks"`
	DurationSec float64 `json:"durationSec"` // del score o computado desde TotalTime
	SpeedIndex  float64 `json:"speedIndex"`
}

type LLMExecutiveSummary struct {
	TMTA      LLMExecOnePart `json:"tmt_a"`        // type == "a"
	TMTAplusB LLMExecOnePart `json:"tmt_a_plus_b"` // type == "a+b"
}

type LLMLanguageFluencySummary struct {
	Present     bool   `json:"present"`
	Score0to100 int    `json:"score_0_100"`
	UniqueValid int    `json:"uniqueValid,omitempty"` // si lo añades en tu modelo, aquí queda el sitio
	Words       int    `json:"words"`
	Language    string `json:"language"`
	Category    string `json:"category"`
	Proficiency string `json:"proficiency"`
}

type LLMVisualSpatialSummary struct {
	Present bool   `json:"present"`
	Score   int    `json:"score"`
	Note    string `json:"note"`
	Alias   string `json:"alias"`
}

type LLMSummary struct {
	LettersCancellation LLMLettersSummary         `json:"letters_cancellation"`
	VisualMemory        LLMVisualMemorySummary    `json:"visual_memory"`
	VerbalMemory        LLMVerbalMemorySummary    `json:"verbal_memory"`
	ExecutiveFunctions  LLMExecutiveSummary       `json:"executive_functions"`
	LanguageFluency     LLMLanguageFluencySummary `json:"language_fluency"`
	VisualSpatial       LLMVisualSpatialSummary   `json:"visual_spatial"`
}

// =============== BUILD SUMMARY ==============

func buildLLMSummary(ev domain.Evaluation) LLMSummary {
	return LLMSummary{
		LettersCancellation: buildLetters(ev),
		VisualMemory:        buildVisualMemory(ev),
		VerbalMemory:        buildVerbalMemory(ev),
		ExecutiveFunctions:  buildExecutive(ev),
		LanguageFluency:     buildLanguage(ev),
		VisualSpatial:       buildVisualSpatial(ev),
	}
}

func buildLetters(ev domain.Evaluation) LLMLettersSummary {
	lc := ev.LetterCancellationSubTest
	return LLMLettersSummary{
		Present:        lc.PK != "",
		Accuracy:       lc.CancellationScore.Accuracy,
		Omissions:      lc.CancellationScore.Omissions,
		OmissionsRate:  lc.CancellationScore.OmissionsRate,
		CommissionRate: lc.CancellationScore.CommissionRate,
		HitsPerMin:     lc.CancellationScore.HitsPerMin,
		ErrorsPerMin:   lc.CancellationScore.ErrorsPerMin,
		CpPerMin:       lc.CancellationScore.CpPerMin,
		TimeSec:        lc.TimeInSecs,
	}
}

func buildVisualMemory(ev domain.Evaluation) LLMVisualMemorySummary {
	vm := ev.VisualMemorySubTest
	score := clamp(vm.Score.Val, 0, 2)
	vmNorm := int(math.Round(float64(score) / 2.0 * 100.0))
	comment := ""
	if score == 0 && strings.TrimSpace(vm.Note.Val) == "" {
		comment = "Score=0 puede ser rendimiento muy bajo o fallo de copia; revisar contexto/nota del evaluador."
	}
	return LLMVisualMemorySummary{
		Present:      vm.PK != "" || score >= 0 || vm.Note.Val != "",
		Score0to2:    score,
		VMNorm0to100: vmNorm,
		Note:         vm.Note.Val,
		Comment:      comment,
	}
}

func buildVerbalMemory(ev domain.Evaluation) LLMVerbalMemorySummary {
	vm := ev.VerbalmemorySubTest
	return LLMVerbalMemorySummary{
		Present:           vm.Pk != "",
		Score0to100:       vm.Score.Score,
		Hits:              vm.Score.Hits,
		Omissions:         vm.Score.Omissions,
		Intrusions:        vm.Score.Intrusions,
		Perseverations:    vm.Score.Perseverations,
		Accuracy:          vm.Score.Accuracy,
		IntrusionRate:     vm.Score.IntrusionRate,
		PerseverationRate: vm.Score.PerseverationRate,
	}
}

func buildExecutive(ev domain.Evaluation) LLMExecutiveSummary {
	var out LLMExecutiveSummary

	for _, part := range ev.ExecutiveFunctionSubTest {
		t := strings.ToLower(fmt.Sprintf("%v", part.Type)) // soporta enum/string
		one := LLMExecOnePart{
			Present:     part.PK != "",
			NumberItems: part.NumberOfItems,
			Errors:      part.TotalErrors,
			Correct:     part.TotalCorrect,
			Clicks:      part.TotalClicks,
			DurationSec: durationSec(part),
			SpeedIndex:  part.Score.SpeedIndex,
		}

		switch t {
		case "a":
			out.TMTA = one
		case "a+b", "a_plus_b", "ab":
			out.TMTAplusB = one
		default:
			// si en el futuro agregas otros tipos, puedes agregarlos aquí
		}
	}
	return out
}

func buildLanguage(ev domain.Evaluation) LLMLanguageFluencySummary {
	lf := ev.LanguageFluencySubTest
	words := len(lf.AnswerWords)
	return LLMLanguageFluencySummary{
		Present:     lf.PK != "",
		Score0to100: lf.Score.Score,
		UniqueValid: 0, // si luego añades este dato al modelo, aquí se mapea
		Words:       words,
		Language:    lf.Language,
		Category:    lf.Category,
		Proficiency: lf.Proficiency,
	}
}

func buildVisualSpatial(ev domain.Evaluation) LLMVisualSpatialSummary {
	vs := ev.VisualSpatialSubTest
	return LLMVisualSpatialSummary{
		Present: true,
		Score:   vs.Score.Val,
		Note:    vs.Note.Val,
		Alias:   "Clock Drawing Test (CDT)",
	}
}

// =============== UTILS ======================

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func durationSec(s EFdomain.ExecutiveFunctionsSubtest) float64 {
	// Prioriza el campo ya calculado en score; si no, usa TotalTime
	if s.Score.DurationSec > 0 {
		return s.Score.DurationSec
	}
	if s.TotalTime > 0 {
		return float64(s.TotalTime) / float64(time.Second)
	}
	return 0
}
