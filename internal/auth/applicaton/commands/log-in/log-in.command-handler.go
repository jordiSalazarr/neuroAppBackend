package login

import (
	"errors"

	"neuro.app.jordi/internal/auth/domain"
)

var ErrSignInUser = errors.New("couldn't sign in")
var ErrInvalidCredentials = errors.New("invalid user credentials")

func LogInCommandHandler(command LoginCommand, userRepo domain.UserRepository, es domain.EncryptionService) (*domain.User, string, error) {
	user, err := userRepo.GetUserByMail(command.Mail)
	if err != nil {
		return nil, "", ErrSignInUser
	}
	inputPassword := domain.Password{
		Plain: command.Password,
	}
	inputPassword.Hash(es)
	if user.Password.Hashed != inputPassword.Hashed || !user.IsVerified {
		return nil, "", ErrInvalidCredentials
	}

	token := es.GenerateTokenFromID(user.ID)
	return &user, token, nil
}
