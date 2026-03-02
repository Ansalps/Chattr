package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"
	"github.com/lib/pq"

	//"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/repository/interfacesRepository"
	"gorm.io/gorm"
)

type AuthSubscriptionRepository struct {
	DB *gorm.DB
}

func NewAuthSubscriptionRepository(db *gorm.DB) interfacesRepository.AuthSubscriptionRepository {
	return &AuthSubscriptionRepository{
		DB: db,
	}
}

func (ad *AuthSubscriptionRepository) CheckAdminExistsByEmail(ctx context.Context, email string) (*domain.Admin, error) {
	var admin domain.Admin
	res := ad.DB.WithContext(ctx).Where("email = ?", email).First(&admin)
	if res.Error != nil {
		if errors.Is(res.Error, context.DeadlineExceeded) {
			return nil, fmt.Errorf("%w: %v:",domain.ErrDatabaseConnectionTimeOut,res.Error) // This happens if the 10s limit is hit
		}
		return nil, res.Error
	}
	return &admin, nil
}
func (ad *AuthSubscriptionRepository) DeletePendingUser(ctx context.Context,email string) error {
	query := `DELETE FROM users WHERE email=? and status='pending'`
	if err := ad.DB.WithContext(ctx).Exec(query, email).Error; err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("%w: %v:",domain.ErrDatabaseConnectionTimeOut,err) // This happens if the 10s limit is hit
		}
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) CheckUserExistsByEmail(ctx context.Context,email string) (*domain.User, error) {
	var user domain.User
	res := ad.DB.WithContext(ctx).Where("email=?", email).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error,context.DeadlineExceeded){
			return nil,fmt.Errorf("%w: %v:",domain.ErrDatabaseConnectionTimeOut,res.Error)
		}
		return nil, res.Error
	}
	return &user, nil
}

func (ad *AuthSubscriptionRepository) CheckUserExistsByUseraname(ctx context.Context,username string) (*domain.User, error) {
	var user domain.User
	res := ad.DB.WithContext(ctx).Where("user_name=?", username).First(&user)
	if res.Error != nil {
		if errors.Is(res.Error,context.DeadlineExceeded){
			return nil,fmt.Errorf("%w: %v:",domain.ErrDatabaseConnectionTimeOut,res.Error)
		}
		return nil, res.Error
	}
	return &user, nil
}

func (ad *AuthSubscriptionRepository) TemporarySavingUserOtp(otp int, userEmail string, expiration time.Time) error {

	query := `INSERT INTO otps (email, otp, expiration) VALUES ($1, $2, $3)`
	err := ad.DB.Exec(query, userEmail, otp, expiration).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) CreateUser(userData *requestmodels.UserSignUpRequest) (*responsemodels.UserSignupResponse, error) {
	// var user responsemodels.UserSignupResponse
	// query := "INSERT INTO users (name,user_name, email, password) VALUES($1, $2, $3, $4) RETURNING id, user_name, name, email"
	// err := ad.DB.Raw(query, userData.Name, userData.UserName, userData.Email, userData.Password).Scan(&user).Error
	// if err != nil {
	// 	return nil, err
	// }
	// return &user, nil
	user := domain.User{
		Name:     userData.Name,
		UserName: userData.UserName,
		Email:    userData.Email,
		Password: userData.Password,
		Phone:    userData.Phone,
	}
	if err := ad.DB.Create(&user).Error; err != nil {
		return nil, err
	}
	userRes := responsemodels.UserSignupResponse{
		ID:       user.ID,
		UserName: user.UserName,
		Name:     user.Name,
		Email:    user.Email,
	}
	return &userRes, nil
}

func (ad *AuthSubscriptionRepository) CheckOtpExistsByEmail(otpReq requestmodels.OtpRequest) (*domain.Otp, error) {
	var otp domain.Otp
	res := ad.DB.Where("email=?", otpReq.Email).First(&otp)
	if res.Error != nil {
		return nil, res.Error
	}
	return &otp, nil
}

func (ad *AuthSubscriptionRepository) ChangeOtpStatus(email string) error {
	query := `UPDATE otps set status='verifed' where email=?`
	err := ad.DB.Exec(query, email).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) DeleteOtpByEmail(email string) error {
	query := `DELETE from otps where email=?`
	err := ad.DB.Exec(query, email).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) ChangeUserStatusByEmail(email string) error {
	query := `UPDATE users set status='active' where email=?`
	err := ad.DB.Exec(query, email).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) UpdatePassword(resetPasswordReq requestmodels.ResetPasswordRequest) error {
	query := `UPDATE users SET password=? WHERE email=?`
	err := ad.DB.Exec(query, resetPasswordReq.Password, resetPasswordReq.Email).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) CheckUserStatus(userid uint64) (string, error) {
	var status string
	query := `SELECT status FROM users WHERE id=?`
	err := ad.DB.Raw(query, userid).Scan(&status).Error
	if err != nil {
		return "", err
	}
	return status, nil
}

