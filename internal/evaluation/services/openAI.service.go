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
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Fecha de Evaluación: %s\n", eval.CreatedAt.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("Paciente: %s\n", eval.PatientName))
	b.WriteString(fmt.Sprintf("Puntuación Total: %d/100\n", eval.TotalScore))
	b.WriteString("———\n\n")

	for _, section := range eval.Sections {
		if section.Score > 0 {
			b.WriteString(fmt.Sprintf("▶️ Dominio: %s (Puntuación: %d/100)\n", section.Name, section.Score))
			for _, q := range section.Questions {
				b.WriteString(fmt.Sprintf(" - Pregunta: %s\n", q.Answer))
				b.WriteString(fmt.Sprintf("   ➤ Respuesta: %s\n", q.Response))
				if q.Correct != "" {
					b.WriteString(fmt.Sprintf("   ✔️ Correcta: %s\n", q.Correct))
				}
				b.WriteString(fmt.Sprintf("   🟢 Puntos: %d\n", q.Score))
			}
			b.WriteString("\n")
		}
	}
	return b.String()
}
