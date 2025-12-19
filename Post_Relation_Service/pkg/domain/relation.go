package domain

import "time"

type RelationType string

const (
	RelationFollows RelationType = "follows"
	RelationBlocked RelationType = "blocked"
)

type Relation struct {
	ID uint `gorm:"primaryKey"`

	FollowerID  uint `gorm:"not null;uniqueIndex:idx_follower_following"`
	FollowingID uint `gorm:"not null;uniqueIndex:idx_follower_following"`

	RelationType RelationType `gorm:"type:varchar(20);not null;default:'follows'"`

	CreatedAt time.Time
	UpdatedAt time.Time
}
