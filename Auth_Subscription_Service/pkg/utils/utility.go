package utils

func (rn RandomNumber) RandomNumber() int {
	randomInt := rand.Intn(9000) + 1000
	return randomInt
}