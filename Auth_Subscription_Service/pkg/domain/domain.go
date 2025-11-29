package domain

import (
	"time"

	"gorm.io/gorm"
)

type Admin struct {
	ID       uint   `json:"id" gorm:"uniqueKey; not null"`
	Email    string `json:"email" gorm:"validate:required"`
	Password string `json:"password" gorm:"validate:required"`
}

type User struct {
	gorm.Model
	Name          string
	UserName      string
	Email         string
	Password      string
	Bio           string
	ProfileImgUrl string
	Links         string
	Status        string `gorm:"type:text;default:'pending';check:status IN ('blocked','deleted','pending','active','verified','rejected')"`
	BlueTick	bool	`gorm:"default:false"`
}

type Otp struct {
	ID         uint `gorm:"primaryKey"`
	Email      string
	OTP        uint64
	Expiration time.Time
	Status string `gorom:"type:text;default:'not verified';check:status IN ('not verified','verified')"`
}

type SubscriptionPlan struct{
	gorm.Model
	RazorpayPlanId string
	Name string
	Price int64
	Currency string
	Period string
	Interval uint64
	Description string
	IsActive bool	`gorm:"default:true"`
}
type UserSubscription struct {
    gorm.Model

    UserID uint64 `gorm:"index"`
    SubscriptionPlanID uint64 `gorm:"index"`

    RazorpaySubscriptionId string `gorm:"unique;not null"`
	RazorpayPlanId string

    Status string  // pending, active, cancelled, expired

    StartAt      time.Time
    EndAt        time.Time
    NextChargeAt time.Time

    TotalCount     int
    RemainingCount int
    PaidCount      int

    CancelledAt  time.Time
    CancelReason string
}

type Payment struct{
	gorm.Model
	RazorpaySubscriptionId string
	RazorpayPaymentId string
	RazorpayInvoiceId string
	Amount float64
	Currency string
	PaymentStatus string 
	PaymentMethod string
	TransactionDate time.Time
}