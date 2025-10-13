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

type MockOpenAIService struct{}

func NewMockOpenAIService() MockOpenAIService {
	return MockOpenAIService{}
}

func (m MockOpenAIService) GenerateAnalysis(ev domain.Evaluation) (string, error) {
	return "Mocked analysis", nil
}

func (m MockOpenAIService) Ask(prompt string) (string, error) {
	return "Mocked response", nil
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

func (oa OpenAIService) GenerateAnalysis(ev domain.Evaluation) (string, error) {
	safe := sanitizeEvaluation(ev)
	safe.PK = ""
	formattedEval := formatEvaluationForLLM(safe)

	prompt := fmt.Sprintf(
		`Eres un/a neuropsicólogo/a clínico especializado/a en enfermedad de Parkinson avanzada.
Vas a analizar una evaluación **anónima** compuesta por subtests estandarizados.
Trabaja SOLO con los datos proporcionados (no inventes, no infieras identidad, edad o demografía) y **omite cualquier referencia personal**.

REGLAS CRÍTICAS
- Ten en cuenta la edad del paciente cuando esté disponible en la entrada.
- Usa únicamente subtests **con datos válidos**. Considera “sin datos” cualquier subtest con status en {pending, processing}, campos nulos, vacíos o marcados como no evaluados.
- Distingue “0 válido” vs “0 ausente”:
  • Si el subtest **acepta 0 como resultado posible** (p.ej., BVMT 0–2 por figura o CDT con Score=0) → **trátalo como dato válido** (peor rendimiento), NO como ausencia.
- Interpretación de métricas (signo):
  • Desempeño (Score 0–100, Accuracy, SpeedIndex): ↑ = mejor.
  • Errores/Tasas (IntrusionRate, PerseverationRate, CommissionRate, OmissionsRate): ↑ = peor.
  • Tiempo/Latencias (DurationSec, TMT): ↑ = peor (enlentecimiento).
- Si hay discrepancias internas, explica posibles causas (velocidad vs precisión, fatiga, impulsividad, efecto aprendizaje, fluctuaciones dopaminérgicas).
- Señala artefactos/alertas de calidad (nota del evaluador, blur/IoU/SSIM cuando existan) y **suaviza** conclusiones si afectan el resultado.

NORMALIZACIÓN Y UMBRALES (guía clínica no diagnóstica)
- Escalas 0–100: 80–100 preservado; 60–79 fragilidad leve; 40–59 leve–moderado; 0–39 moderado–severo.
- **Memoria Visual — BVMT (0–2 por figura):**
  Si hay N figuras con figureScores∈{0,1,2}:
    VM_norm = (sum(figureScores) / (2*N)) * 100
    • 85–100: reproducción preservada/casi completa
    • 67–84: fragilidad leve
    • 34–66: compromiso leve–moderado
    • 0–33: compromiso moderado–severo
  Si solo hay agregado (totalScore 0..2N), aplica la misma normalización.
- Subtests con múltiples ensayos: evalúa **patrón de aprendizaje** (mejora o fatiga).

DOMINIOS Y MÉTRICAS (resumen operativo)
1) **Atención sostenida — Letters Cancellation**
   Métricas: Accuracy, Omissions, CommissionRate, HitsPerMin, ErrorsPerMin, CpPerMin. **El “score” global importa menos.**
   Enfócate en aciertos/errores (omisiones/comisiones) y en el equilibrio velocidad-precisión.

2) **Memoria Visual — BVMT (evaluación humana 0–2/figura)**
   Datos esperados: figureScores, totalScore (0–2N), notas del evaluador.
   Criterios por figura:
     • 2 = forma y orientación correctas.
     • 1 = parcialmente correcta (tamaño/orientación, incompleta).
     • 0 = incorrecta/irreconocible.
   Normaliza a 0–100 (VM_norm) y **reporta N, sumatorio y VM_norm**.
   Si hay notas de baja calidad (“baja calidad”, “artefacto”, “iluminación”, “movimiento”), **advierte posible sesgo**.

3) **Memoria Verbal — Inmediata y Diferida**
   Estructura de entrada esperada (si existe): verbal_memory.immediate y verbal_memory.delayed, cada uno con:
   Score(0–100), Hits, Omissions, Intrusions, Perseverations, Accuracy, IntrusionRate, PerseverationRate.
   Reglas interpretativas:
   - **Baja Inmediata + Baja Diferida en proporción similar** → problema de **codificación/atención** (posible arrastre por atención/velocidad).
   - **Inmediata preservada/aceptable + Diferida baja** → **déficit de consolidación/recuperación** (fragilidad mnésica genuina).
   - **Intrusiones/Perseveraciones elevadas** → **fallo de monitorización/ control ejecutivo**.
   - Considera **patrón de aprendizaje** entre ensayos si está disponible.
   Si solo hay una de las dos (inmediata o diferida), **indícalo** y limita la inferencia.

4) **Funciones ejecutivas — TMT (A y A+B)**
   Métricas: duraciones (seg).
   Umbrales orientativos: **A < 100 s** normal; **A+B < 350 s** normal (si superan → enlentecimiento/ set-shifting comprometido).
   Pautas:
   - A normal y A+B lento → déficit de **set-shifting** (componente ejecutivo).
   - A lento ya sugiere **velocidad de procesamiento**/atención comprometida (Parkinson: confundir con bradicinesia).
   - Si hay muchos errores/correcciones (si están disponibles), indícalo.

5) **Fluencia verbal — (p.ej., Semántica)**
  En este test lo mas importante es la cantidad de palabras correctas producidas, así que menciónalo si o si.
   Métricas: Score(0–100), UniqueValid, WordsPerMinute, IntrusionRate, PerseverationRate.
   Déficit léxico/ejecutivo: ↓UniqueValid/WPM, ↑Intrusions/Perseverations.

6) **Visuoespacial / Construcción — Clock Drawing Test (CDT, Shulman 0–5)**
   5 = mejor. Puntajes bajos → alteración visuoespacial/ejecutiva; revisa notas del evaluador si existen.

PONDERACIÓN Y COHERENCIA
- Prioriza conclusiones donde **varias métricas dentro del mismo dominio** convergen (consistencia interna).
- Si hay desacuerdos entre dominios, explica **coherencia inter-dominios** (p.ej., atención baja + TMT lento + fluencia reducida → patrón ejecutivo/atencional).
- Declara **grado de incertidumbre** cuando los datos sean escasos/contradictorios o de mala calidad.
- En Parkinson, considera el **arrastre motor** (bradicinesia) sobre tareas cronometradas; no lo confundas con déficit cognitivo puro si otras métricas no lo respaldan.

TAREA
1) **Perfil neurológico predominante (elige SOLO UNO y justifícalo brevemente citando subtests y métricas clave):**
   - Amnésico
   - Fronto-temporal
   - Atencional
   - Depresivo
   - Disexecutivo (vascular)
   *Si el perfil es compatible con funcionamiento normal, dilo claramente (no todos presentan alteraciones).*
   *Si la evidencia no es concluyente, indícalo y explica por qué.*

2) **Interpretación clínica detallada por dominio**
   - Para cada subtest **con datos válidos**: nómbralo y resume el patrón (preservado/fragilidad/alteración) y los mecanismos probables (atencional, ejecutiva, codificación/recuperación, velocidad de procesamiento, impulsividad).
   - **Memoria Visual (BVMT):** reporta N figuras, sumatorio, VM_norm y cómo mapean las notas del evaluador a la interpretación.
   - **Memoria Verbal:** separa **Inmediata** y **Diferida**; contrasta codificación vs consolidación/recuperación; comenta intrusiones/perseveraciones si son relevantes.
   - Si la **calidad** es mala (artefactos, notas), advierte y **suaviza** conclusiones.

3) **Resumen general**
   - Estado cognitivo global en 2–3 frases.
   - Coherencia inter-dominios (p.ej., atención baja + TMT lento + fluencia reducida = patrón ejecutivo/atencional).

4) **Recomendaciones (si procede)**
   - Repetir subtests con mala calidad o resultados atípicos.
   - Ampliar batería ejecutiva si hay disfunción; cribado afectivo si enlentecimiento/ apatía; considerar neuroimagen si patrón vascular.
   - Higiene del sueño; revisar **medicación dopaminérgica** si hay impulsividad/enlentecimiento que pueda sesgar.

ENTRADA (JSON ANÓNIMO):
%s

FORMATO DE SALIDA (español, tono clínico y profesional). **Usa SIEMPRE Markdown con títulos y subtítulos en negrita**:

**Título del informe** (una línea con el hallazgo global)

**Perfil predominante**
[tu elección + justificación breve con referencias a subtests]

**Interpretación por subtest**
- **Letters Cancellation:** [...]
- **Memoria visual (BVMT 0–2/figura):** [...]
- **Memoria verbal — Inmediata:** [...]
- **Memoria verbal — Diferida:** [...]
- **Funciones ejecutivas (TMT A / A+B):** [...]
- **Fluencia verbal:** [...]
- **Clock Drawing Test (CDT):** [...]

**Resumen general**
[2–3 frases sobre el estado global y coherencia inter-dominios]

**Recomendaciones**
[Puntos accionables breves]

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
		Model: openai.GPT4Dot1,
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
