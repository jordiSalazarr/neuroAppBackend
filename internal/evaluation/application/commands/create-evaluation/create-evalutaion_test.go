package createevaluation

import (
	"context"
	"testing"

	"neuro.app.jordi/internal/pkg"
)

// for each test, define the commands and the output in a table format
var mockCommands = []CreateEvaluationCommand{
	{PatientName: "John Doe", SpecialistMail: "john.doe@example.com", PatientAge: 30, SpecialistID: "spec121"},
	{PatientName: "", SpecialistMail: "jane.doe@example.com", PatientAge: 25, SpecialistID: "spec123"},
	{PatientName: "Alice", SpecialistMail: "alice@example.com", PatientAge: 0, SpecialistID: "spec1233"},
}

func TestCreateEvaluationCommandHandler(t *testing.T) {
	app := pkg.NewMockApp()
	tests := []struct {
		name       string
		command    CreateEvaluationCommand
		shouldPass bool
	}{
		{name: "Valid command", command: mockCommands[0], shouldPass: true},
		{name: "Invalid name", command: mockCommands[1], shouldPass: false},
		{name: "Invalid age", command: mockCommands[2], shouldPass: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CreateEvaluationCommandHandler(test.command, context.TODO(), app.Repositories.EvaluationsRepository)
			if (err == nil) != test.shouldPass {
				t.Errorf("Expected command to pass: %v, got error: %v", test.shouldPass, err)
			}
		})
	}
}
