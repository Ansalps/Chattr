package domain

type Admin struct {
	ID        uint   `json:"id" gorm:"uniqueKey; not null"`
	Email     string `json:"email" gorm:"validate:required"`
	Password  string `json:"password" gorm:"validate:required"`
}

