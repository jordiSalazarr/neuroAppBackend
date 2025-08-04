package adminacceptsuser

import (
	"errors"
	"time"

	"neuro.app.jordi/internal/auth/domain"
)

var ErrInvalidOperation = errors.New("the operation is not valid for the current user state")

func AdminAcceptsUserCommandHandler(command AcceptUserCommand, userRepository domain.UserRepository) error {
	user, err := userRepository.GetUserById(command.UserID)
	if err != nil {
		return err
	}
	// admin, err := userRepository.GetUserById(command.AdminID)
	// if err != nil {
	// 	return err
	// }
	// if !admin.IsAdmin || user.IsAcceptedByAdmin || !user.IsVerified {
	// 	return ErrInvalidOperation
	// }

	user.IsAcceptedByAdmin = true
	user.UpdatedAt = time.Now()
	if err := userRepository.Update(user); err != nil {
		return err
	}
	return nil
}
