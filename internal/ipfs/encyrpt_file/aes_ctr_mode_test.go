package encyrptfile_test

import (
	"testing"
	"time"

	ef "github.com/NII-DG/gogs/internal/ipfs/encyrpt_file"
)

var password = "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

func TestEncrypted_1k(t *testing.T) {
	filePath := "D:/Myrepository/testdata/gogs/2_10kbyte.txt"
	now := time.Now()
	address, err := ef.Encrypted(filePath, password)
	if err != nil {
		t.Logf("Fialure Encrypted(). Error : %v\n", err)
		t.Fail()
	}

	if len(address) == 0 {
		t.Fail()
	}
	since := time.Since(now).Nanoseconds()
	t.Logf("Sucess TestEncrypted_1k(t *testing.T). time[%v ns]\n", since)
}

var N = 100

func bench(b *testing.B, filePath string, f func(string, string) (string, error)) {
	for i := 0; i < N; i++ {
		_, err := f(filePath, password)
		if err != nil {
			b.Logf("Fialure Encrypted(). Error : %v\n", err)
		}
	}
}

func BenchmarkEncrypted_1k(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/1_1kbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10k(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/2_10kbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_100k(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/3_100kbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_1M(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/4_1Mbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10M(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/5_10Mbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_100M(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/6_100Mbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_1G(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/7_1Gbyte.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10G(b *testing.B) {
	bench(b, "D:/Myrepository/testdata/gogs/8_10Gbyte.txt", ef.Encrypted)
}
