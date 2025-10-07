package speechtotext

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	"neuro.app.jordi/internal/shared/config"
)

type OpenAISpeechToText struct {
	APIKey     string
	Model      string
	HTTPClient *http.Client
}

// NewOpenAISpeechToText crea una instancia configurada
func NewOpenAISpeechToText() *OpenAISpeechToText {
	apiKey := config.GetConfig().OpenAIKey
	if apiKey == "" {
		return &OpenAISpeechToText{}
	}
	return &OpenAISpeechToText{
		APIKey: apiKey,
		Model:  "gpt-4o-mini-transcribe", // o "whisper-1" si prefieres
		HTTPClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ---------------------------------------------------------------------------
// Implementación del método
func (s *OpenAISpeechToText) GetTextFromSpeech(audio []byte) (string, error) {
	if s.APIKey == "" {
		return "", fmt.Errorf("missing OpenAI API key")
	}
	if len(audio) == 0 {
		return "", fmt.Errorf("empty audio input")
	}

	// Prepara request multipart/form-data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Campo 'file' (requerido por la API)
	part, err := writer.CreateFormFile("file", "audio.webm")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(part, bytes.NewReader(audio)); err != nil {
		return "", fmt.Errorf("copy audio: %w", err)
	}

	// Campo 'model'
	if err := writer.WriteField("model", s.Model); err != nil {
		return "", fmt.Errorf("write model field: %w", err)
	}
	// Campo 'language' opcional: fuerza español
	if err := writer.WriteField("language", "es"); err != nil {
		return "", fmt.Errorf("write language field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequest(
		http.MethodPost,
		"https://api.openai.com/v1/audio/transcriptions",
		&body,
	)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := s.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("openai request error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("openai error: %s - %s", resp.Status, string(b))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Text, nil
}