func (ad *AuthSubscriptionRepository) ChangeUserStatusToBlockedByUserId(blockUserReq requestmodels.BlockUserRequest) error {
	query := `UPDATE users SET status='blocked' WHERE id=?`
	err := ad.DB.Exec(query, blockUserReq.UserId).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) ChangeUserStatusToActiveByUserId(unblockUserReq requestmodels.UnblockUserRequest) error {
	query := `UPDATE users SET status='active' WHERE id=?`
	err := ad.DB.Exec(query, unblockUserReq.UserId).Error
	if err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) GetAllUsers(getAllUsersReq requestmodels.GetAllUsersRequest) (responsemodels.GetAllUsersResponse, error) {
	var user []responsemodels.User
	query := `SELECT * FROM users ORDER BY id LIMIT $1 OFFSET $2`
	err := ad.DB.Raw(query, getAllUsersReq.Limit, getAllUsersReq.Offset).Scan(&user).Error
	if err != nil {
		return responsemodels.GetAllUsersResponse{}, err
	}
	return responsemodels.GetAllUsersResponse{
		Users: user,
	}, nil
}

func (ad *AuthSubscriptionRepository) CreateSubscriptionPlan(dto *domain.CreatedPlanDTO) (responsemodels.CreateSubscriptionPlanResponse, error) {

	subscriptionPlan := &domain.SubscriptionPlan{
		RazorpayPlanId: dto.ID,
		Name:           dto.Name,
		Price:          dto.Amount, // Converting Paise to Rupees
		Currency:       dto.Currency,
		Period:         dto.Period,
		Interval:       uint64(dto.Interval),
		Description:    dto.Description,
		IsActive:       dto.Active,
	}

	if err := ad.DB.Create(&subscriptionPlan).Error; err != nil {
		return responsemodels.CreateSubscriptionPlanResponse{}, err
	}

	return responsemodels.CreateSubscriptionPlanResponse{
		ID:          uint64(subscriptionPlan.ID),
		CreatedAt:   subscriptionPlan.CreatedAt,
		UpdatedAt:   subscriptionPlan.UpdatedAt,
		Name:        subscriptionPlan.Name,
		Price:       subscriptionPlan.Price,
		Currency:    subscriptionPlan.Currency,
		Period:      subscriptionPlan.Period,
		Interval:    subscriptionPlan.Interval,
		Description: subscriptionPlan.Description,
		IsActive:    subscriptionPlan.IsActive,
	}, nil
}

func (ad *AuthSubscriptionRepository) CreateSubscription(req requestmodels.SubscribeRequest, dto *domain.CreatedSubscriptionDTO) (responsemodels.SubscribeResponse, error) {

	userSubscription := &domain.UserSubscription{
		UserID:                 req.UserId,
		SubscriptionPlanID:     req.PlanId,
		RazorpaySubscriptionId: dto.ID,
		RazorpayPlanId:         dto.PlanID,
		Status:                 dto.Status,   // Usually "created"
		ShortUrl:               dto.ShortURL, // Save the URL to the DB!
		TotalCount:             dto.TotalCount,
		RemainingCount:         dto.RemainingCount,
		PaidCount:              dto.PaidCount,
	}

	if err := ad.DB.Create(&userSubscription).Error; err != nil {
		return responsemodels.SubscribeResponse{}, err
	}

	return responsemodels.SubscribeResponse{
		ID:        userSubscription.ID,
		CreatedAt: userSubscription.CreatedAt,
		UpdatedAt: userSubscription.UpdatedAt,

		UserID: userSubscription.UserID,

		RazorpaySubscriptionId: userSubscription.RazorpaySubscriptionId,
		SubscriptionPlanID:     userSubscription.SubscriptionPlanID,
		RazorpayPlanId:         userSubscription.RazorpayPlanId,

		Status:   userSubscription.Status,
		ShortUrl: userSubscription.ShortUrl,

		TotalCount:     userSubscription.TotalCount,
		RemainingCount: userSubscription.RemainingCount,
		PaidCount:      userSubscription.PaidCount,
	}, nil
}

func (ad *AuthSubscriptionRepository) FetchStatusFromSubcriptionPlan(id uint64) (bool, error) {
	var status bool
	query := `SELECT is_active FROM subscription_plans WHERE id=?`
	result := ad.DB.Raw(query, id).Scan(&status)
	if result.Error != nil {
		return false, result.Error
	}
	if result.RowsAffected == 0 {
		return false, gorm.ErrRecordNotFound
	}
	return status, nil
}

