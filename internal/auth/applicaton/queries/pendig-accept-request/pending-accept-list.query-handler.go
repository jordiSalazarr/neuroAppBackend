package pendigacceptrequest

import "neuro.app.jordi/internal/auth/domain"

func PendingAcceptRequestQueryHandler(query PendingAcceptRequestQuery, userRepository domain.UserRepository) ([]*domain.User, error) {
	_, _ = userRepository.GetUserByMail(query.AdminID)
	// if err != nil || (user.ID != "" && !user.IsAdmin) {
	// 	return nil, err
	// }
	users, err := userRepository.UsersPendingAcceptRequest()
	if err != nil {
		return nil, err
	}
	return users, nil
}
