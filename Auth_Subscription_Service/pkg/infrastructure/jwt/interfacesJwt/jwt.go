package interfacesJwt

import "time"

type Jwt interface {
	GenerateToken(securityKey string, UserId uint64, email, role string, tokenType string, duration time.Duration) (string, error)
}
