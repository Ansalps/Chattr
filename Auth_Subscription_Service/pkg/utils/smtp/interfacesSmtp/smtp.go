package interfacesSmtp

type Smtp interface{
	SendVerifcationEmailWithOtp(otp int,recieverEmail string,recieverName string)error
	SendResetPasswordEmailOtp(otp int,recieverEmail string)error
}