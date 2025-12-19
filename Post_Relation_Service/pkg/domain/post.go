package domain

import (
	"time"
)

type postStatus string

const (
	Normal   postStatus = "normal"
	Archived postStatus = "archived"
)

type Post struct {
	ID         uint `gorm:"primarykey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	UserID     uint `gorm:"not null"`
	Caption    string
	PostStatus postStatus `gorm:"default:normal"`
}

type PostMedia struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	PostID    uint `gorm:"not null; index"`
	Post      Post `gorm:"constraint:OnDelete:CASCADE"`
	MediaUrl  string
}

type PostLike struct {
	UserID uint `gorm:"primaryKey"`
	PostID uint `gorm:"primaryKey"`

	CreatedAt time.Time

	Post Post `gorm:"foreignKey:PostID"`
}

type Comment struct{
	ID uint `gorm:"primarykey"`
	CreatedAt time.Time 
	UpdatedAt time.Time 
	UserID uint `gorm:"not null;index"`
	PostID uint `gorm:"not null;index"`
	Post  Post `gorm:"foreignkey:PostID;OnDelete:CASCADE"`
	CommentText string `gorm:"type:text;not null"`
	ParentCommentID *uint
	ParentComment *Comment `gorm:"foreignkey:ParentCommentID"`
}
