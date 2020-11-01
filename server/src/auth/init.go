package auth

import (
	"os"
	"time"
)

var (
	secretSalt               []byte
	semiSecretSalt           []byte
	randData64, randData40   chan []byte
	selfEncodedTokenLifeTime time.Duration = 15 * time.Minute
)

const (
	hashTime       = 4
	hashMemory     = 1024 * 128
	hashThreads    = 4
	hashLength     = 64
	randBufferSize = 10
)

func init() {
	setSecretSalt(os.Getenv("AUTH_SECRET_SALT"))
	_ = os.Setenv("AUTH_SECRET_SALT", "42")
	generateRandomData(randBufferSize)
}
