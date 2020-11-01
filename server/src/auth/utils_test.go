package auth

import (
	. "github.com/franela/goblin"
	"os"
	"testing"
)

const (
	testId = 0x0f0f0f0f0f0f0f0f
)

func init() {
	if err := os.Setenv("AUTH_SECRET_SALT", "MXd4b1ltY0hYUjJZQXp4SjR3YXEzbTBEZWhxUG5uVHQwU0tvek1nWGhvUk0zd3BHR1VlR3FiUWNi ZkNWSDJSeQ=="); err != nil {
		panic(err)
	}
}

func TestRandomData(t *testing.T) {
	g := Goblin(t)
	const testCount = 500
	g.Describe("Random data generation", func() {
		g.It("Should generate 40 bytes values", func() {
			for i := 0; i < testCount; i++ {
				g.Assert(len(<-randData40) == 40).IsTrue()
			}
		})
		g.It("Should generate 64 bytes values", func() {
			for i := 0; i < testCount; i++ {
				g.Assert(len(<-randData64) == 64).IsTrue()
			}
		})
	})
}

func TestHashPassword(t *testing.T) {
	g := Goblin(t)
	g.Describe("Hash password", func() {
		g.It("Should return correctly sized digest", func() {
			g.Assert(len(HashPassword([]byte("password"), []byte("salt"))) == hashLength).IsTrue()
		})
	})
}

func BenchmarkHashPassword(b *testing.B) {
	for i := 0; i < b.N; i++ {
		HashPassword([]byte("password"), []byte("salt"))
	}
}
