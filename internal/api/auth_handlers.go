package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
	adminacceptsuser "neuro.app.jordi/internal/auth/applicaton/commands/admin-accepts-user"
	login "neuro.app.jordi/internal/auth/applicaton/commands/log-in"
	signup "neuro.app.jordi/internal/auth/applicaton/commands/sign-up"
	verifyuser "neuro.app.jordi/internal/auth/applicaton/commands/verify-user"
	pendigacceptrequest "neuro.app.jordi/internal/auth/applicaton/queries/pendig-accept-request"
	"neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/shared/midleware"
)

func (app *App) SignUp(c *gin.Context) {
	var command signup.SignUpCommand
	if err := c.ShouldBindJSON(&command); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	user, err := signup.SignUpCommandHandler(command, app.EncryptionService, app.UserRepository, app.MailService)
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

func (app *App) Login(c *gin.Context) {
	var command login.LoginCommand
	app.Logger.Debug(c, "Login request received")
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c, "Error parsing login command: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	user, token, err := login.LogInCommandHandler(command, app.UserRepository, app.EncryptionService)
	user.Password = domain.Password{}
	if err != nil {
		app.Logger.Error(c, "Login failed: "+err.Error())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	app.Logger.Info(c, "User logged in successfully")
	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Login successful",
		"user":    user,
	})
}

func (app *App) VerifyUser(c *gin.Context) {
	var command verifyuser.VerifyUserCommand
	app.Logger.Debug(c, "Verify user request received")
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c, "Error parsing verify user command: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	user, token, err := verifyuser.VerifyUserCommandHandler(command, app.UserRepository, *app.JwtService)
	user.Password = domain.Password{} // Remove password from response
	if err != nil {
		app.Logger.Error(c, "Verification failed: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	app.Logger.Info(c, "User verified successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": "user verified successfully",
		"token":   token,
		"user":    user,
	})
}

func (app *App) AcceptUser(c *gin.Context) {
	var command adminacceptsuser.AcceptUserCommand
	app.Logger.Debug(c, "Accept  user request received")
	if err := c.ShouldBindJSON(&command); err != nil {
		app.Logger.Error(c, "Error parsing accept user command: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{"error parsing input": err.Error()})
		return
	}
	adminID, ok := midleware.GetUserIdFromRequest(c)
	if !ok {
		app.Logger.Error(c, "Admin ID not found in request context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	command.AdminID = adminID
	app.Logger.Debug(c, "Admin ID: "+command.AdminID)
	err := adminacceptsuser.AdminAcceptsUserCommandHandler(command, app.UserRepository)
	if err != nil {
		app.Logger.Error(c, "Accept user failed: "+err.Error())
		if err == adminacceptsuser.ErrInvalidOperation {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid operation"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		}
		return
	}
	app.Logger.Info(c, "User accepted successfully")
	c.JSON(http.StatusOK, gin.H{
		"success": "user accepted successfully",
	})
}

func (app *App) GetPendingAcceptRequests(c *gin.Context) {
	var query pendigacceptrequest.PendingAcceptRequestQuery
	app.Logger.Debug(c, "Get pending accept requests query received")
	adminID, ok := midleware.GetUserIdFromRequest(c)
	if !ok {
		app.Logger.Error(c, "Admin ID not found in request context")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}
	query.AdminID = adminID
	users, err := pendigacceptrequest.PendingAcceptRequestQueryHandler(query, app.UserRepository)
	app.Logger.Debug(c, "Admin ID: "+query.AdminID)
	if err != nil {
		app.Logger.Error(c, "Error fetching pending accept requests: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal server error"})
		return
	}
	app.Logger.Info(c, "Pending accept requests fetched successfully")
	c.JSON(http.StatusOK, users)
}
