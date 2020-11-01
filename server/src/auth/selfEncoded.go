package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"sync"
	"time"
)

var (
	revoked sync.Map
)

func init() {
	go func() {
		t := time.NewTicker(1 * time.Minute)
		for {
			<-t.C
			revoked.Range(func(id, timeStamp interface{}) bool {
				if time.Since(time.Unix(timeStamp.(int64), 0)) <= 0 {
					revoked.Delete(id)
				}
				return true
			})
		}
	}()
}

func Revoke(userId uint64) {
	revoked.Store(userId, time.Now().Unix()+1)
}

func SetSelfEncodedTokenLifeTime(t time.Duration) {
	selfEncodedTokenLifeTime = t
}

func SelfEncodeToken(id uint64, i interface{}) string {
	tok := Token{
		Issued:   time.Now(),
		Expire:   time.Now().Add(selfEncodedTokenLifeTime),
		Id:       id,
		Metadata: i,
	}
	buf := bytes.NewBuffer(nil)
	if err := gob.NewEncoder(buf).Encode(tok.sign(signatureArgon2id)); err != nil {
		return ""
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes())
}

func ValidateSelfEncoded(token string) (t Token, v bool) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}
	var sig signed
	if err = gob.NewDecoder(bytes.NewReader(data)).Decode(&sig); err != nil {
		return
	}
	if bytes.Equal(sig.Signature, sig.Token.sign(signatureArgon2id).Signature) {
		if sig.Token.Expire.Sub(time.Now()).Seconds() <= 0 {
			return
		}
		if time.Now().Sub(sig.Token.Issued) >= selfEncodedTokenLifeTime {
			return
		}
		if value, ok := revoked.Load(sig.Token.Id); ok {
			if sig.Token.Issued.Sub(time.Unix(value.(int64), 0)) <= 0 {
				return
			}
		}
		return sig.Token, true
	}
	return
}
