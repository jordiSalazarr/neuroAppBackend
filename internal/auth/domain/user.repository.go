package domain

type UserRepository interface {
	GetUserById(id string) (User, error)
	GetUserByMail(mail string) (User, error)
	Insert(User) error
	Delete(id string) error
	Exists(mail string) bool
	Update(user User) error
	UsersPendingAcceptRequest() ([]*User, error)
}
