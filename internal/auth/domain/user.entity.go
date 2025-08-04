package domain

import (
	"errors"
	"fmt"
	"math/rand"
	"regexp"
	"time"

	"github.com/google/uuid"
)

var emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`
var ErrInvalidMail = errors.New("invalid mail, please try again")

type Password struct {
	Plain  string
	Hashed string
}

func (p *Password) Hash(es EncryptionService) {
	p.Hashed = es.Hash(p.Plain)
}

func (p Password) Compare(es EncryptionService) bool {
	return es.Compare(p.Plain, p.Hashed)
}

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

func (u *User) GenerateVerificationCode() {
	u.VerificationCode = fmt.Sprintf("%08d", rand.Intn(100000000))
	u.VerificationCodeExpiresAt = time.Now().Add(10 * time.Minute)
}

type User struct {
	ID                            string
	Email                         UserMail
	Password                      Password
	Name                          UserName
	IsVerified                    bool
	IsAdmin                       bool
	IsActive                      bool
	IsAcceptedByAdmin             bool
	HasAcceptedTermsAndConditions bool
	VerificationCode              string
	VerificationCodeExpiresAt     time.Time
	PasswordResetCode             string
	PasswordResetCodeExpiresAt    time.Time
	CreatedAt                     time.Time
	UpdatedAt                     time.Time
}

func NewUser(mail, plainPassword, name string, es EncryptionService) (*User, error) {
	password := Password{
		Plain: plainPassword,
	}
	password.Hash(es)
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
		Password:                      password,
		IsVerified:                    false,
		IsAdmin:                       false,
		IsActive:                      true,
		VerificationCode:              "",
		IsAcceptedByAdmin:             false,
		HasAcceptedTermsAndConditions: true,
		VerificationCodeExpiresAt:     time.Time{},
		PasswordResetCode:             "",
		PasswordResetCodeExpiresAt:    time.Time{},
		CreatedAt:                     time.Now(),
		UpdatedAt:                     time.Now(),
	}, nil
}