func (ad *AuthSubscriptionRepository) ActivateSubscriptionPlan(activateSubscriptionPlanReq requestmodels.ActivateSubscriptionPlanRequest) (responsemodels.ActivateSubscriptionPlanResponse, error) {
	// Update only is_active
	if err := ad.DB.Model(&domain.SubscriptionPlan{}).
		Where("id = ?", activateSubscriptionPlanReq.ID).
		Update("is_active", true).Error; err != nil {

		return responsemodels.ActivateSubscriptionPlanResponse{}, err
	}

	// Fetch the updated row
	var subscriptionPlan domain.SubscriptionPlan
	if err := ad.DB.First(&subscriptionPlan, activateSubscriptionPlanReq.ID).Error; err != nil {
		return responsemodels.ActivateSubscriptionPlanResponse{}, err
	}

	// Build response
	return responsemodels.ActivateSubscriptionPlanResponse{
		ID:             uint64(subscriptionPlan.ID),
		CreatedAt:      subscriptionPlan.CreatedAt,
		UpdatedAt:      subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name:           subscriptionPlan.Name,
		Price:          subscriptionPlan.Price,
		Currency:       subscriptionPlan.Currency,
		Period:         subscriptionPlan.Period,
		Interval:       subscriptionPlan.Interval,
		Description:    subscriptionPlan.Description,
		IsActive:       subscriptionPlan.IsActive,
	}, nil
}

func (ad *AuthSubscriptionRepository) DeactivateSubscriptionPlan(deactivateSubscriptionPlanReq requestmodels.DeactivateSubscriptionPlanRequest) (responsemodels.DeactivateSubscriptionPlanResponse, error) {
	// Update only is_active
	if err := ad.DB.Model(&domain.SubscriptionPlan{}).
		Where("id = ?", deactivateSubscriptionPlanReq.ID).
		Update("is_active", false).Error; err != nil {

		return responsemodels.DeactivateSubscriptionPlanResponse{}, err
	}

	// Fetch the updated row
	var subscriptionPlan domain.SubscriptionPlan
	if err := ad.DB.First(&subscriptionPlan, deactivateSubscriptionPlanReq.ID).Error; err != nil {
		return responsemodels.DeactivateSubscriptionPlanResponse{}, err
	}

	// Build response
	return responsemodels.DeactivateSubscriptionPlanResponse{
		ID:             uint64(subscriptionPlan.ID),
		CreatedAt:      subscriptionPlan.CreatedAt,
		UpdatedAt:      subscriptionPlan.UpdatedAt,
		RazorpayPlanId: subscriptionPlan.RazorpayPlanId,
		Name:           subscriptionPlan.Name,
		Price:          subscriptionPlan.Price,
		Currency:       subscriptionPlan.Currency,
		Period:         subscriptionPlan.Period,
		Interval:       subscriptionPlan.Interval,
		Description:    subscriptionPlan.Description,
		IsActive:       subscriptionPlan.IsActive,
	}, nil
}

func (ad *AuthSubscriptionRepository) GetAllSubscriptionPlans(getAllSubscriptionPlansReq requestmodels.GetAllSubscriptionPlansRequest) (responsemodels.GetAllSubscriptionPlansResponse, error) {
	var subscriptionPlans []responsemodels.SubscriptionPlan
	query := `SELECT * FROM subscription_plans ORDER BY id LIMIT $1 OFFSET $2`
	err := ad.DB.Raw(query, getAllSubscriptionPlansReq.Limit, getAllSubscriptionPlansReq.Offset).Scan(&subscriptionPlans).Error
	if err != nil {
		return responsemodels.GetAllSubscriptionPlansResponse{}, nil
	}
	return responsemodels.GetAllSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	}, nil
}

func (ad *AuthSubscriptionRepository) GetAllActiveSubscriptionPlans(getAllActiveSubscriptionPlansReq requestmodels.GetAllActiveSubscriptionPlansRequest) (responsemodels.GetAllActiveSubscriptionPlansResponse, error) {
	var subscriptionPlans []responsemodels.SubscriptionPlan
	query := `SELECT * FROM subscription_plans WHERE is_active=$1 ORDER BY ID LIMIT $2 OFFSET $3`
	err := ad.DB.Raw(query, true, getAllActiveSubscriptionPlansReq.Limit, getAllActiveSubscriptionPlansReq.Offset).Scan(&subscriptionPlans).Error
	if err != nil {
		return responsemodels.GetAllActiveSubscriptionPlansResponse{}, nil
	}
	return responsemodels.GetAllActiveSubscriptionPlansResponse{
		SubscriptionPlans: subscriptionPlans,
	}, nil
}

func (ad *AuthSubscriptionRepository) FetchRazorpayPlanIdFromId(id uint64) (string, error) {
	var RazorpayPlanId string
	query := `SELECT razorpay_plan_id FROM subscription_plans WHERE id=?`
	result := ad.DB.Raw(query, id).Scan(&RazorpayPlanId)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return RazorpayPlanId, nil
}

