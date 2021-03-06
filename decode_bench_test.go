package toml

import (
	"testing"
)

func BenchmarkUnmarshal(b *testing.B) {
	var v testStruct
	data := loadTestData("test.toml")
	b.ResetTimer()
	b.SetBytes(int64(len(data)))

	for i := 0; i < b.N; i++ {
		if err := Unmarshal(data, &v); err != nil {
			b.Fatal(err)
		}
	}
}
