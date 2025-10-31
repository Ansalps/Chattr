package requestmodels

type AdminLoginRequest struct {
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=6,max=20"`
}

type UserSignUpRequest struct {
    Name            string `json:"Name" binding:"required,min=3,max=30"`
    UserName        string `json:"UserName" binding:"required,min=3,max=30"`
    Email           string `json:"Email" binding:"required,email"`
    Password        string `json:"Password" binding:"required,min=3,max=30"`
    ConfirmPassword string `json:"ConfirmPassword" binding:"required,eqfield=Password"`
}