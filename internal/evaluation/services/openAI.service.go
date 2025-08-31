package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"neuro.app.jordi/internal/evaluation/domain"
	EFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/executive-functions"
	LFdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/language-fluency"
	LCdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/letter-cancellation"
	VEMdomain "neuro.app.jordi/internal/evaluation/domain/sub-tests/verbal-memory"
	"neuro.app.jordi/internal/shared/config"
)

type OpenAIService struct {
	client *openai.Client
	ApiKey string
}

func NewOpenAIService() OpenAIService {
	apiKey := config.GetConfig().OpenAIKey
	if apiKey == "" {
		return OpenAIService{}
	}
	return OpenAIService{
		client: openai.NewClient(apiKey),
		ApiKey: apiKey,
	}
}

// TODO: this will be the final prompt for the evaluation analysis
func (oa OpenAIService) GenerateAnalysis(ev domain.Evaluation) (string, error) {
	// 1) Sanitiza y formatea SOLO lo necesario (sin datos personales)
	safe := sanitizeEvaluation(ev)
	formattedEval := formatEvaluationAsText(safe)

	prompt := fmt.Sprintf(
		`Eres un/a neuropsicólogo/a clínico especializado/a en enfermedad de Parkinson avanzada.
Vas a analizar una evaluación **anónima** compuesta por subtests estandarizados. 
Trabaja únicamente con los datos proporcionados (no inventes, no infieras identidades ni demografía) y **omite cualquier referencia personal**.

REGLAS CRÍTICAS
- Considera SOLO subtests con puntuación > 0 (si Score=0 o faltan datos, ignóralos en las conclusiones).
- Cuando la métrica sea “Score” (0–100), valores más altos = mejor rendimiento.
- En métricas de error o tasa (p. ej., intrusionsRate, perseverations, commissionRate, omissionsRate), valores más altos = peor rendimiento.
- Si hay discrepancias entre métricas de un mismo subtest, explica la posible causa (velocidad vs precisión, fatiga, impulsividad, etc.).
- Si un subtest está “pending”, “processing” o sin Score, indícalo como “sin datos” y NO lo uses para conclusiones.

DOMINIOS Y MÉTRICAS (resumen operativo)
1) Atención Sostenida — Letters Cancellation
   - Score (0–100, mayor=mejor), Accuracy, Omissions, CommissionRate, Hits/Errors per min, CpPerMin.
   - Déficit típico: ↑omissions/commissionRate y ↓accuracy/score → perfil atencional.
2) Memoria Visual — BVMT (BVMT-R)
   - Score.FinalScore (0–100), apoyo de IoU/SSIM/PSNR para calidad/parecido.
   - Déficit típico: FinalScore bajo; si calidad gráfica (IoU/SSIM/PSNR) es muy baja, avisar posible sesgo de captura.
3) Memoria Verbal — VerbalMemory
   - Score (0–100), Hits, Omissions, Intrusions, Perseverations, Accuracy, IntrusionRate, PerseverationRate.
   - Déficit típico amnésico: ↓score/accuracy con ↑omissions; intrusions/perseverations orientan a control ejecutivo/monitorización.
4) Funciones Ejecutivas — ExecutiveFunctions (p. ej. TMT A/B)
   - Score (0–100), Accuracy, SpeedIndex, CommissionRate, DurationSec.
   - Déficit ejecutivo: ↓score/accuracy/speedIndex con ↑commissionRate/tiempos.
5) Fluencia Verbal — LanguageFluency (p. ej., semántica)
   - Score (0–100), UniqueValid, Intrusions, Perseverations, WordsPerMinute, IntrusionRate, PersevRate.
   - Déficit léxico/ejecutivo: ↓uniqueValid/WPM, ↑intrusions/perseverations.

UMBRAL HEURÍSTICO (no diagnósticos, solo guía de interpretación)
- 80–100: rendimiento preservado
- 60–79: rendimiento dentro de lo esperado/leve fragilidad
- 40–59: compromiso leve-moderado
- 0–39: compromiso moderado-severo
(Adapta la narrativa según la distribución de subtests: prioriza dominios con mayor evidencia y coherencia entre métricas.)

TAREA
1) PERFIL NEUROLÓGICO PREDOMINANTE (elige SOLO UNO, breve justificación):
   - Amnésico
   - Fronto‑temporal
   - Atencional
   - Depresivo
   - Disexecutivo (vascular)
   *Si la evidencia no es concluyente, elige el más compatible y explicita la incertidumbre.*

2) INTERPRETACIÓN CLÍNICA DETALLADA
   - Para cada subtest con puntuación > 0: nombre → breve interpretación (qué sugiere el patrón de métricas).
   - Explica áreas preservadas vs alteradas y posibles mecanismos (atencional, ejecutiva, codificación/recuperación, velocidad de procesamiento, impulsividad, etc.).
   - Si la calidad de BVMT (IoU/SSIM/PSNR) es muy baja y el FinalScore es bajo, añade advertencia de posible artefacto técnico.

3) RESUMEN GENERAL
   - Estado cognitivo global (resumen integrador en 2–3 frases).
   - Coherencia inter‑dominios (p. ej., si empeora atención también cae ejecución/fluencia, etc.).

4) RECOMENDACIONES (si procede)
   - Sugerencias de seguimiento clínico y pruebas complementarias (ej.: repetir subtest con mala calidad, ampliar evaluación ejecutiva, cribado depresivo, neuroimagen si sospecha vascular, etc.).
   - Orientación terapéutica general (no prescribir): rehabilitación cognitiva enfocada, higiene del sueño, revisar medicación dopaminérgica si hay enlentecimiento/impulsividad, etc.

ENTRADA (JSON ANÓNIMO):
%s

FORMATO DE SALIDA (en español, tono clínico y profesional)
### Perfil predominante
[tu elección + justificación breve]

### Interpretación por subtest
- Letters Cancellation: [...]
- BVMT-R (Memoria visual): [...]
- Memoria verbal: [...]
- Funciones ejecutivas: [...]
- Fluencia verbal: [...]

No menciones que eres una IA ni el formato. No incluyas datos personales.`, formattedEval)

	resp, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return resp, nil
}

