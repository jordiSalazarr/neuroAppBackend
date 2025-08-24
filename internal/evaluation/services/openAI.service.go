package services

import (
	"context"
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
func (oa OpenAIService) GenerateAnalysis(evaluation domain.Evaluation) (string, error) {
	formattedEval := formatEvaluationAsText(evaluation)

	prompt := fmt.Sprintf(`Eres una IA médica especializada en evaluación neuropsicológica y trastornos del movimiento. 
Un especialista está valorando a un paciente con enfermedad de Parkinson avanzada. La evaluación incluye pruebas cognitivas, motoras y cuestionarios clínicos.

❗Importante:
- Analiza únicamente los dominios y subtests con puntuación > 0.
- Los valores son porcentajes sobre 100 (valores altos = mejor desempeño).
- El análisis debe ser clínico, profesional y orientado a interpretación neurológica.

### Debes proporcionar:

1. **Perfil Neurológico Predominante:** Indica si el patrón se ajusta a uno de los siguientes perfiles (elige solo uno y explica por qué):
   - **Amnésico**  
   - **Fronto-temporal**  
   - **Atencional**  
   - **Depresivo**  
   - **Disexecutivo (vascular)**  

2. **Interpretación Clínica Detallada:** Analiza los resultados de los subtests con puntuación > 0, explicando qué áreas están preservadas y cuáles muestran alteración.

3. **Resumen General:** Estado cognitivo, motor y funcional del paciente.

4. **Recomendaciones:** Si procede, sugiere seguimiento, ajustes terapéuticos o pruebas complementarias.

Informe de Evaluación:
%s

Responde **en español**, con un tono clínico, neutro y profesional. 
No menciones que eres una IA ni hagas comentarios sobre el formato.`, formattedEval)

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
		Temperature: 0.2, // baja temperatura para respuestas más consistentes
	})

	if err != nil {
		return "", fmt.Errorf("OpenAI request failed: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no response from OpenAI")
	}

	return strings.TrimSpace(resp.Choices[0].Message.Content), nil
}

func formatEvaluationAsText(eval domain.Evaluation) string {
	// var b strings.Builder

	// b.WriteString(fmt.Sprintf("Fecha de Evaluación: %s\n", eval.CreatedAt.Format("2006-01-02 15:04")))
	// b.WriteString(fmt.Sprintf("Paciente: %s\n", eval.PatientName))
	// b.WriteString(fmt.Sprintf("Puntuación Total: %d/100\n", eval.TotalScore))
	// b.WriteString("———\n\n")

	// for _, section := range eval.Sections {
	// 	if section.Score > 0 {
	// 		b.WriteString(fmt.Sprintf("▶️ Dominio: %s (Puntuación: %d/100)\n", section.Name, section.Score))
	// 		for _, q := range section.Questions {
	// 			b.WriteString(fmt.Sprintf(" - Pregunta: %s\n", q.Answer))
	// 			b.WriteString(fmt.Sprintf("   ➤ Respuesta: %s\n", q.Response))
	// 			if q.Correct != "" {
	// 				b.WriteString(fmt.Sprintf("   ✔️ Correcta: %s\n", q.Correct))
	// 			}
	// 			b.WriteString(fmt.Sprintf("   🟢 Puntos: %d\n", q.Score))
	// 		}
	// 		b.WriteString("\n")
	// 	}
	// }
	return "b.String()"
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
		Temperature: 0.2, // baja temperatura para respuestas más consistentes
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
