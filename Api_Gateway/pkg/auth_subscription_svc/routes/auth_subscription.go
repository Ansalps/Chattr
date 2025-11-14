package routes

import (
	"fmt"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func AuthSubscriptionRoutes(router *gin.Engine, authSubscriptionHandler *handler.AuthSubscriptionHandler,tokenSecurityKey *config.Token) {
	router.POST("/refresh",middleware.VerifyJwt([]string{"user","admin"},"refresh",tokenSecurityKey.AdminRefreshKey),authSubscriptionHandler.AccessRegenerator)
	router.POST("/admin/login", authSubscriptionHandler.AdminLogin)
	router.PATCH("/admin/block",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.BlockUser)
	router.PATCH("/admin/unblock",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.UnblockUser)
	router.GET("/admin/get-all-users",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.GetAllUsers)
	router.POST("/admin/subscription-plan",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.CreateSubscriptionPlan)
	router.PUT("/admin/subscription-plan/:id",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.UpdateSubscriptionPlan)
	router.PATCH("/admin/subscription-plan/activate/:id",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.ActivateSubscriptionPlan)
	router.PATCH("/admin/subscription-plan/deactivate/:id",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.DeactivateSubscriptionPlan)
	router.GET("/admin/subscription-plan/get-all-subscription-plans",middleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.GetAllSubscriptionPlans)

	router.POST("/user/signup",authSubscriptionHandler.UserSignUp)
	router.POST("/user/verify-otp",middleware.VerifyJwt([]string{"otpverification"},"access",tokenSecurityKey.OtpVerificationSecurityKey),authSubscriptionHandler.VerifyOtp)
	router.POST("/user/resend-otp",middleware.VerifyJwt([]string{"otpverifcation"},"access",tokenSecurityKey.OtpVerificationSecurityKey),authSubscriptionHandler.ResendOtp)
	router.POST("/user/reset-password",middleware.VerifyJwt([]string{"resetpassword"},"access",tokenSecurityKey.ResetPasswordSecurityKey),authSubscriptionHandler.ResetPassword)
	router.GET("/admin/subscription-plan/get-all-active-subscription-plans",middleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.GetAllActiveSubscriptionPlans)
	fmt.Println("is it reaching in registering routes")
}
