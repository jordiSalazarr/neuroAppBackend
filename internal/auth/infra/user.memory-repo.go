package infra

import (
	"errors"

	"neuro.app.jordi/internal/auth/domain"
)

var ErrUserNotFound = errors.New("user not found")
var ErrUserAlreadyExists = errors.New("user already exists")

type UserInMemory struct {
	Users map[string]*domain.User
}

// NewUserInMemory crea una nueva instancia de UserInMemory
func NewUserInMemory() *UserInMemory {
	return &UserInMemory{
		Users: make(map[string]*domain.User),
	}
}

// GetUserById obtiene un usuario por su ID
func (repo *UserInMemory) GetUserById(id string) (domain.User, error) {
	user, exists := repo.Users[id]
	if !exists {
		return domain.User{}, ErrUserNotFound
	}
	return *user, nil
}

// GetUserByMail obtiene un usuario por su correo electrónico
func (repo *UserInMemory) GetUserByMail(mail string) (domain.User, error) {
	for _, user := range repo.Users {
		if user.Email.Mail == mail {
			return *user, nil
		}
	}
	return domain.User{}, ErrUserNotFound
}

// Insert inserta un nuevo usuario
func (repo *UserInMemory) Insert(user domain.User) error {
	if repo.Exists(user.Email.Mail) {
		return ErrUserAlreadyExists
	}
	repo.Users[user.ID] = &user
	return nil
}

// Delete elimina un usuario por su ID
func (repo *UserInMemory) Delete(id string) error {
	if _, exists := repo.Users[id]; !exists {
		return ErrUserNotFound
	}
	delete(repo.Users, id)
	return nil
}

// Exists verifica si un usuario existe por su correo electrónico
func (repo *UserInMemory) Exists(mail string) bool {
	for _, user := range repo.Users {
		if user.Email.Mail == mail {
			return true
		}
	}
	return false
}

// Update actualiza un usuario existente
func (repo *UserInMemory) Update(user domain.User) error {
	if _, exists := repo.Users[user.ID]; !exists {
		return ErrUserNotFound
	}
	repo.Users[user.ID] = &user
	return nil
}

func (repo *UserInMemory) UsersPendingAcceptRequest() ([]*domain.User, error) {
	var pendingUsers []*domain.User
	for _, user := range repo.Users {
		if user.IsVerified && !user.IsAcceptedByAdmin {
			pendingUsers = append(pendingUsers, user)
		}
	}
	return pendingUsers, nil
}
