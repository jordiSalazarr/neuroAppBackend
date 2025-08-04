package signup

import (
	"errors"

	"neuro.app.jordi/internal/auth/domain"
	"neuro.app.jordi/internal/shared/mail"
)

var ErrInvalidCredentials = errors.New("error creating user")

func SignUpCommandHandler(command SignUpCommand, es domain.EncryptionService, userRepo domain.UserRepository, mailService mail.MailProvider) (*domain.User, string, error) {
	user, err := domain.NewUser(command.Mail, command.Password, command.Name, es)
	if err != nil {
		return nil, "", err
	}
	if exists := userRepo.Exists(user.Email.Mail); exists {
		return nil, "", ErrInvalidCredentials
	}
	user.GenerateVerificationCode()

	err = userRepo.Insert(*user)
	if err != nil {
		return nil, "", err
	}

	err = mailService.SendEmail(user.Email.Mail, "Welcome to NeuroApp", "Please verify your email address by clicking the link below.\n\nVerification Code: "+user.VerificationCode, "", nil)
	return nil, "", err
}
