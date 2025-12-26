package interfacesrepository

import "time"

type RedisRepository interface {
	BlacklistToken(jti string, exp time.Time) error
	IsTokenBlacklisted(jti string) (bool, error)
}
