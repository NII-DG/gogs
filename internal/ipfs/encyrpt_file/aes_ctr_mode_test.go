package encyrptfile_test

import (
	"testing"

	ef "github.com/NII-DG/gogs/internal/ipfs/encyrpt_file"
)

func TestEncrypted_1k(t *testing.T) {

	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

	address, err := ef.Encrypted("D:/Myrepository/testdata/gogs/1_1kbyte.txt", password)
	if err != nil {
		t.Logf("Fialure Encrypted(). Error : %v\n", err)
		t.Fail()
	}

	if len(address) == 0 {
		t.Fail()
	}
	t.Logf("Sucess TestEncrypted_1k(t *testing.T)\n")
}

var N = 100

var password = "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

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
