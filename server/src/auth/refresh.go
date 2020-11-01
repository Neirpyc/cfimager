package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/binary"
	"encoding/gob"
	"time"
)

func RefreshToken(id uint64) (token string, dbPart []byte) {
	array := <-randData64
	binary.LittleEndian.PutUint64(array, uint64(time.Now().UnixNano()))
	binary.LittleEndian.PutUint64(array[8:], uint64(time.Now().Unix()))

	tok := Token{
		Issued:   time.Time{},
		Expire:   time.Time{},
		Id:       id,
		Metadata: array,
	}
	buf := bytes.NewBuffer(nil)

	if err := gob.NewEncoder(buf).Encode(tok.sign(signatureSha512)); err != nil {
		return
	}
	return base64.StdEncoding.EncodeToString(buf.Bytes()), array
}

func ValidateRefreshToken(token string) (id uint64, valid bool, dbPart []byte) {
	data, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return
	}
	var sig signed
	if err = gob.NewDecoder(bytes.NewReader(data)).Decode(&sig); err != nil {
		return
	}

	if bytes.Equal(sig.Signature, sig.Token.sign(signatureSha512).Signature) {
		var ok bool
		if dbPart, ok = sig.Token.Metadata.([]byte); !ok {
			return 0, false, nil
		}
		return sig.Token.Id, true, dbPart
	}
	return
}
