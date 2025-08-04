package createevaluation

// import (
// 	"context"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// 	"neuro.app.jordi/internal/evaluation/domain"
// 	"neuro.app.jordi/internal/shared/mail"
// )

// func ptrFromInt(i int) *int {
// 	return &i
// }

// var mockCommands = []CreateEvaluationCommand{{
// 	TotalScore:       100,
// 	MemoryScore:      ptrFromInt(90),
// 	AtentionScore:    ptrFromInt(20),
// 	MotoreScore:      ptrFromInt(30),
// 	SpatialViewScore: ptrFromInt(0),
// 	PatientName:      "John Doe",
// 	SpecialistMail:   "wharecer@gmail.com",
// 	SpecialistID:     "specialist-023",
// }, {
// 	TotalScore:       100,
// 	MemoryScore:      ptrFromInt(90),
// 	AtentionScore:    ptrFromInt(20),
// 	MotoreScore:      ptrFromInt(30),
// 	SpatialViewScore: ptrFromInt(1),
// 	PatientName:      "John Doe",
// 	SpecialistMail:   "brrrr",
// 	SpecialistID:     "specialist-123",
// },
// 	{
// 		TotalScore:       100,
// 		MemoryScore:      ptrFromInt(90),
// 		AtentionScore:    ptrFromInt(20),
// 		MotoreScore:      ptrFromInt(30),
// 		SpatialViewScore: ptrFromInt(1),
// 		PatientName:      "John DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn DoeJohn Doe",
// 		SpecialistMail:   "brrrr",
// 		SpecialistID:     "specialist-123",
// 	},
// }
// var llmService = domain.MockInterface{}
// var fileFormatterService = domain.MockFileFormatterService{}
// var evaluationsRepository = domain.MockEvaluationsRepository{}
// var mailService = mail.NewMailer()

// func TestCreateEvaluationCommandHandler(t *testing.T) {

// 	tests := []struct {
// 		command    CreateEvaluationCommand
// 		shouldPass bool
// 		hasPassed  bool
// 	}{
// 		{command: mockCommands[0], shouldPass: true, hasPassed: true},
// 		{command: mockCommands[1], shouldPass: false, hasPassed: true},
// 		{command: mockCommands[2], shouldPass: false, hasPassed: true},
// 	}

// 	for _, test := range tests {
// 		err := CreateEvaluationCommandHandler(test.command, context.TODO(), llmService, fileFormatterService, evaluationsRepository, mailService)
// 		if err != nil {
// 			test.hasPassed = false
// 		}
// 		assert.Equal(t, test.shouldPass, test.hasPassed, "Expected command to pass or fail based on input: %v", test.command)
// 	}
// }
