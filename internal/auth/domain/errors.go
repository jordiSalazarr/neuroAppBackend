package domain

import "errors"

var ErrMailNotFound = errors.New("mail is not valid")
var ErrUsernameInvalid = errors.New("username not valid")
