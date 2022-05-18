package encyrptfile_test

import (
	"testing"

	ef "github.com/NII-DG/gogs/internal/ipfs/encyrpt_file"
	log "unknwon.dev/clog/v2"
)

func TestEncrypted_1k(t *testing.T) {

	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

	address, err := ef.Encrypted("D:/Myrepository/testdata/gogs/1_1kbyte.txt", password)
	if err != nil {
		log.Error("Fialure Encrypted(). Error : %v\n", err)
		t.Fail()
	}

	if len(address) == 0 {
		t.Fail()
	}
	log.Info("Sucess TestEncrypted_1k(t *testing.T)\n")
}

var N = 9

var result string

func BenchmarkEncrypted_1k(b *testing.B) {

	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"
	b.ResetTimer()
	var address string
	var err error
	for i := 0; i < b.N; i++ {
		address, err = ef.Encrypted("D:/Myrepository/testdata/gogs/1_1kbyte.txt", password)
		if err != nil {
			log.Error("Fialure Encrypted(). Error : %v\n", err)
		}
	}

	result = address
}

// func BenchmarkEncrypted_10k(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/2_10kbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_100k(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/3_100kbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_1M(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/4_1Mbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_10M(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/5_10Mbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_100M(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/6_100Mbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_1G(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/7_1Gbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }

// func BenchmarkEncrypted_10G(b *testing.B) {

// 	password := "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

// 	for i := 0; i < b.N; i++ {
// 		_, err := ef.Encrypted("D:/Myrepository/testdata/gogs/8_10Gbyte.txt", password)
// 		if err != nil {
// 			log.Error("Fialure Encrypted(). Error : %v\n", err)
// 		}
// 	}
// }
