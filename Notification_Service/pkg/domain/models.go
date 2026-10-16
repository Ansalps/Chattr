package domain
import (
	"time"
)

type Notification struct {
	ID         uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint64    `gorm:"index" json:"user_id"`      // The Recipient (Owner)
	ActorID    uint64    `json:"actor_id"`                 // The person who triggered it
	TargetID   string    `json:"target_id"`                // The Post ID, Comment ID, etc.
	Type       string    `gorm:"size:50" json:"type"`      // "POST_LIKE", "POST_COMMENT", "FOLLOW"
	Message    string    `gorm:"type:text" json:"message"` // The formatted text: "Ansal liked your post"
	//IsRead     bool      `gorm:"default:false" json:"is_read"`
	CreatedAt  time.Time `json:"created_at"`
}
type PaginationDetails struct{
	CurrentPage int
	PageSize int
}
type NotificationResponse struct{
	Notifications []Notification
	Pagination PaginationDetails
}