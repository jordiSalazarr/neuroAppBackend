package verifyuser

type VerifyUserCommand struct {
	Email string `json:"email"`
	Code  string `json:"code"`
}