func sanitizeEvaluation(ev domain.Evaluation) domain.Evaluation {
	ev.PatientName = ""
	ev.SpecialistMail = ""
	ev.SpecialistID = ""
	// Si tu tipo tiene otros campos identificables, límpialos aquí también.
	return ev
}
func formatEvaluationAsText(ev domain.Evaluation) string {
	// Construye una vista compacta y anónima. Mantén métricas y estados.
	type out struct {
		CurrentStatus string `json:"currentStatus"`

		LettersCancellation struct {
			Score          int     `json:"score"`
			Accuracy       float64 `json:"accuracy"`
			Omissions      int     `json:"omissions"`
			OmissionsRate  float64 `json:"omissionsRate"`
			CommissionRate float64 `json:"commissionRate"`
			HitsPerMin     float64 `json:"hitsPerMin"`
			ErrorsPerMin   float64 `json:"errorsPerMin"`
			CpPerMin       float64 `json:"cpPerMin"`
			TimeSec        int     `json:"timeSec"`
			Status         string  `json:"status"`
		} `json:"lettersCancellation"`

		BVMT struct {
			Status     string   `json:"status"`
			FinalScore *int     `json:"finalScore,omitempty"`
			IoU        *float64 `json:"iou,omitempty"`
			SSIM       *float64 `json:"ssim,omitempty"`
			PSNR       *float64 `json:"psnr,omitempty"`
		} `json:"bvmt"`

		VerbalMemory struct {
			Score             int     `json:"score"`
			Hits              int     `json:"hits"`
			Omissions         int     `json:"omissions"`
			Intrusions        int     `json:"intrusions"`
			Perseverations    int     `json:"perseverations"`
			Accuracy          float64 `json:"accuracy"`
			IntrusionRate     float64 `json:"intrusionRate"`
			PerseverationRate float64 `json:"perseverationRate"`
			Type              string  `json:"type"`
		} `json:"verbalMemory"`

		Executive struct {
			Score          int     `json:"score"`
			Accuracy       float64 `json:"accuracy"`
			SpeedIndex     float64 `json:"speedIndex"`
			CommissionRate float64 `json:"commissionRate"`
			DurationSec    float64 `json:"durationSec"`
			Type           string  `json:"type"`
		} `json:"executive"`

		Fluency struct {
			Score          int     `json:"score"`
			UniqueValid    int     `json:"uniqueValid"`
			Intrusions     int     `json:"intrusions"`
			Perseverations int     `json:"perseverations"`
			TotalProduced  int     `json:"totalProduced"`
			WordsPerMin    float64 `json:"wordsPerMinute"`
			IntrusionRate  float64 `json:"intrusionRate"`
			PersevRate     float64 `json:"persevRate"`
			Category       string  `json:"category"`
		} `json:"fluency"`
	}

	var o out
	o.CurrentStatus = string(ev.CurrentStatus)

	// Letters Cancellation
	o.LettersCancellation.Score = ev.LetterCancellationSubTest.CancellationScore.Score
	o.LettersCancellation.Accuracy = ev.LetterCancellationSubTest.CancellationScore.Accuracy
	o.LettersCancellation.Omissions = ev.LetterCancellationSubTest.CancellationScore.Omissions
	o.LettersCancellation.OmissionsRate = ev.LetterCancellationSubTest.CancellationScore.OmissionsRate
	o.LettersCancellation.CommissionRate = ev.LetterCancellationSubTest.CancellationScore.CommissionRate
	o.LettersCancellation.HitsPerMin = ev.LetterCancellationSubTest.CancellationScore.HitsPerMin
	o.LettersCancellation.ErrorsPerMin = ev.LetterCancellationSubTest.CancellationScore.ErrorsPerMin
	o.LettersCancellation.CpPerMin = ev.LetterCancellationSubTest.CancellationScore.CpPerMin
	o.LettersCancellation.TimeSec = ev.LetterCancellationSubTest.TimeInSecs
	o.LettersCancellation.Status = "available"

	// BVMT (maneja puntero de Score)
	o.BVMT.Status = string(ev.VisualMemorySubTest.Status)
	if ev.VisualMemorySubTest.Score != nil {
		o.BVMT.FinalScore = &ev.VisualMemorySubTest.Score.FinalScore
		o.BVMT.IoU = &ev.VisualMemorySubTest.Score.IoU
		o.BVMT.SSIM = &ev.VisualMemorySubTest.Score.SSIM
		o.BVMT.PSNR = &ev.VisualMemorySubTest.Score.PSNR
	}

	// Verbal Memory
	o.VerbalMemory.Score = ev.VerbalmemorySubTest.Score.Score
	o.VerbalMemory.Hits = ev.VerbalmemorySubTest.Score.Hits
	o.VerbalMemory.Omissions = ev.VerbalmemorySubTest.Score.Omissions
	o.VerbalMemory.Intrusions = ev.VerbalmemorySubTest.Score.Intrusions
	o.VerbalMemory.Perseverations = ev.VerbalmemorySubTest.Score.Perseverations
	o.VerbalMemory.Accuracy = ev.VerbalmemorySubTest.Score.Accuracy
	o.VerbalMemory.IntrusionRate = ev.VerbalmemorySubTest.Score.IntrusionRate
	o.VerbalMemory.PerseverationRate = ev.VerbalmemorySubTest.Score.PerseverationRate
	o.VerbalMemory.Type = string(ev.VerbalmemorySubTest.Type)

	// Executive
	o.Executive.Score = ev.ExecutiveFunctionSubTest.Score.Score
	o.Executive.Accuracy = ev.ExecutiveFunctionSubTest.Score.Accuracy
	o.Executive.SpeedIndex = ev.ExecutiveFunctionSubTest.Score.SpeedIndex
	o.Executive.CommissionRate = ev.ExecutiveFunctionSubTest.Score.CommissionRate
	o.Executive.DurationSec = ev.ExecutiveFunctionSubTest.Score.DurationSec
	o.Executive.Type = string(ev.ExecutiveFunctionSubTest.Type)

	// Fluency
	o.Fluency.Score = ev.LanguageFluencySubTest.Score.Score
	o.Fluency.UniqueValid = ev.LanguageFluencySubTest.Score.UniqueValid
	o.Fluency.Intrusions = ev.LanguageFluencySubTest.Score.Intrusions
	o.Fluency.Perseverations = ev.LanguageFluencySubTest.Score.Perseverations
	o.Fluency.TotalProduced = ev.LanguageFluencySubTest.Score.TotalProduced
	o.Fluency.WordsPerMin = ev.LanguageFluencySubTest.Score.WordsPerMinute
	o.Fluency.IntrusionRate = ev.LanguageFluencySubTest.Score.IntrusionRate
	o.Fluency.PersevRate = ev.LanguageFluencySubTest.Score.PersevRate
	o.Fluency.Category = ev.LanguageFluencySubTest.Category

	b, _ := json.Marshal(o)
	return string(b)
}

