package main

import (
	"crypto/rand"
	"errors"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"golang.org/x/crypto/argon2"
	"math/big"
	"time"
)

const (
	randDataSize = 64
)

var (
	randData chan []byte
)

type MailerRequest struct {
	Email forms.Email
	Token []byte
}

func init() {
	randData = make(chan []byte, 10)
	go generateRandomData()
}

func generateRandomData() {
	max := big.NewInt(256)
	var result *big.Int
	var err error
	max.Exp(big.NewInt(256), big.NewInt(randDataSize), nil)
	for {
		err = errors.New("")
		for err != nil {
			result, err = rand.Int(rand.Reader, max)
			time.Sleep(1 * time.Millisecond)
		}
		randData <- result.Bytes()
	}
}

func hashPassword(password []byte, salt []byte) (hash []byte) {
	return argon2.IDKey(password, salt, 4, 128*1024, 8, 64)
}