func (ad *AuthSubscriptionRepository) UpdateUserSubscripion(id string, subscription map[string]interface{}) (responsemodels.VerifySubscriptionPaymentResponse, error) {
	startAtUnix, ok := subscription["start_at"].(float64)
	if !ok {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("current_start is probably nil")
	}
	endAtUnix, ok := subscription["end_at"].(float64)
	if !ok {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("current_end is probably nil")
	}
	nextChargeAtUnix, ok := subscription["charge_at"]
	if !ok {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("charge_at is probably nil")
	}

	startAt := time.Unix(int64(startAtUnix), 0)
	endAt := time.Unix(int64(endAtUnix), 0)
	nextChargeAt := time.Unix(int64(nextChargeAtUnix.(float64)), 0)
	// Extract numeric fields
	paidCount, ok := subscription["paid_count"].(float64)
	if !ok {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("paid_count is missing or invalid")
	}

	remainingCount, ok := subscription["remaining_count"].(float64)
	if !ok {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("remaining_count is missing or invalid")
	}
	//fmt.Println("start at ***", startAt, "endAt **", endAt, "nextChargeAt", nextChargeAt)
	// Update DB
	updateData := map[string]any{
		"status":          "active",
		"start_at":        startAt,
		"end_at":          endAt,
		"next_charge_at":  nextChargeAt,
		"paid_count":      int(paidCount),
		"remaining_count": int(remainingCount),
	}
	if err := ad.DB.Model(&domain.UserSubscription{}).
		Where("razorpay_subscription_id = ?", id).
		Updates(updateData).Error; err != nil {
		return responsemodels.VerifySubscriptionPaymentResponse{}, fmt.Errorf("failed to update subscription: %w", err)
	}
	// Fetch the updated row
	var userSubscription domain.UserSubscription
	query := `SELECT * FROM user_subscriptions where razorpay_subscription_id=?`
	if err := ad.DB.Raw(query, id).Scan(&userSubscription).Error; err != nil {
		return responsemodels.VerifySubscriptionPaymentResponse{}, err
	}
	return responsemodels.VerifySubscriptionPaymentResponse{
		ID:                     uint64(userSubscription.ID),
		CreatedAt:              userSubscription.CreatedAt,
		UpdatedAt:              userSubscription.UpdatedAt,
		UserID:                 userSubscription.UserID,
		RazorpaySubscriptionId: userSubscription.RazorpaySubscriptionId,
		Status:                 userSubscription.Status,
		// StartAt:                userSubscription.StartAt,
		// EndAt:                  userSubscription.EndAt,
		// NextChargeAt:           userSubscription.NextChargeAt,
		TotalCount:     userSubscription.TotalCount,
		RemainingCount: userSubscription.RemainingCount,
		PaidCount:      userSubscription.PaidCount,
	}, nil
}

type SubPlan struct {
	Price    int64
	Currency string
}

func (ad *AuthSubscriptionRepository) FetchAmountCurrencyFromSubscriptionPlan(id uint64) (int64, string, error) {
	var plan SubPlan
	query := `SELECT price,currency FROM subscription_plans WHERE id=?`
	if err := ad.DB.Raw(query, id).Scan(&plan).Error; err != nil {
		return 0, "", err
	}
	return plan.Price, plan.Currency, nil
}

func (ad *AuthSubscriptionRepository) FetchRazorpaySubscriptionIdFromUserId(userid uint64) (string, error) {
	var razorpaySubscriptionId string
	query := `SELECT razorpay_subscription_id FROM user_subscriptions WHERE user_id=? and status!='cancelled' and status!='completed'`
	result := ad.DB.Raw(query, userid).Scan(&razorpaySubscriptionId)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return razorpaySubscriptionId, nil
}

func (ad *AuthSubscriptionRepository) SetCancelReason(req requestmodels.UnsubscribeRequest) (responsemodels.UnsubscribeResponse, error) {
	var subId uint64
	query := `UPDATE user_subscriptions SET cancel_reason=$1 WHERE razorpay_subscription_id=$2 returning id`
	result := ad.DB.Raw(query, req.CancelReason, req.RazorpaySubId).Scan(&subId)
	if result.Error != nil {
		return responsemodels.UnsubscribeResponse{}, result.Error
	}
	// if result.RowsAffected==0{
	// 	return responsemodels.UnsubscribeResponse{},gorm.ErrRecordNotFound
	// }
	//fmt.Println("printing user id", userSubscription.UserID)
	return responsemodels.UnsubscribeResponse{
		SubId: subId,
	}, nil
}

func (ad *AuthSubscriptionRepository) FetchUserIdFromSubscriptionId(razorpaySubId string) (uint64, error) {
	var userid uint64
	query := `SELECT user_id from user_subscriptions WHERE razorpay_subscription_id=?`
	if err := ad.DB.Raw(query, razorpaySubId).Scan(&userid).Error; err != nil {
		return 0, err
	}
	return userid, nil
}

