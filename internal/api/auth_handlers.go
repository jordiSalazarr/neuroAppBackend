package api

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	signup "neuro.app.jordi/internal/auth/applicaton/commands/sign-up"
	getbymail "neuro.app.jordi/internal/auth/applicaton/queries/getByMail"
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
	if user.ID != "" {
		msg = "usuario creado correctamente, por favor verifica tu correo en " + user.Email.Mail

	}

	c.JSON(http.StatusOK, gin.H{
		"success": msg,
	})
}

func (app *App) RegisterUserInfo(c *gin.Context) {
	mail := c.Params.ByName("mail")
	name := c.Params.ByName("name")
	if mail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"invalid mail": "no mail found"})
		return
	}
	query := getbymail.GetUserByMailQuery{
		Mail: mail,
	}
	user, err := getbymail.GetUserByMailQueryHandler(c.Request.Context(), query, app.Repositories.UserRepository)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			command := signup.SignUpCommand{
				Mail: mail,
				Name: name,
			}
			user, _, err = signup.SignUpCommandHandler(c.Request.Context(), command, app.Repositories.UserRepository, app.Services.MailService)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error parsing input": "error getting user"})
				return
			}
			c.JSON(http.StatusOK, gin.H{
				"user": user,
			})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error parsing input": "error getting user"})
			return

		}
	}

	c.JSON(http.StatusOK, gin.H{
		"user": user,
	})

}
