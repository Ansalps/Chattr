package responsemodels

type AdminLoginResponse struct{
	Admin AdminDetailsResponse
	Token string
}
type AdminDetailsResponse struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
}

type UserSignupResponse struct{
	ID uint
	UserName string
	Name string
	Email string
}