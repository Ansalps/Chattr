package domain

import "gorm.io/gorm"

type postStatus string

const(
	Normal postStatus="normal"
	Archived postStatus="archived"
)
type Post struct{
	gorm.Model
	UserID uint	`gorm:"not null"`
	Caption string
	PostStatus postStatus `gorm:"default:normal"`
}

type PostMedia struct{
	gorm.Model
	PostID uint `gorm:"not null; index"`
	Post Post `gorm:"constraint:OnDelete:CASCADE"`
	MediaUrl string
}