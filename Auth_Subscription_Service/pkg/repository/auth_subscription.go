package repository

import (
	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/domain"
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

func (ad *AuthSubscriptionRepository) CheckAdminExistsByEmail(email string) (domain.Admin, error) {
	var admin domain.Admin
	res := ad.DB.Where("email = ?", email).First(&admin)
	if res.Error != nil {
		return domain.Admin{}, res.Error
	}
	return admin, nil
}
