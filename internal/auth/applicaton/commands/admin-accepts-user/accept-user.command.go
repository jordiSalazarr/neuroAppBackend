package adminacceptsuser

type AcceptUserCommand struct {
	UserID  string `json:"user_id"`
	AdminID string `json:"admin"`
}
