package verifyuser

import (
	"errors"
	"time"

	"neuro.app.jordi/internal/auth/domain"
	jwtService "neuro.app.jordi/internal/shared/jwt"
)

var ErrInvalidVerificationCode = errors.New("invalid verification code")

func VerifyUserCommandHandler(command VerifyUserCommand, userRepo domain.UserRepository, es jwtService.Service) (*domain.User, string, error) {
	user, err := userRepo.GetUserByMail(command.Email)
	if err != nil {
		return nil, "", err
	}

	if user.Email.Mail == "" || user.VerificationCode != command.Code || user.VerificationCodeExpiresAt.Before(time.Now()) {
		return nil, "", ErrInvalidVerificationCode
	}
	user.IsVerified = true
	user.VerificationCode = ""
	user.VerificationCodeExpiresAt = time.Now()
	err = userRepo.Update(user)
	if err != nil {
		return nil, "", err
	}

	token, err := es.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}
	user.Password = domain.Password{} // Remove password from response
	return &user, token, nil
}
