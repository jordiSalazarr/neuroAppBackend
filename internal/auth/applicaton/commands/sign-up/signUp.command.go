package signup

type SignUpCommand struct {
	Mail     string `json:"mail"`
	Name     string `json:"name"`
	Password string `json:"password"`
}
