package domain

import (
	"errors"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
var ErrInvalidMail = errors.New("invalid mail, please try again")

type UserMail struct {
	Mail string
}

type UserName struct {
	Name string
}

func NewUserName(username string) (UserName, error) {
	if len(username) > 20 || len(username) <= 3 {
		return UserName{}, ErrUsernameInvalid
	}

	return UserName{
		Name: username,
	}, nil
}

func NewMail(mail string) (UserMail, error) {
	re := regexp.MustCompile(emailRegex)
	if ok := re.MatchString(mail); !ok {
		return UserMail{}, ErrInvalidMail
	}
	return UserMail{
		Mail: mail,
	}, nil
}

type User struct {
	ID                            string
	Email                         UserMail
	Name                          UserName
	IsActive                      bool
	HasAcceptedTermsAndConditions bool
	Phone                         string
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

func NewUser(name, mail string) (*User, error) {

	id := uuid.NewString()
	userMail, err := NewMail(mail)
	if err != nil {
		return nil, err
	}
	userName, err := NewUserName(name)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:                            id,
		Email:                         userMail,
		Name:                          userName,
		IsActive:                      true,
		HasAcceptedTermsAndConditions: false,
		CreatedAt:                     time.Now(),
		UpdatedAt:                     time.Now(),
	}, nil
}
