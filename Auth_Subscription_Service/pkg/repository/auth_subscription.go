package repository

import (
	"fmt"
	"time"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/requestmodels"
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/models/responsemodels"

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

func (ad *AuthSubscriptionRepository) CheckAdminExistsByEmail(email string) (*domain.Admin, error) {
	var admin domain.Admin
	res := ad.DB.Where("email = ?", email).First(&admin)
	if res.Error != nil {
		return nil, res.Error
	}
	return &admin, nil
}
func (ad *AuthSubscriptionRepository)DeletePendingUser(email string)error{
	query:=`DELETE FROM users WHERE email=? and status='pending'`
	if err:=ad.DB.Exec(query,email).Error; err!=nil{
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository) CheckUserExistsByEmail(email string) (*domain.User, error) {
	var user domain.User
	res := ad.DB.Where("email=?", email).First(&user)
	if res.Error != nil {
		return nil, res.Error
	}
	return &user, nil
}

func (ad *AuthSubscriptionRepository) CheckUserExistsByUseraname(username string) (*domain.User, error) {
	var user domain.User
	res := ad.DB.Where("user_name=?", username).First(&user)
	if res.Error != nil {
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

func (ad *AuthSubscriptionRepository) CreateSubscriptionPlan(plan map[string]interface{}) (responsemodels.CreateSubscriptionPlanResponse, error) {
	// Extracting the root-level fields
	razorpayPlanId, ok := plan["id"].(string)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("id is missing or not a string")
	}

	period, ok := plan["period"].(string)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("period is missing or not a string")
	}

	interval, ok := plan["interval"].(float64)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("interval is missing or not a number")
	}

	// Extracting the nested item
	itemData, ok := plan["item"].(map[string]interface{})
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item is missing or not a map")
	}

	// Parse the nested fields inside `item`
	name, ok := itemData["name"].(string)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item name is missing or not a string")
	}

	amount, ok := itemData["amount"].(float64)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item amount is missing or not a number")
	}

	currency, ok := itemData["currency"].(string)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item currency is missing or not a string")
	}

	description, ok := itemData["description"].(string)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item description is missing or not a string")
	}

	isActive, ok := itemData["active"].(bool)
	if !ok {
		return responsemodels.CreateSubscriptionPlanResponse{}, fmt.Errorf("item active status is missing or not a boolean")
	}

	// Step 3 — Create the SubscriptionPlan struct
	subscriptionPlan := &domain.SubscriptionPlan{
		RazorpayPlanId: razorpayPlanId,
		Name:           name,
		Price:          int64(amount)/100, // Convert amount to int64
		Currency:       currency,
		Period:         period,
		Interval:       uint64(interval), // Convert interval to uint64
		Description:    description,
		IsActive:       isActive,
	}

	if err:=ad.DB.Create(&subscriptionPlan).Error;err!=nil{
		return responsemodels.CreateSubscriptionPlanResponse{},err
	}

	return responsemodels.CreateSubscriptionPlanResponse{
		ID: uint64(subscriptionPlan.ID),
		CreatedAt: subscriptionPlan.CreatedAt,
		UpdatedAt: subscriptionPlan.UpdatedAt,
		Name: subscriptionPlan.Name,
		Price: subscriptionPlan.Price,
		Currency: subscriptionPlan.Currency,
		Period: subscriptionPlan.Period,
		Interval: subscriptionPlan.Interval,
		Description: subscriptionPlan.Description,
		IsActive: subscriptionPlan.IsActive,
	}, nil
}

