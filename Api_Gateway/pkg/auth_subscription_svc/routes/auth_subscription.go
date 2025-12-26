package routes

import (
	"fmt"

	"github.com/Ansalps/Chattr_Api_Gateway/pkg/auth_subscription_svc/handler"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/config"
	"github.com/Ansalps/Chattr_Api_Gateway/pkg/middleware"
	"github.com/gin-gonic/gin"
)

func AuthSubscriptionRoutes(router *gin.Engine, authSubscriptionHandler *handler.AuthSubscriptionHandler,tokenSecurityKey *config.Token,authMiddleware *middleware.AuthMiddleware) {
	router.POST("/admin/refresh",authMiddleware.VerifyJwt([]string{"user","admin"},"refresh",tokenSecurityKey.AdminRefreshKey),authSubscriptionHandler.AccessRegenerator)
	router.POST("/user/refresh",authMiddleware.VerifyJwt([]string{"user","admin"},"refresh",tokenSecurityKey.UserRefreshKey),authSubscriptionHandler.AccessRegenerator)
	router.POST("/admin/login", authSubscriptionHandler.AdminLogin)
	router.PATCH("/admin/block",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.BlockUser)
	router.PATCH("/admin/unblock",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.UnblockUser)
	router.GET("/admin/get-all-users",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.GetAllUsers)
	router.POST("/admin/subscription-plan",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.CreateSubscriptionPlan)

	router.PATCH("/admin/subscription-plan/activate/:id",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.ActivateSubscriptionPlan)
	router.PATCH("/admin/subscription-plan/deactivate/:id",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.DeactivateSubscriptionPlan)
	router.GET("/admin/subscription-plan/get-all-subscription-plans",authMiddleware.VerifyJwt([]string{"admin"},"access",tokenSecurityKey.AdminSecurityKey),authSubscriptionHandler.GetAllSubscriptionPlans)

	router.POST("/user/signup",authSubscriptionHandler.UserSignUp)
	router.POST("/user/verify-otp",authMiddleware.VerifyJwt([]string{"otpverification"},"access",tokenSecurityKey.OtpVerificationSecurityKey),authSubscriptionHandler.VerifyOtp)
	router.POST("/user/resend-otp",authMiddleware.VerifyJwt([]string{"otpverifcation"},"access",tokenSecurityKey.OtpVerificationSecurityKey),authSubscriptionHandler.ResendOtp)
	router.POST("/user/forgot-password",authSubscriptionHandler.ForgotPassord)
	router.POST("/user/reset-password",authMiddleware.VerifyJwt([]string{"resetpassword"},"access",tokenSecurityKey.ResetPasswordSecurityKey),authSubscriptionHandler.ResetPassword)
	router.POST("/user/login",authSubscriptionHandler.UserLogin)
	router.GET("/user/subscription-plan/get-all-active-subscription-plans",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.GetAllActiveSubscriptionPlans)
	router.POST("/user/subscribe/:plan_id",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.Subscribe)
	router.POST("/user/verify-subscription-payment",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.VerifySubscriptionPayment)
	router.POST("/user/unsubscribe/:sub_id",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.Unsubscribe)

	router.GET("/user/get-profile-information",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.GetProfileInformation)
	router.PATCH("/user/edit-profile-information",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.EditProfileInformation)
	router.GET("/user/:user_id/get-public-profile",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.GetPublicProfile)
	router.POST("/user/set-profile-image",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.SetProfileImage)
	router.PATCH("/user/change-password",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.ChangePassword)
	//router.POST("/webhook",authSubscriptionHandler.Webhook)

	router.GET("/user/search",authSubscriptionHandler.SearchUser)
	fmt.Println("is it reaching in registering routes")

	router.POST("/user/logout",authMiddleware.VerifyJwt([]string{"user"},"access",tokenSecurityKey.UserSecurityKey),authSubscriptionHandler.Logout)
}