func (oa OpenAIService) LettersCancellationAnalysis(subtest *LCdomain.LettersCancellationSubtest, patientAge int) (string, error) {
	prompt := fmt.Sprintf(`Eres una IA clínica especializada en neuropsicología y trastornos del movimiento.
Contexto: paciente con enfermedad de Parkinson avanzada. Tu tarea es interpretar SOLO el subtest de Cancelación de Letras, sin diagnosticar, y proponer una hipótesis de perfil predominante basada exclusivamente en este subtest.

Datos del subtest (Cancelación de Letras):
- Edad del paciente: %d años
- Duración: %d s
- Objetivos totales (targets): %d
- Aciertos (hits): %d
- Errores (comisiones): %d
- Omisiones: %d
- Métricas derivadas:
  - score (0-100): %d
  - cpPerMin (rendimiento neto/min): %.2f
  - accuracy (H/N): %.3f
  - omissionsRate: %.3f
  - commissionRate: %.3f
  - hitsPerMin: %.2f
  - errorsPerMin: %.2f

Instrucciones clínicas:
- Considera que valores altos indican mejor desempeño.
- No inventes normas poblacionales ni uses referencias no proporcionadas.
- No generalices a otros dominios (memoria, lenguaje…) más allá de lo que sugiere este subtest atencional.
- Reconoce la limitación de inferir un perfil global desde una única prueba.
- Aun así, elige UN único "perfil neurológico predominante" como hipótesis basada SOLO en este patrón atencional:

Perfiles posibles (elige exactamente UNO):
["Amnésico","Fronto-temporal","Atencional","Depresivo","Disexecutivo (vascular)","Indeterminado"]

Criterios orientativos (no normativos):
- Patrón atencional/impulsivo: accuracy baja y/o commissionRate alta; cpPerMin bajo.
- Patrón enlentecido/apatía: hitsPerMin bajo con pocos errores; cpPerMin moderado-bajo.
- Disexecutivo: errores y variabilidad sugieren problemas de control/monitorización sostenida.
- Amnésico: NO debe inferirse desde este subtest salvo evidencia indirecta muy débil.
- Fronto-temporal/Depresivo: solo si el patrón encaja claramente (p. ej., enlentecimiento marcado vs. impulsividad), y deja constancia de la incertidumbre.

Salida:
Responde EXCLUSIVAMENTE en JSON válido con este esquema:
{
  "predominantProfile": "<uno de: Amnésico | Fronto-temporal | Atencional | Depresivo | Disexecutivo (vascular) | Indeterminado>",
  "confidence": <número entre 0 y 1>,
  "analysis": "<3-6 frases clínicas, claras y concisas, basadas en las métricas provistas. Explica por qué el perfil elegido encaja y qué limitaciones tiene inferirlo desde un único subtest.>"
}`, patientAge, subtest.TimeInSecs, subtest.TotalTargets, subtest.Correct, subtest.Errors,
		subtest.CancellationScore.Omissions, subtest.CancellationScore.Score, subtest.CancellationScore.CpPerMin,
		subtest.CancellationScore.Accuracy, subtest.CancellationScore.OmissionsRate, subtest.CancellationScore.CommissionRate,
		subtest.CancellationScore.HitsPerMin, subtest.CancellationScore.ErrorsPerMin)
	res, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (oa OpenAIService) Ask(prompt string) (string, error) {
	resp, err := oa.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4, // puedes usar gpt-4o si tu SDK lo soporta
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "Eres un experto en neuropsicología y evaluaciones clínicas en enfermedad de Parkinson. Tu tarea es generar informes diagnósticos precisos.",
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
		Temperature: 0,
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func (oa OpenAIService) VerbalMemoryAnalysis(subtest *VEMdomain.VerbalMemorySubtest, patientAge int) (string, error) {
	prompt := fmt.Sprintf(`Eres una IA clínica especializada en neuropsicología y trastornos del movimiento.
Contexto: paciente con enfermedad de Parkinson avanzada. Tu tarea es interpretar SOLO el subtest de Memoria Verbal, sin diagnosticar, y proponer una hipótesis de perfil predominante basada
exclusivamente en este subtest.
Datos del subtest (Memoria Verbal):
- Edad del paciente: %d años
- Tipo de subtest: %s
- Palabras dadas (given): %s
- Palabras recordadas (recalled): %s
- Métricas derivadas:
  - score (0-100): %v
  - hits: %d
  - omissions: %d
  - intrusions: %d
  - perseverations: %d
  - accuracy: %.3f
  - intrusionRate: %.3f
  - perseverationRate: %.3f
Instrucciones clínicas:
- Considera que valores altos indican mejor desempeño.
- No inventes normas poblacionales ni uses referencias no proporcionadas.
- No generalices a otros dominios (atención, lenguaje…) más allá de lo que sugiere este subtest de memoria.
- Reconoce la limitación de inferir un perfil global desde una única prueba.
- Aun así, elige UN único "perfil neurológico predominante" como hipótesis basada SOLO en este patrón mnemónico:
Perfiles posibles (elige exactamente UNO):
["Amnésico","Fronto-temporal","Atencional","Depresivo","Disexecutivo (vascular)","Indeterminado"]
Criterios orientativos (no normativos):
- Patrón amnésico: hits bajos y omissions altas, con intrusions/perseverations bajas.
- Patrón atencional/impulsivo: intrusions y/o perseverations elevadas.
- Patrón enlentecido/apatía: omissions moderadas con pocos hits y errores; intrusions/perseverations bajas.
- Disexecutivo: intrusions y perseverations sugieren problemas de control/monitorización.
- Fronto-temporal/Depresivo: solo si el patrón encaja claramente (p. ej., enlentecimiento marcado vs. impulsividad), y deja constancia de la incertidumbre.
`, patientAge, subtest.Type, strings.Join(subtest.GivenWords, ", "), strings.Join(subtest.RecalledWords, ", "),
		subtest.Score, subtest.Score.Hits, subtest.Score.Omissions, subtest.Score.Intrusions, subtest.Score.Perseverations,
		subtest.Score.Accuracy, subtest.Score.IntrusionRate, subtest.Score.PerseverationRate)
	res, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (oa OpenAIService) ExecutiveFunctionsAnalysis(subtest *EFdomain.ExecutiveFunctionsSubtest, patientAge int) (string, error) {
	prompt := fmt.Sprintf(`Eres una IA clínica especializada en neuropsicología y trastornos del movimiento.
Contexto: paciente con enfermedad de Parkinson u otro transtorno neurocognitivo. Tu tarea es interpretar SOLO el subtest de Funciones Ejecutivas, sin diagnosticar, y proponer una hipótesis de perfil predominante basada
exclusivamente en este subtest.
Datos del subtest (Funciones Ejecutivas):
- Edad del paciente: %d años
- Tipo de subtest: %s
- Número de ítems: %d
- Total de clics: %d
- Total de errores: %d
- Total correcto: %d
- Tiempo total (s): %.2f
- Métricas derivadas:
  - score (0-100): %d
  - accuracy: %.3f
  - speedIndex: %.3f
  - commissionRate: %.3f
  - durationSec: %.2f
Instrucciones clínicas:
- Considera que valores altos indican mejor desempeño.
- No inventes normas poblacionales ni uses referencias no proporcionadas.
- No generalices a otros dominios (memoria, lenguaje…) más allá de lo que sugiere este subtest de funciones ejecutivas.
- Reconoce la limitación de inferir un perfil global desde una única prueba.
- Aun así, elige UN único "perfil neu		rológico predominante" como hipótesis basada SOLO en este patrón ejecutivo:
Perfiles posibles (elige exactamente UNO):
["Amnésico","Fronto-temporal","Atencional","Depresivo","Disexecutivo (vascular)","Indeterminado"]
Criterios orientativos (no normativos):
- Patrón disexecutivo: accuracy baja y/o commissionRate alta; speedIndex bajo.
- Patrón enlentecido/apatía: accuracy moderada con speedIndex muy bajo; commissionRate baja.
- Patrón atencional/impulsivo: accuracy baja con commissionRate alta.
- Amnésico: NO debe inferirse desde este subtest salvo evidencia indirecta muy	 débil.
- Fronto-temporal/Depresivo: solo si el patrón encaja claramente (p. ej., enlentecimiento marcado vs. impulsividad), y deja constancia de la incertidumbre.
	`, patientAge, subtest.Type, subtest.NumberOfItems, subtest.TotalClicks, subtest.TotalErrors,
		subtest.TotalCorrect, subtest.TotalTime.Seconds(), subtest.Score.Score, subtest.Score.Accuracy,
		subtest.Score.SpeedIndex, subtest.Score.CommissionRate, subtest.Score.DurationSec)
	res, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return res, nil
}

func (oa OpenAIService) LanguageFluencyAnalysis(subtest *LFdomain.LanguageFluency, patientAge int) (string, error) {
	prompt := fmt.Sprintf(`Eres una IA clínica especializada en neuropsicología y trastornos del movimiento.
Contexto: paciente con enfermedad de Parkinson avanzada. Tu tarea es interpretar SOLO el subtest de Fluidez Verbal, sin diagnosticar, y proponer una hipótesis de perfil predominante basada
exclusivamente en este subtest.				
Datos del subtest (Fluidez Verbal):
- Edad del paciente: %d años
- Idioma: %s
- Nivel de competencia: %s
- Categoría: %s
- Palabras dadas (answerWords): %s
- Métricas derivadas:
  - score (0-100): %d
  - uniqueValid: %d
  - intrusions: %d
  - perseverations: %d
  - totalProduced: %d
  - wordsPerMinute: %.2f
  - intrusionRate: %.3f
  - persevRate: %.3f
Instrucciones clínicas:
- Considera que valores altos indican mejor desempeño.
- No inventes normas poblacionales ni uses referencias no proporcionadas.
- No generalices a otros dominios (atención, memoria…) más allá de lo que


sugiere este subtest de fluidez verbal.
- Reconoce la limitación de inferir un perfil global desde una única prueba.
- Aun así, elige UN único "perfil neurológico predominante" como hipó
tesis basada SOLO en este patrón verbal:
Perfiles posibles (elige exactamente UNO):
["Amnésico","Fronto-temporal","Atencional","Depresivo","Dis
executivo (vascular)","Indeterminado"]
Criterios orientativos (no normativos):
- Patrón amnésico: uniqueValid bajo con intrusions/perseverations bajas.
- Patrón atencional/impulsivo: intrusions y/o perseverations elevadas
- Patrón enlentecido/apatía: uniqueValid moderado-bajo con pocos errores; intrusions/perseverations bajas.
- Disexecutivo: intrusions y perseverations sugieren problemas de control/monitorización.
- Fronto-temporal/Depresivo: solo si el patrón encaja claramente (p
. ej., enlentecimiento marcado vs. impulsividad), y deja constancia de la incertidumbre.
`, patientAge, subtest.Language, subtest.Proficiency, subtest.Category, strings.Join(subtest.AnswerWords, ", "),
		subtest.Score.Score, subtest.Score.UniqueValid, subtest.Score.Intrusions, subtest.Score.Perseverations,
		subtest.Score.TotalProduced, subtest.Score.WordsPerMinute, subtest.Score.IntrusionRate, subtest.Score.PersevRate)
	res, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return res, nil
}
