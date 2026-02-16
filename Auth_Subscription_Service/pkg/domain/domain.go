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
    RazorpayCustomerID string `gorm:"column:razorpay_customer_id;uniqueIndex;size:255"`
}

type Otp struct {
	ID         uint `gorm:"primaryKey"`
	Email      string
	OTP        uint64
	Expiration time.Time
	Status string `gorom:"type:text;default:'not verified';check:status IN ('not verified','verified')"`
}

// type SubscriptionPlan struct{
// 	gorm.Model
// 	RazorpayPlanId string
// 	Name string
// 	Price int64
// 	Currency string
// 	Period string
// 	Interval uint64
// 	Description string
// 	IsActive bool	`gorm:"default:true"`
// }
type SubscriptionPlan struct {
    ID             uint64    `gorm:"primaryKey;autoIncrement"`
    CreatedAt      time.Time 
    UpdatedAt      time.Time
    
    RazorpayPlanId string    `gorm:"uniqueIndex;not null"`
    Name           string    `gorm:"not null"`
    Price          int64     `gorm:"not null"` // Stored in sub-units (e.g., Paise)
    Currency       string    `gorm:"size:10;not null"`
    Period         string    `gorm:"size:20;not null"` 
    Interval       uint64    `gorm:"not null"` // Aligned with your requirement
    Description    string    `gorm:"type:text"`
    IsActive       bool      `gorm:"default:true;index"`
}

type UserSubscription struct {
    ID             uint64    `gorm:"primaryKey;autoIncrement"`
    CreatedAt      time.Time 
    UpdatedAt      time.Time
    
    // Use foreignKey tags to tell GORM exactly how these relate
    UserID             uint64 `gorm:"index;not null"`
    SubscriptionPlanID uint64 `gorm:"index;not null"`

    RazorpaySubscriptionId string `gorm:"uniqueIndex;not null;size:255"` // size helps with certain DB indexes
    RazorpayPlanId         string `gorm:"index;not null"`


    Status   string `gorm:"index;default:'created'"` // Index this as you will query active subs often
    ShortUrl string `gorm:"type:text"` // URLs can sometimes be long

    StartAt      *time.Time
    EndAt        *time.Time
    NextChargeAt *time.Time

    TotalCount     int `gorm:"not null;default:0"`
    RemainingCount int `gorm:"not null;default:0"`
    PaidCount      int `gorm:"not null;default:0"`

    CancelledAt  *time.Time
    CancelReason string `gorm:"type:text"`
}

// type Payment struct{
// 	gorm.Model
// 	RazorpaySubscriptionId string
// 	RazorpayPaymentId string
// 	RazorpayInvoiceId string
// 	Amount float64
// 	Currency string
// 	PaymentStatus string 
// 	PaymentMethod string
// 	TransactionDate time.Time
// }
type SubscriptionPayment struct {
    ID                     uint64      `gorm:"primaryKey"`
    CreatedAt              time.Time
    UserID                 uint64      `gorm:"not null;index"`
    //UserSubscriptionID     uint      `gorm:"not null;index"` // Foreign key to your user_subscriptions table
    RazorpayPaymentID      string    `gorm:"unique;not null"`
    RazorpaySubscriptionID string    `gorm:"index"`
    RazorpayInvoiceID      string    `gorm:"index"`
    Amount                 int64     `gorm:"not null"` // Store in Paise (integer) to avoid floating point issues
    Currency               string    `gorm:"size:10;default:'INR'"`
    Method                 string    `gorm:"size:20"`  // card, upi, netbanking
    Status                 string    `gorm:"size:20"`  // captured, failed, refunded
    TransactionDate        time.Time `gorm:"not null"`
}