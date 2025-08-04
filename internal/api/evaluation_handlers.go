package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	createevaluation "neuro.app.jordi/internal/evaluation/application/commands/create-evaluation"
)

func (app *App) CreateEvaluation(c *gin.Context) {
	var command createevaluation.CreateEvaluationCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}

	err := createevaluation.CreateEvaluationCommandHandler(command, c, app.LLMService, app.FileFormater, app.EvaluationsRepository, app.MailService)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	// Return a JSON response
	c.JSON(http.StatusOK, gin.H{"successt": "check new file .pdf"})
}
