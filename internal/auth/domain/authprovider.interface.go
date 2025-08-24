package domain

type AuthProvider interface {
	Login(username, password string) (string, error)
	SignUp(username, password string) (string, error)
	ResetPassword(username, newPassword string) error
}
