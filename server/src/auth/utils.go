package auth

import (
	"bytes"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"encoding/gob"
	"errors"
	"golang.org/x/crypto/argon2"
	"math/big"
	"time"
)

const (
	signatureArgon2id = iota
	signatureSha512
)

type signature int

type Token struct {
	Issued   time.Time
	Expire   time.Time
	Id       uint64
	Metadata interface{}
}

func (t *Token) sign(s signature) signed {
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(*t); err != nil {
		return signed{}
	}
	sig := signed{
		Token:     *t,
		Signature: nil,
	}
	switch s {
	case signatureSha512:
		s := sha512.Sum512(append(buf.Bytes(), semiSecretSalt...))
		sig.Signature = s[:]
	default:
		sig.Signature = argon2.IDKey(buf.Bytes(), secretSalt, 1, 1024, 1, 64)
	}
	return sig
}

type signed struct {
	Token     Token
	Signature []byte
}

func setSecretSalt(salt string) {
	if saltBytes, err := base64.StdEncoding.DecodeString(salt); err != nil {
		panic(err)
	} else {
		secretSalt = saltBytes
	}
	semiSecretSalt = argon2.IDKey(secretSalt, nil, 1, 1024*512, 8, 64)
}

func generateRandomData(bufferSize int) {
	var result *big.Int
	var err error
	if randData64 == nil {
		randData64 = make(chan []byte, bufferSize)
	}
	if randData40 == nil {
		randData40 = make(chan []byte, bufferSize)
	}
	max64 := big.NewInt(255)
	max64.Exp(big.NewInt(255), big.NewInt(64), nil)
	max40 := big.NewInt(255)
	max40.Exp(big.NewInt(255), big.NewInt(40), nil)
	empty := make([]byte, 64)
	go func() {
		for true {
			err = errors.New("")
			for err != nil {
				result, err = rand.Int(rand.Reader, max40)
			}
			data := result.Bytes()
			if len(data) > 40 {
				data = data[:40]
			}
			randData40 <- append(empty[:40-len(data)], data...)
		}
	}()
	go func() {
		for true {
			err = errors.New("")
			for err != nil {
				result, err = rand.Int(rand.Reader, max64)
			}
			data := result.Bytes()
			if len(data) > 64 {
				data = data[:64]
			}
			randData64 <- append(empty[:64-len(data)], data...)
		}
	}()
}

func HashPassword(password []byte, salt []byte) []byte {
	return argon2.IDKey(password, salt, hashTime, hashMemory, hashThreads, hashLength)
}

func Salt() []byte {
	return <-randData64
}
