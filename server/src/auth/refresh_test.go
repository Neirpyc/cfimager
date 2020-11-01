package auth

import (
	"bytes"
	"encoding/base64"
	. "github.com/franela/goblin"
	"testing"
)

func TestRefreshToken(t *testing.T) {
	g := Goblin(t)
	g.Describe("Encoding", func() {
		dateOk, bytesDateOk := RefreshToken(testId)
		g.It("Should return a signed Token", func() {
			id, valid, dbToken := ValidateRefreshToken(dateOk)
			g.Assert(valid).IsTrue()
			g.Assert(id == testId).IsTrue()
			g.Assert(bytes.Equal(dbToken, bytesDateOk)).IsTrue()
		})
		g.It("Should not accept an unsigned Token", func() {
			invalid, err := base64.StdEncoding.DecodeString(dateOk)
			g.Assert(err).IsNil()
			invalid[2]++
			id, valid, dbToken := ValidateRefreshToken(base64.StdEncoding.EncodeToString(invalid))
			g.Assert(id == 0).IsTrue()
			g.Assert(valid).IsFalse()
			g.Assert(dbToken == nil).IsTrue()

			invalid = []byte(dateOk)
			invalid[2]--
			invalid[76]++
			id, valid, dbToken = ValidateRefreshToken(base64.StdEncoding.EncodeToString(invalid))
			g.Assert(id == 0).IsTrue()
			g.Assert(valid).IsFalse()
			g.Assert(dbToken == nil).IsTrue()
		})
	})
}

func BenchmarkRefreshToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		RefreshToken(testId)
	}
}

func BenchmarkValidateRefreshToken(b *testing.B) {
	tok, _ := RefreshToken(testId)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ValidateRefreshToken(tok)
	}
}
