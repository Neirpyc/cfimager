package main

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"github.com/Neirpyc/cfimager/server/src/forms"
	"golang.org/x/crypto/argon2"
	"math/big"
	"regexp"
	"time"
)

const (
	randDataSize = 64
)

var (
	randData    chan []byte
	emailRegexp *regexp.Regexp
)

type MailerRequest struct {
	Email forms.Email
	Token []byte
}

type CodeSuccessError struct {
	Code         int
	SuccessError SuccessError
}

func init() {
	randData = make(chan []byte, 10)
	go generateRandomData()
	emailRegexp = regexp.MustCompile("(?:[a-z0-9!#$%&'*+/=?^_`{|}~-]+(?:\\.[a-z0-9!#$%&'*+/=?^_`{|}~-]+)*|\"(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21\\x23-\\x5b\\x5d-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])*\")@(?:(?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\\.)+[a-z0-9](?:[a-z0-9-]*[a-z0-9])?|\\[(?:(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?)\\.){3}(?:25[0-5]|2[0-4][0-9]|[01]?[0-9][0-9]?|[a-z0-9-]*[a-z0-9]:(?:[\\x01-\\x08\\x0b\\x0c\\x0e-\\x1f\\x21-\\x5a\\x53-\\x7f]|\\\\[\\x01-\\x09\\x0b\\x0c\\x0e-\\x7f])+)\\])")
}

func generateRandomData() {
	max := big.NewInt(256)
	var result *big.Int
	var err error
	max.Exp(big.NewInt(256), big.NewInt(randDataSize), nil)
	for true {
		err = errors.New("")
		for err != nil {
			result, err = rand.Int(rand.Reader, max)
			time.Sleep(1 * time.Millisecond)
		}
		randData <- result.Bytes()
	}
}

func base64ToBytes(b64 []byte) (result []byte) {
	_, _ = base64.StdEncoding.Decode(result, b64)
	return
}

func xorBytes(dst []byte, src []byte) {
	ln := len(dst)
	if ln < len(src) {
		L.Infof("Xoring bytes with unmatched length %d and %d\n", ln, len(src))
		ln = len(src)
	}
	for i := 0; i < ln; i++ {
		dst[i] ^= src[i]
	}
}

func hashPassword(password []byte, salt []byte) (hash []byte) {
	return argon2.IDKey(password, salt, 4, 128*1024, 8, 64)
}