func (ad *AuthSubscriptionRepository) TurnBlueTickTrueForUserId(userid uint64) error {
	//fmt.Println("most probably user_id", userid)
	query := `UPDATE users SET blue_tick=true where id=?`
	if err := ad.DB.Exec(query, userid).Error; err != nil {
		return err
	}
	return nil
}

// func (ad *AuthSubscriptionRepository) PopulatePayment(payment map[string]interface{}, verifySubscripitionPaymentReq requestmodels.VerifySubscriptionPaymentRequest) (domain.Payment, error) {
// 	//var payment domain.Payment
// 	razorpayPaymentId := verifySubscripitionPaymentReq.RazorpayPaymentId
// 	razorpaySubscriptionId := verifySubscripitionPaymentReq.RazorpaySubscriptionId
// 	// Extracting the root-level fields
// 	razorpayInvoiceId, ok := payment["invoice_id"].(string)
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("invoice id is missing or not a string")
// 	}

// 	amount, ok := payment["amount"].(float64)
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("amount is missing or not a float64")
// 	}

// 	// Numbers are float64 in the subscription map
// 	currency, ok := payment["currency"].(string)
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("currency is missing or not a string")
// 	}

// 	paymentStatus, ok := payment["status"].(string)
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("status is missing or not a string")
// 	}

// 	paymentMethod, ok := payment["method"].(string)
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("method is missing or not a string")
// 	}

// 	createdAtUnix, ok := payment["created_at"].(float64) // `created_at` is usually a float64 in JSON
// 	if !ok {
// 		return domain.Payment{}, fmt.Errorf("created_at is missing or not a number")
// 	}

// 	// Convert Unix timestamp to time.Time
// 	createdAt := time.Unix(int64(createdAtUnix), 0) // Unix timestamp is in seconds, so we use 0 for nanoseconds

// 	// Step 3 — Create the UserSubscription struct
// 	paymentCreate := &domain.Payment{
// 		RazorpaySubscriptionId: razorpaySubscriptionId,
// 		RazorpayPaymentId:      razorpayPaymentId,
// 		RazorpayInvoiceId:      razorpayInvoiceId,
// 		Amount:                 amount / 100,
// 		Currency:               currency,
// 		PaymentStatus:          paymentStatus,
// 		PaymentMethod:          paymentMethod,
// 		TransactionDate:        createdAt,
// 	}

// 	if err := ad.DB.Create(&paymentCreate).Error; err != nil {
// 		return domain.Payment{}, err
// 	}

// 	return *paymentCreate, nil
// }

func (ad *AuthSubscriptionRepository) FetchRazorpayPlanIdFromRazrorpaySubscriptionId(subId string) (string, error) {
	var palnId string
	query := `SELECT razorpay_plan_id FROM user_subscriptions where razorpay_subscription_id=?`
	if err := ad.DB.Raw(query, subId).Scan(&palnId).Error; err != nil {
		return "", err
	}
	return palnId, nil
}

type periodInterval struct {
	Period   string
	Interval uint64
}

func (ad *AuthSubscriptionRepository) FetchIntervalPeriodFromSubscriptionPlan(planId string) (string, uint64, error) {
	var p periodInterval
	query := `SELECT period,interval FROM subscription_plans where razorpay_plan_id=?`
	if err := ad.DB.Raw(query, planId).Scan(&p).Error; err != nil {
		return "", 0, err
	}
	return p.Period, p.Interval, nil
}

func (ad *AuthSubscriptionRepository) FetchTotalCountFromUserSubscription(subId string) (int, error) {
	var totalCount int
	query := `SELECT total_count FROM user_subscriptions WHERE razorpay_subscription_id=?`
	if err := ad.DB.Raw(query, subId).Scan(&totalCount).Error; err != nil {
		return 0, err
	}
	return totalCount, nil
}

func (ad *AuthSubscriptionRepository) UpdateTimeUserSubscription(startAt, endAt, nextChargeAt time.Time, subid string) (responsemodels.VerifySubscriptionPaymentResponse, error) {
	var subscribe responsemodels.VerifySubscriptionPaymentResponse
	updated_at := time.Now()
	query := `UPDATE user_subscriptions SET updated_at=?,start_at=?,end_at=?,next_charge_at=? WHERE razorpay_subscription_id=?`
	if err := ad.DB.Exec(query, updated_at, startAt, endAt, nextChargeAt, subid).Error; err != nil {
		return responsemodels.VerifySubscriptionPaymentResponse{}, err
	}
	query1 := `SELECT * FROM user_subscriptions WHERE razorpay_subscription_id=?`
	if err := ad.DB.Raw(query1, subid).Scan(&subscribe).Error; err != nil {
		return responsemodels.VerifySubscriptionPaymentResponse{}, err
	}
	return subscribe, nil
}

