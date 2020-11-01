package auth

import (
	"encoding/base64"
	. "github.com/franela/goblin"
	"testing"
	"time"
)

func TestSelfEncodeToken(t *testing.T) {
	g := Goblin(t)
	g.Describe("Encoding", func() {
		dateOk := SelfEncodeToken(testId, nil)
		g.It("Should return a signed Token", func() {
			id, valid := ValidateSelfEncoded(dateOk)
			g.Assert(valid).IsTrue()
			g.Assert(id.Id == testId).IsTrue()
		})
		g.Xit("Should not accept an expired Token", func() {
			id, valid := ValidateSelfEncoded(dateOk /*dateExpired*/)
			g.Assert(valid).IsFalse()
			g.Assert(id.Id == 0).IsTrue()
		})
		g.It("Should not accept an unsigned Token", func() {
			invalid, err := base64.StdEncoding.DecodeString(dateOk)
			g.Assert(err).IsNil()
			invalid[2]++
			id, valid := ValidateSelfEncoded(base64.StdEncoding.EncodeToString(invalid))
			g.Assert(id.Id == 0).IsTrue()
			g.Assert(valid).IsFalse()

			invalid = []byte(dateOk)
			invalid[2]--
			invalid[76]++
			id, valid = ValidateSelfEncoded(base64.StdEncoding.EncodeToString(invalid))
			g.Assert(id.Id == 0).IsTrue()
			g.Assert(valid).IsFalse()
		})
		g.It("Should not accept a Token with a wrong issue date", func() {
			id, valid := ValidateSelfEncoded(dateOk)
			g.Assert(valid).IsTrue()
			g.Assert(id.Id == 0).IsFalse()
			SetSelfEncodedTokenLifeTime(-10 * time.Second)
			id, valid = ValidateSelfEncoded(dateOk)
			g.Assert(valid).IsFalse()
			g.Assert(id.Id == 0).IsTrue()
			SetSelfEncodedTokenLifeTime(15 * time.Minute)
		})
		g.It("Should not accept a revoked Token", func() {
			id, valid := ValidateSelfEncoded(dateOk)
			g.Assert(valid).IsTrue()
			g.Assert(id.Id == 0).IsFalse()
			Revoke(id.Id)
			id, valid = ValidateSelfEncoded(dateOk)
			g.Assert(valid).IsFalse()
			g.Assert(id.Id == 0).IsTrue()
		})
	})
}

func BenchmarkSelfEncodeToken(b *testing.B) {
	for i := 0; i < b.N; i++ {
		SelfEncodeToken(testId, nil)
	}
}

func BenchmarkValidateSelfEncoded(b *testing.B) {
	tok := SelfEncodeToken(testId, nil)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		ValidateSelfEncoded(tok)
	}
}
