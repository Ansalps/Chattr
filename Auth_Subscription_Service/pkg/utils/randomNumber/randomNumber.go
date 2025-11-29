package randomnumber

import (
	"math/rand"

	"github.com/Ansalps/Chattr_Auth_Subscription_Service/pkg/utils/randomNumber/interfacesRandomNumber"
)

type RandomNumber struct{}

func NewRandomNumberUtil() interfacesRandomNumber.RandomNumber {
	return &RandomNumber{}
}

func (rn RandomNumber) RandomNumber() int {
	randomInt := rand.Intn(9000) + 1000
	return randomInt
}
