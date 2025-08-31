package domain

import "context"

type UserRepository interface {
	GetUserById(ctx context.Context, id string) (User, error)
	GetUserByMail(ctx context.Context, mail string) (User, error)
	Insert(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, mail string) bool
	Update(ctx context.Context, user User) error
}