func (ad *AuthSubscriptionRepository) FetchNextChargeAtFromUserSubcription(subid string) (time.Time, error) {
	var nextChargeAt time.Time
	query := `SELECT next_charge_at from user_subscriptions WHERE razorpay_subscription_id=?`
	if err := ad.DB.Raw(query, subid).Scan(&nextChargeAt).Error; err != nil {
		return time.Time{}, err
	}
	return nextChargeAt, nil
}

func (ad *AuthSubscriptionRepository) TurnOffBlueTickForUserId(userid uint64) error {
	query := `UPDATE users SET blue_tick=false WHERE id=?`
	if err := ad.DB.Exec(query, userid).Error; err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) UpdateProfileImage(userid uint64, imageUrl string) error {
	query := `UPDATE users SET profile_img_url=? WHERE id=?`
	if err := ad.DB.Exec(query, imageUrl, userid).Error; err != nil {
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) CheckUserExistsById(userId uint64) (bool, error) {

	var exists bool

	err := ad.DB.Raw(
		"SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)",
		userId,
	).Scan(&exists).Error

	if err != nil {
		return false, err
	}
	return exists, nil
}
func (ad *AuthSubscriptionRepository) GetProfileInformation(req requestmodels.GetProfileInformationRequest) (responsemodels.GetProfileInformationResponse, error) {
	var resp responsemodels.GetProfileInformationResponse
	query := `SELECT id as user_id,name,user_name,email,bio,profile_img_url,links,blue_tick,phone FROM users WHERE id=$1`
	result := ad.DB.Raw(query, req.UserId).Scan(&resp)
	if result.Error != nil {
		return responsemodels.GetProfileInformationResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.GetProfileInformationResponse{}, gorm.ErrRecordNotFound
	}
	//fmt.Println("resp in repo", resp, resp.UserID)
	return resp, nil
}
func (ad *AuthSubscriptionRepository) EditProfileInformation(userId uint64, updateData map[string]interface{}) (responsemodels.EditProfile, error) {
	//fmt.Println("data in repo", updateData)
	// 1. Still perform the update
	if err := ad.DB.Model(&domain.User{}).Where("id = ?", userId).Updates(updateData).Error; err != nil {
		return responsemodels.EditProfile{}, err
	}

	resp := responsemodels.EditProfile{
		UserID: userId,
	}

	// 2. Only populate fields that were in the original update request
	if val, ok := updateData["name"].(string); ok {
		//fmt.Println("is reaching")
		resp.Name = &val
	}
	if val, ok := updateData["bio"].(string); ok {
		//	fmt.Println("is reaching 2")
		resp.Bio = &val
	}
	if val, ok := updateData["links"].(string); ok {
		//fmt.Println("is reaching 3")
		resp.Links = &val
	}
	if val, ok := updateData["phone"].(string); ok {
		//fmt.Println("is reaching 4")
		resp.Links = &val
	}
	//fmt.Println("resp", resp)
	return resp, nil
}
func (ad *AuthSubscriptionRepository) FetchHashedPassword(req requestmodels.ChangePassword) (string, error) {
	var password string
	query := `SELECT password FROM users WHERE id=$1`
	result := ad.DB.Raw(query, req.UserID).Scan(&password)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return password, nil
}
func (ad *AuthSubscriptionRepository) ChangePassword(req requestmodels.ChangePassword, passwordHash string) (responsemodels.ChangePasswordResponse, error) {
	query := `UPDATE users SET updated_at=$1,password=$2 WHERE id=$3`
	if err := ad.DB.Exec(query, time.Now(), passwordHash, req.UserID).Error; err != nil {
		return responsemodels.ChangePasswordResponse{}, err
	}
	return responsemodels.ChangePasswordResponse{
		UserID: req.UserID,
	}, nil
}
func (ad *AuthSubscriptionRepository) SearchUser(req requestmodels.SearchUser) (responsemodels.SearchUserResponse, error) {
	var userMeataData []responsemodels.UserMetaData
	text := "%" + req.SearchText + "%"
	query := `SELECT id as user_id,user_name,name,profile_img_url FROM users WHERE user_name ILIKE $1 LIMIT $2 OFFSET $3`
	result := ad.DB.Raw(query, text, req.Limit, req.Offset).Scan(&userMeataData)
	if result.Error != nil {
		return responsemodels.SearchUserResponse{}, result.Error
	}
	return responsemodels.SearchUserResponse{
		Usermetadata: userMeataData,
	}, nil
}
func (ad *AuthSubscriptionRepository) FetchUserPublicData(userid uint64) (responsemodels.UserPublicDataResponse, error) {
	var resp responsemodels.UserPublicDataResponse
	query := `SELECT id as user_id,user_name,name,profile_img_url,bio,links,blue_tick,razorpay_customer_id,phone FROM users WHERE id=$1`
	result := ad.DB.Raw(query, userid).Scan(&resp)
	if result.Error != nil {
		return responsemodels.UserPublicDataResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.UserPublicDataResponse{}, gorm.ErrRecordNotFound
	}
	return resp, nil
}

func (ad *AuthSubscriptionRepository) UpdateUserRazorpayCustomerID(userid uint64, customerid string) error {
	query := `UPDATE users SET razorpay_customer_id=$1 WHERE id=$2`
	result := ad.DB.Exec(query, customerid, userid)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ad *AuthSubscriptionRepository) FetchUserMetaData(userids []uint64) (map[uint64]responsemodels.UserMetaData, error) {
	var resp []responsemodels.UserMetaData
	query := `SELECT id as user_id,user_name,name,profile_img_url,blue_tick FROM users WHERE id IN ?`

	result := ad.DB.Raw(query, userids).Scan(&resp)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	//conver slice --> map
	m := make(map[uint64]responsemodels.UserMetaData, len(resp))
	for _, r := range resp {
		m[r.UserID] = r
	}
	return m, nil
}

func (ad *AuthSubscriptionRepository) CheckUserListExists(userids []uint64) ([]uint64, error) {
	var UserId []uint64
	query := `SELECT id as user_id FROM users WHERE id IN ?`
	result := ad.DB.Raw(query, userids).Scan(&UserId)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return UserId, nil
}

func (ad *AuthSubscriptionRepository) GetSubscriptionDetails(req requestmodels.GetSubscriptionDetails) (responsemodels.GetSubscriptionDetails, error) {
	var resp responsemodels.GetSubscriptionDetails
	query := `
    SELECT 
        us.*, 
        sp.id AS plan_id, 
        sp.created_at AS plan_created_at, 
        sp.updated_at AS plan_updated_at,
        sp.name, 
        sp.price, 
        sp.currency,
        sp.period,
        sp.interval,
        sp.description,
        sp.is_active,
        sp.razorpay_plan_id
    FROM user_subscriptions us 
    JOIN subscription_plans sp ON us.razorpay_plan_id = sp.razorpay_plan_id 
    WHERE us.user_id = $1 and (us.status='active' or us.status='halted' or us.status='created')
    LIMIT 1`
	result := ad.DB.Raw(query, req.UserID).Scan(&resp)
	if result.Error != nil {
		return responsemodels.GetSubscriptionDetails{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.GetSubscriptionDetails{}, gorm.ErrRecordNotFound
	}
	return resp, nil

}

func (ad *AuthSubscriptionRepository) UpddateActivatedSubscription(req requestmodels.WebhookSubscriptionActivatedRequest) (responsemodels.WebhookSubscriptionActivatedResponse, error) {
	query := `UPDATE user_subscriptionS SET status=$1,paid_count=$2,remaining_count=$3,start_at=$4,end_at=$5 WHERE razorpay_subscription_id=$6`
	result := ad.DB.Exec(query, req.Status, req.PaidCount, req.RemainingCount, req.StartAt, req.EndAt, req.RazorpaySubscriptionId)
	if result.Error != nil {
		log.Println("database error", result.Error)
		return responsemodels.WebhookSubscriptionActivatedResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		log.Println("no rows affected")
		return responsemodels.WebhookSubscriptionActivatedResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.WebhookSubscriptionActivatedResponse{
		RazorpaySubcriptionId: req.RazorpaySubscriptionId,
	}, nil
}

func (ad *AuthSubscriptionRepository) UpdateNextChargeAt(nextChargeAt time.Time, razorpaySubId string) error {
	query := `UPDATE user_subscriptions SET next_charge_at=$1 WHERE razorpay_subscription_id=$2`
	result := ad.DB.Exec(query, nextChargeAt, razorpaySubId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
func (ad *AuthSubscriptionRepository) UpdateStatusToActive(razorpaySubId string) error {
	query := `UPDATE user_subscriptions SET status='active' WHERE razorpay_subscription_id=$1`
	result := ad.DB.Exec(query, razorpaySubId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ad *AuthSubscriptionRepository) UpdatePayment(req requestmodels.WebhookSubscriptionChargedRequest) (responsemodels.WebhookSubscriptionChargedResponse, error) {
	payment := &domain.SubscriptionPayment{
		UserID:                 req.UserID,
		RazorpayPaymentID:      req.PaymentID,
		RazorpaySubscriptionID: req.RazorpaySubscriptionId,
		RazorpayInvoiceID:      req.InvoiceID,
		Amount:                 req.Amount,
		Currency:               req.Currency,
		Method:                 req.Method,
		Status:                 req.Status,
		TransactionDate:        req.TransactionDate,
	}
	err := ad.DB.Create(&payment).Error
	if err != nil {
		return responsemodels.WebhookSubscriptionChargedResponse{}, err
	}
	return responsemodels.WebhookSubscriptionChargedResponse{
		RazorpaySubcriptionId: req.RazorpaySubscriptionId,
	}, nil
}

func (ad *AuthSubscriptionRepository) UpdateStatusHalted(req requestmodels.WebhookSubscriptionHaltedRequest) error {
	query := `UPDATE user_subscriptions SET status=$1 WHERE razorpay_subscription_id=$2`
	result := ad.DB.Exec(query, req.Status, req.RazorpaySubscriptionId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ad *AuthSubscriptionRepository) UpdateSubscriptionCancelled(req requestmodels.WebhookSubscriptionCancelledRequest) (responsemodels.WebhookSubscriptionCancelledResponse, error) {
	query := `UPDATE user_subscriptions SET status=$1,cancelled_at=$2 WHERE razorpay_subscription_id=$3`
	result := ad.DB.Exec(query, req.Status, req.CancelledAt, req.RazorpaySubscriptionId)
	if result.Error != nil {
		return responsemodels.WebhookSubscriptionCancelledResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.WebhookSubscriptionCancelledResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.WebhookSubscriptionCancelledResponse{
		RazorpaySubcriptionId: req.RazorpaySubscriptionId,
	}, nil
}

func (ad *AuthSubscriptionRepository) UpdateSubscripionCompleted(req requestmodels.WebhookSubscriptionCompletedRequest) (responsemodels.WebhookSubscriptionCompletedResponse, error) {
	query := `UPDATE user_subscriptions SET status=$1 WHERE razorpay_subscription_id=$2`
	result := ad.DB.Exec(query, req.Status, req.RazorpaySubscriptionId)
	if result.Error != nil {
		return responsemodels.WebhookSubscriptionCompletedResponse{}, result.Error
	}
	if result.RowsAffected == 0 {
		return responsemodels.WebhookSubscriptionCompletedResponse{}, gorm.ErrRecordNotFound
	}
	return responsemodels.WebhookSubscriptionCompletedResponse{
		RazorpaySubcriptionId: req.RazorpaySubscriptionId,
	}, nil
}

func (ad *AuthSubscriptionRepository) IsEligibleForSubsciption(req requestmodels.SubscribeRequest) (bool, error) {
	var num int64 // Use int64 for GORM counts

	// 1. Use parentheses to group the status logic
	// 2. Use single quotes for string literals in SQL
	query := `SELECT COUNT(*) FROM user_subscriptions 
              WHERE user_id = ? 
              AND (status = 'active' OR status = 'halted' OR status='created')`

	// Use .Row() or .Count() for cleaner code, but Raw is fine if preferred
	err := ad.DB.Raw(query, req.UserId).Scan(&num).Error
	if err != nil {
		return false, err
	}
	//fmt.Println("userid",req.UserId)
	//fmt.Println("req",req)
	//fmt.Println("num",num)
	// If count is 0, they are eligible
	return num == 0, nil
}

func (ad *AuthSubscriptionRepository) UpdateCount(req requestmodels.WebhookSubscriptionChargedRequest) error {
	query := `UPDATE user_subscriptions SET paid_count=$1,remaining_count=$2 WHERE razorpay_subscription_id=$3`
	result := ad.DB.Exec(query, req.PaidCount, req.RemainingCount, req.RazorpaySubscriptionId)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (ad *AuthSubscriptionRepository) FetchUserSubscription(razorpaySubId string) (string, error) {
	var status string
	query := `select status from user_subscriptions where razorpay_subscription_id=$1`
	result := ad.DB.Raw(query, razorpaySubId).Scan(&status)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return status, nil
}

func (ad *AuthSubscriptionRepository) DoesUserExists(userid uint64) (bool, error) {
	var num int64
	query := `select count(*) from users where id=$1`
	err := ad.DB.Raw(query, userid).Scan(&num).Error
	if err != nil {
		return false, err
	}
	return num != 0, nil
}

func (ad *AuthSubscriptionRepository) CheckAllUsersExists(userIDs []uint64) ([]uint64, error) {
	var existingIDs []uint64

	query := `SELECT id FROM users WHERE id = ANY($1)`
	err := ad.DB.Raw(query, pq.Array(userIDs)).Scan(&existingIDs).Error
	if err != nil {
		return nil, err
	}

	existingMap := make(map[uint64]struct{})
	for _, id := range existingIDs {
		existingMap[id] = struct{}{}
	}

	var missing []uint64
	for _, id := range userIDs {
		if _, ok := existingMap[id]; !ok {
			missing = append(missing, id)
		}
	}

	return missing, nil
}

func (ad *AuthSubscriptionRepository) FetchSubStatus(rsubid string) (string, error) {
	var status string
	query := `select status from user_subscriptions where razorpay_subscription_id=$1`
	result := ad.DB.Raw(query, rsubid).Scan(&status)
	if result.Error != nil {
		return "", result.Error
	}
	if result.RowsAffected == 0 {
		return "", gorm.ErrRecordNotFound
	}
	return status, nil
}
