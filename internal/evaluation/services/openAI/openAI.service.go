package services

import (
	"context"
	"fmt"
	"strings"

	openai "github.com/sashabaranov/go-openai"
	"neuro.app.jordi/internal/evaluation/domain"
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
	safe := sanitizeEvaluation(ev)
	safe.PK = ""
	formattedEval := formatEvaluationForLLM(safe)

	prompt := fmt.Sprintf(
		`Eres un/a neuropsicólogo/a clínico especializado/a en enfermedad de Parkinson avanzada.
Vas a analizar una evaluación **anónima** compuesta por subtests estandarizados. 
Trabaja SOLO con los datos proporcionados (no inventes, no infieras identidad, edad o demografía) y **omite cualquier referencia personal**.

REGLAS CRÍTICAS
- Ten en cuenta la edad del paciente
- Usa únicamente subtests **con datos válidos**. Considera “sin datos” cualquier subtest con status en {pending, processing}, campos nulos, o marcado explícitamente como no evaluado.
- Diferencia entre “0 válido” y “0 ausente”:
  • Si el subtest **define 0 como resultado posible** (p.ej., Memoria Visual 0–2 por figura o CDT con Score=0) → **sí es un dato válido** (peor rendimiento).
- Interpretación de métricas:
  • Métricas de desempeño (p.ej., Score 0–100, Accuracy, SpeedIndex): ↑ = mejor.
  • Métricas de error/tasa (p.ej., intrusionsRate, perseverations, commissionRate, omissionsRate): ↑ = peor.
  • Duraciones/latencias (DurationSec, tiempos TMT): ↑ = peor (más enlentecimiento).
- Si hay discrepancias dentro de un subtest, explica posibles causas (velocidad vs precisión, fatiga, impulsividad, efecto aprendizaje, etc.).
- Señala anomalías de calidad (artefactos) y **suaviza** las conclusiones cuando afecten al resultado.

NORMALIZACIÓN Y UMBRALES (guía heurística, no diagnóstica)
- Escalas 0–100 (cuando existan): 80–100 preservado; 60–79 leve fragilidad; 40–59 leve–moderado; 0–39 moderado–severo.
- **Memoria Visual (humana 0–2 por figura):** si hay N figuras y figureScores en {0,1,2}, calcula
  VM_norm = (sum(figureScores) / (2*N)) * 100.
  • 85–100: reproducción preservada o casi completa
  • 67–84: leve fragilidad (errores puntuales de forma/orientación/tamaño)
  • 34–66: compromiso leve–moderado (reconocimiento parcial, omisiones relevantes)
  • 0–33: compromiso moderado–severo (reproducciones irreconocibles/ausentes)
  Si solo hay un agregado (score_id 0–2 por figura o totalScore0to(2N)), aplica la misma normalización.
- En subtests con varias tentativas/ensayos, considera **patrón de aprendizaje** (mejora o fatiga).

DOMINIOS Y MÉTRICAS (resumen operativo)
1) Atención sostenida — Letters Cancellation
   Métricas:  Accuracy, Omissions, CommissionRate, Hits/Errors per min, CpPerMin. SCORE DOES NOT MATTER
   Fijate sobretodo en los aciertos y errores, el score no importa mucho.
   Déficit típico: ↑omissions/commissionRate y ↓accuracy/score → perfil atencional.

2) **Memoria Visual — BVMT (versión con evaluación humana 0–2/figura)**
   Datos esperados: figureScores:[0|1|2], score_id por figura, totalScore (0–2N), note del evaluador.
   Criterios (por figura):
     • 2 = forma y orientación correctas
     • 1 = parcialmente correcta (error en tamaño/orientación, incompleta)
     • 0 = incorrecta/irreconocible
   Interpretación: Normaliza a 0–100 (VM_norm) según arriba. Si hay notas de baja calidad (“baja calidad”, “artefacto”, “iluminación”, “movimiento”), **avisar posible sesgo**.

3) Memoria Verbal — VerbalMemory
   Métricas: Score (0–100), Hits, Omissions, Intrusions, Perseverations, Accuracy, IntrusionRate, PerseverationRate.
   Déficit amnésico: ↓score/accuracy con ↑omissions; intrusions/perseverations sugieren fallo en control/monitorización.

4) Funciones ejecutivas — ExecutiveFunctions (p.ej., TMT A/B)
   Métricas: Score (0–100), Accuracy, SpeedIndex, CommissionRate, DurationSec.
   Déficit ejecutivo/velocidad: ↓score/accuracy/speedIndex con ↑commissionRate/tiempos.

5) Fluencia verbal — LanguageFluency (p.ej., semántica)
   Métricas: Score (0–100), UniqueValid, Intrusions, Perseverations, WordsPerMinute, IntrusionRate, PersevRate.
   Déficit léxico/ejecutivo: ↓uniqueValid/WPM, ↑intrusions/perseverations.

6) Visuoespacial / Construcción — Clock Drawing Test (CDT)
   Métricas: Score(0-5) con la escala de Shulman 5 es lo mejor

REGLAS DE PONDERACIÓN
- Prioriza conclusiones donde **varias métricas del mismo dominio** convergen (consistencia interna).
- Si dominios difieren, explica la **coherencia inter-dominios** (p.ej., atención baja puede arrastrar ejecución/fluencia).
- Declara **incertidumbre** cuando los datos sean escasos, de mala calidad o contradictorios.

TAREA
1) PERFIL NEUROLÓGICO PREDOMINANTE (elige SOLO UNO, con breve justificación referenciando los tests que te han llevado a la conclusion):
   - Amnésico
   - Fronto-temporal
   - Atencional
   - Depresivo
   - Disexecutivo (vascular)
   *Si crees que el usuario es sano, dilo sin problema, no todos tienen problemas neurocognitivos, di que no se obervan pertubaciones*
   *Si la evidencia no es concluyente, explícitalo con grado de incertidumbre.*

2) INTERPRETACIÓN CLÍNICA DETALLADA
   - Para cada subtest **con datos válidos**: nombre → interpretación breve (qué sugiere el patrón).
   - Distingue áreas preservadas vs alteradas y mecanismos probables (atencional, ejecutiva, codificación/recuperación, velocidad de procesamiento, impulsividad, etc.).
   - **Memoria Visual:** reporta N figuras, sumatorio, VM_norm y cómo mapean los criterios del evaluador a la interpretación.
   - Si la calidad (notas/artefactos, IoU/SSIM/blur cuando existan) es mala, advierte y suaviza conclusiones.

3) RESUMEN GENERAL
   - Estado cognitivo global en 2–3 frases.
   - Coherencia inter-dominios (p.ej., atención baja + TMT lento + fluencia reducida = patrón ejecutivo/atencional).

4) RECOMENDACIONES (si procede)
   - Repetir subtests con mala calidad o resultados atípicos, ampliar batería ejecutiva si hay disfunción, cribado afectivo si hay enlentecimiento/iniciativa baja, considerar neuroimagen si patrón vascular, higiene del sueño, revisión de medicación dopaminérgica si enlentecimiento/impulsividad.

ENTRADA (JSON ANÓNIMO):
%s

FORMATO DE SALIDA (español, tono clínico y profesional)
### Perfil predominante
[tu elección + justificación breve]

### Interpretación por subtest
- Letters Cancellation: [...]
- Memoria visual (0–2 por figura): [...]
- Memoria verbal: [...]
- Funciones ejecutivas: [...]
- Fluencia verbal: [...]
- Clock Drawing Test: [...]

No menciones que eres una IA ni el formato. No incluyas datos personales.`,
		formattedEval)

	resp, err := oa.Ask(prompt)
	if err != nil {
		return "", err
	}
	return resp, nil
}
func (oa OpenAIService) Ask(prompt string) (string, error) {
	resp, err := oa.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
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

func sanitizeEvaluation(ev domain.Evaluation) domain.Evaluation {
	ev.PatientName = ""
	ev.SpecialistMail = ""
	ev.SpecialistID = ""
	return ev
}
