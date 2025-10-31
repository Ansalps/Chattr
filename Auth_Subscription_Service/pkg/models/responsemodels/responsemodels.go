package responsemodels


type AdminLoginResponse struct{
	Admin AdminDetails
	Token string
}
type AdminDetails struct {
	ID        uint   `json:"id"`
	Email     string `json:"email"`
}

type UserSignupResponse struct{
	ID uint
	UserName string
	Name string
	Email string
}