func (ad *AuthSubscriptionRepository) CreateSubscription(subscribeReq requestmodels.SubscribeRequest, subscription map[string]interface{}) (responsemodels.SubscribeResponse, error) {
	fmt.Println("hi, is it entering in create Subscription")

	// Extracting the root-level fields
	razorpaySubscriptionId, ok := subscription["id"].(string)
	if !ok {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("id is missing or not a string")
	}

	razorpayPlanId, ok := subscription["plan_id"].(string)
	if !ok {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("plan_id is missing or not a string")
	}

	status := "pending"

	// Numbers are float64 in the subscription map
	totalCountF, ok := subscription["total_count"].(float64)
	if !ok {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("total_count is missing or not a number")
	}
	totalCount := int(totalCountF)

	remainingCountF, ok := subscription["remaining_count"].(float64)
	if !ok {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("remaining_count is missing or not a number")
	}
	remainingCount := int(remainingCountF)

	paidCountF, ok := subscription["paid_count"].(float64)
	if !ok {
		return responsemodels.SubscribeResponse{}, fmt.Errorf("paid_count is missing or not a number")
	}
	paidCount := int(paidCountF)

	// Step 3 — Create the UserSubscription struct
	userSubscription := &domain.UserSubscription{
		UserID:                  subscribeReq.UserId,
		RazorpaySubscriptionId:  razorpaySubscriptionId,
		RazorpayPlanId:          razorpayPlanId,
		Status:                  status,
		TotalCount:              totalCount,
		PaidCount:               paidCount,
		RemainingCount:          remainingCount,
	}

	if err := ad.DB.Create(&userSubscription).Error; err != nil {
		return responsemodels.SubscribeResponse{}, err
	}

	return responsemodels.SubscribeResponse{
		ID:                     uint64(userSubscription.ID),
		CreatedAt:              userSubscription.CreatedAt,
		UpdatedAt:              userSubscription.UpdatedAt,
		UserID:                 userSubscription.UserID,
		RazorpaySubscriptionId: userSubscription.RazorpaySubscriptionId,
		Status:                 userSubscription.Status,
		StartAt:                userSubscription.StartAt,
		EndAt:                  userSubscription.EndAt,
		NextChargeAt:           userSubscription.NextChargeAt,
		TotalCount:             userSubscription.TotalCount,
		RemainingCount:         userSubscription.RemainingCount,
		PaidCount:              userSubscription.PaidCount,
		CancelledAt:            userSubscription.CancelledAt,
		CancelReason:           userSubscription.CancelReason,
	}, nil
}




