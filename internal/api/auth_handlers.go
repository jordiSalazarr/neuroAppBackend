package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	signup "neuro.app.jordi/internal/auth/applicaton/commands/sign-up"
)

func (app *App) SignUp(c *gin.Context) {
	var command signup.SignUpCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	user, _, err := signup.SignUpCommandHandler(c.Request.Context(), command, app.Repositories.UserRepository, app.Services.MailService)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	msg := ""
	if user != nil {
		msg = "usuario creado correctamente, por favor verifica tu correo en " + user.Email.Mail

	}

	c.JSON(http.StatusOK, gin.H{
		"success": msg,
	})
}

func (app *App) GetUserInfo(c *gin.Context) {}
