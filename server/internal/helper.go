package connection

import (
	"math/rand"
	"strconv"
)

func generateUserId() string {
	return strconv.Itoa(rand.Intn(10000000))
}
