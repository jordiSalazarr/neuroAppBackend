package services

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/joho/godotenv"
	openai "github.com/sashabaranov/go-openai"
	"neuro.app.jordi/internal/evaluation/domain"
)

type OpenAIService struct {
	client *openai.Client
	ApiKey string
}

func NewOpenAIService() OpenAIService {
	cwd, _ := os.Getwd()
	fmt.Println("Current working directory:", cwd)
	err := godotenv.Load(".env.local")
	if err != nil {
		fmt.Println(err.Error())
	}
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		panic("OPENAI_API_KEY is not set")
	}

	return OpenAIService{
		client: openai.NewClient(apiKey),
		ApiKey: apiKey,
	}
}

func (oa OpenAIService) GenerateAnalysis(evaluation domain.Evaluation) (string, error) {
	formattedEval := formatEvaluationAsText(evaluation)

	prompt := fmt.Sprintf(`Eres una IA médica especializada en evaluación neuropsicológica. Un especialista está valorando a un paciente con enfermedad de Parkinson avanzada. La evaluación incluye varios subtests cognitivos y motores, además de cuestionarios completados por el paciente y el médico.

❗Importante: algunos subtests pueden tener puntuación 0. Eso significa que no fueron seleccionados en esta evaluación y **no deben ser mencionados en el análisis**.

Esta evaluación corresponde a una valoración basal (antes o después de tratamiento con apomorfina, Duodopa o cirugía DBS). El análisis debe ser claro, clínicamente relevante y profesional.

Los dominios evaluados pueden incluir (si fueron seleccionados):
- Memoria  
- Atención  
- Función motora  
- Orientación espacial  
- MDS-UPDRS  
- PDQ-8  
- Escalas de satisfacción

Por favor, proporciona:

1. Una interpretación clínica concisa de los resultados obtenidos (solo los subtests con puntuación > 0).  
2. Un resumen general del estado cognitivo y motor del paciente.  
3. Sugerencias de seguimiento, si corresponde.

Evalúa únicamente la información disponible en el informe, sin hacer suposiciones sobre datos no incluidos.

Ten en cuenta que los numeros que te llegan son porcentajes sobre 100, un porcentaje más alto significa que el paciente lo ha hecho bien en esa área.

No hagas comentarios sobre lo que se ha incluido en el analisis, limitate a responder.

Informe de evaluación:  
%s

Responde en español, con un tono clínico, neutro y profesional.`, formattedEval)

	resp, err := oa.client.CreateChatCompletion(context.Background(), openai.ChatCompletionRequest{
		Model: openai.GPT4,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: "You are a neuropsychological assessment expert generating clinical reports for Parkinson’s patients.",
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

func formatEvaluationAsText(eval domain.Evaluation) string {
	var b strings.Builder

	b.WriteString(fmt.Sprintf("Evaluation Date: %s\n", eval.CreatedAt.Format("2006-01-02 15:04")))
	b.WriteString(fmt.Sprintf("Patient Name: %s\n", eval.PatientName))
	b.WriteString(fmt.Sprintf("Patient Email: %s\n", eval.SpecialistMail))
	b.WriteString(fmt.Sprintf("Total Score: %d\n", eval.TotalScore))
	b.WriteString("———\n\n")
	return b.String()
}