func (ad *AuthSubscriptionRepository) FetchStatusFromSubcriptionPlan(id uint64)(bool,error){
	var status bool
	query:=`SELECT is_active FROM subscription_plans WHERE id=?`
	if err:=ad.DB.Raw(query,id).Scan(&status).Error; err!=nil{
		return false,err
	}
	return status,nil
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

func (ad *AuthSubscriptionRepository)FetchRazorpayPlanIdFromId(id uint64)(string,error){
	var RazorpayPlanId string
	query:=`SELECT razorpay_plan_id FROM subscription_plans WHERE id=?`
	if err:=ad.DB.Raw(query,id).Scan(&RazorpayPlanId).Error; err!=nil{
		return "",err
	}
	return RazorpayPlanId,nil
}

func (ad *AuthSubscriptionRepository)UpdateUserSubscripion(id string,subscription map[string]interface{})(responsemodels.VerifySubscriptionPaymentResponse,error){
	startAtUnix,ok := subscription["start_at"].(float64)
	if !ok{
		return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("current_start is probably nil")
	}
	endAtUnix,ok := subscription["end_at"].(float64)
	if !ok{
		return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("current_end is probably nil")
	}
	nextChargeAtUnix,ok:=subscription["charge_at"]
	if !ok{
		return responsemodels.VerifySubscriptionPaymentResponse{},fmt.Errorf("charge_at is probably nil")
	}

startAt := time.Unix(int64(startAtUnix), 0)
endAt   := time.Unix(int64(endAtUnix), 0)
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
fmt.Println("start at ***",startAt,"endAt **",endAt,"nextChargeAt",nextChargeAt)
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
	query:=`SELECT * FROM user_subscriptions where razorpay_subscription_id=?`
    if err := ad.DB.Raw(query, id).Scan(&userSubscription).Error; err != nil {
        return responsemodels.VerifySubscriptionPaymentResponse{}, err
    }
	return responsemodels.VerifySubscriptionPaymentResponse{
		ID: uint64(userSubscription.ID),
		CreatedAt: userSubscription.CreatedAt,
		UpdatedAt: userSubscription.UpdatedAt,
		UserID: userSubscription.UserID,
		RazorpaySubscriptionId: userSubscription.RazorpaySubscriptionId,
		Status: userSubscription.Status,
		StartAt: userSubscription.StartAt,
		EndAt: userSubscription.EndAt,
		NextChargeAt: userSubscription.NextChargeAt,
		TotalCount: userSubscription.TotalCount,
		RemainingCount: userSubscription.RemainingCount,
		PaidCount: userSubscription.PaidCount,
	},nil
}

type SubPlan struct {
    Price    int64
    Currency string
}

func (ad *AuthSubscriptionRepository)FetchAmountCurrencyFromSubscriptionPlan(id uint64)(int64,string,error){
	var plan SubPlan
	query:=`SELECT price,currency FROM subscription_plans WHERE id=?`
	if err:=ad.DB.Raw(query,id).Scan(&plan).Error; err!=nil{
		return 0,"",err
	}
	return plan.Price,plan.Currency,nil
}

func (ad *AuthSubscriptionRepository)FetchRazorpaySubscriptionIdFromSubcriptionId(subid uint64)(string,error){
	var razorpaySubscriptionId string
	query:=`SELECT razorpay_subscription_id FROM user_subscriptions WHERE id=?`
	if err:=ad.DB.Raw(query,subid).Scan(&razorpaySubscriptionId).Error; err!=nil{
		return "",nil
	}
	return razorpaySubscriptionId,nil
}

func (ad *AuthSubscriptionRepository)ChangeUserSubscriptionStatusToCancelled(subid uint64,res map[string]interface{})(responsemodels.UnsubscribeResponse,error){
	// Extracting the root-level fields
	status, ok := res["status"].(string)
	if !ok {
		return responsemodels.UnsubscribeResponse{}, fmt.Errorf("status is missing or not a string")
	}
	cancelledAt:=time.Now()
	query:=`UPDATE user_subscriptions SET status=?,cancelled_at=? WHERE id=?`
	if err:=ad.DB.Exec(query,status,cancelledAt,subid).Error; err!=nil{
		return responsemodels.UnsubscribeResponse{},err
	}
	var userSubscription responsemodels.UnsubscribeResponse
	query1:=`SELECT * FROM user_subscriptions WHERE id=?`
	if err:=ad.DB.Raw(query1,subid).Scan(&userSubscription).Error; err!=nil{
		return responsemodels.UnsubscribeResponse{},err
	}
	fmt.Println("printing user id",userSubscription.UserID)
	return userSubscription,nil
}

func (ad *AuthSubscriptionRepository)FetchUserIdFromSubscriptionId(razorpaySubId string)(uint64,error){
	var userid uint64
	query:=`SELECT user_id from user_subscriptions WHERE razorpay_subscription_id=?`
	if err:=ad.DB.Raw(query,razorpaySubId).Scan(&userid).Error; err!=nil{
		return 0,err
	}
	return userid,nil
}

func (ad *AuthSubscriptionRepository)TurnBlueTickTrueForUserId(userid uint64)error{
	query:=`UPDATE users SET blue_tick=true where id=?` 
	if err:=ad.DB.Exec(query,userid).Error; err!=nil{
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository)PopulatePayment(payment map[string]interface{},verifySubscripitionPaymentReq requestmodels.VerifySubscriptionPaymentRequest)(domain.Payment,error){
	//var payment domain.Payment
	razorpayPaymentId:=verifySubscripitionPaymentReq.RazorpayPaymentId
	razorpaySubscriptionId:=verifySubscripitionPaymentReq.RazorpaySubscriptionId
	// Extracting the root-level fields
	razorpayInvoiceId, ok := payment["invoice_id"].(string)
	if !ok {
		return domain.Payment{}, fmt.Errorf("invoice id is missing or not a string")
	}

	amount, ok := payment["amount"].(float64)
	if !ok {
		return domain.Payment{}, fmt.Errorf("amount is missing or not a float64")
	}

	// Numbers are float64 in the subscription map
	currency, ok := payment["currency"].(string)
	if !ok {
		return domain.Payment{}, fmt.Errorf("currency is missing or not a string")
	}

	paymentStatus, ok := payment["status"].(string)
	if !ok {
		return domain.Payment{}, fmt.Errorf("status is missing or not a string")
	}
	

	paymentMethod, ok := payment["method"].(string)
	if !ok {
		return domain.Payment{}, fmt.Errorf("method is missing or not a string")
	}

	createdAtUnix, ok := payment["created_at"].(float64)  // `created_at` is usually a float64 in JSON
if !ok {
    return domain.Payment{}, fmt.Errorf("created_at is missing or not a number")
}

// Convert Unix timestamp to time.Time
createdAt := time.Unix(int64(createdAtUnix), 0)  // Unix timestamp is in seconds, so we use 0 for nanoseconds

	// Step 3 — Create the UserSubscription struct
	paymentCreate := &domain.Payment{
		RazorpaySubscriptionId: razorpaySubscriptionId,
		RazorpayPaymentId: razorpayPaymentId,
		RazorpayInvoiceId: razorpayInvoiceId,
		Amount: amount/100,
		Currency: currency,
		PaymentStatus: paymentStatus,
		PaymentMethod: paymentMethod,
		TransactionDate: createdAt,
	}

	if err := ad.DB.Create(&paymentCreate).Error; err != nil {
		return domain.Payment{}, err
	}

	return *paymentCreate, nil
}

func (ad *AuthSubscriptionRepository) FetchRazorpayPlanIdFromRazrorpaySubscriptionId(subId string)(string,error){
	var palnId string
	query:=`SELECT razrorpay_plan_id FROM user_subscriptions where razrorpay_subscription_id=?`
	if err:=ad.DB.Raw(query,subId).Scan(&palnId).Error; err!=nil{
		return "",err
	}
	return palnId,nil
}
type periodInterval struct{
	Period string
	Interval uint64
}
func (ad *AuthSubscriptionRepository)FetchIntervalPeriodFromSubscriptionPlan(planId string)(string,uint64,error){
	var p periodInterval
	query:=`SELECT period,interval FROM subscription_plans where razrorpay_plan_id=?`
	if err:=ad.DB.Raw(query,planId).Scan(&p).Error; err!=nil{
		return "",0,err
	}
	return p.Period,p.Interval,nil
}

func (ad *AuthSubscriptionRepository)FetchTotalCountFromUserSubscription(subId string)(int,error){
	var totalCount int
	query:=`SELECT total_count FROM user_subscriptions WHERE razorpay_subscription_id=?`
	if err:=ad.DB.Raw(query,subId).Scan(&totalCount).Error; err!=nil{
		return 0,err
	}
	return totalCount,nil
}

func (ad *AuthSubscriptionRepository)UpdateTimeUserSubscription(startAt,endAt,nextChargeAt time.Time,subid string)(responsemodels.VerifySubscriptionPaymentResponse,error){
	var subscribe responsemodels.VerifySubscriptionPaymentResponse
	updated_at:=time.Now()
	query:=`UPDATE user_subscriptions SET updated_at=?,start_at=?,end_at=?,next_charge_at=? WHERE razorpay_subscription_id=?`
	if err:=ad.DB.Exec(query,updated_at,startAt,endAt,nextChargeAt,subid).Error; err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	query1:=`SELECT * FROM user_subscriptions WHERE razropay_subscription_id=?`
	if err:=ad.DB.Raw(query1,subid).Scan(&subscribe).Error; err!=nil{
		return responsemodels.VerifySubscriptionPaymentResponse{},err
	}
	return subscribe,nil
}

func (ad *AuthSubscriptionRepository)FetchNextChargeAtFromUserSubcription(subid string)(time.Time,error){
	var nextChargeAt time.Time
	query:=`SELECT next_charge_at from user_subscriptions WHERE razorpay_subscription_id=?`
	if err:=ad.DB.Raw(query,subid).Scan(&nextChargeAt).Error; err!=nil{
		return time.Time{},err
	}
	return nextChargeAt,nil
}

func (ad *AuthSubscriptionRepository)TurnOffBlueTickForUserId(userid uint64)error{
	query:=`UPDATE users SET blue_tick=false WHERE id=?`
	if err:=ad.DB.Exec(query,userid).Error; err!=nil{
		return err
	}
	return nil
}

func (ad *AuthSubscriptionRepository)UpdateProfileImage(userid uint64,imageUrl string)error{
	query:=`UPDATE users SET profile_img_url=? WHERE id=?`
	if err:=ad.DB.Exec(query,imageUrl,userid).Error; err!=nil{
		return err
	}
	return nil
}