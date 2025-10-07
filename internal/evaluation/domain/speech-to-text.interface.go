package domain

type SpeechToTextService interface {
	GetTextFromSpeech(audio []byte) (string, error)
}

type MockSpeechToText struct{}

func (mock *MockSpeechToText) GetTextFromSpeech(audio []byte) (string, error) {
	return "", nil
}
