package getbymail

import (
	"context"

	"neuro.app.jordi/internal/auth/domain"
)

func GetUserByMailQueryHandler(ctx context.Context, query GetUserByMailQuery, usersRepository domain.UserRepository) (domain.User, error) {
	user, err := usersRepository.GetUserByMail(ctx, query.Mail)
	if err != nil {
		return domain.User{}, err
	}
	return user, nil
}
