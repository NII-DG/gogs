package encyrptfile_test

//実行コマンド
//cd internal\ipfs\encyrpt_file
//全て実行
//go.exe test -benchmem -run=^$ -bench . -timeout 24h -count 3 -trace <fileNm>.trace -cpuprofile <fileNm>.prof -benchtime 100x

//go tool trace --http localhost:6060 a.trace

//go tool pprof -http :6060 a.prof

//一つのベンチマークテストのみ実行の場合
//go.exe test -benchmem -run=^$ -bench ^BenchmarkEncrypted_1k$ github.com/NII-DG/gogs/internal/ipfs/encyrpt_file -benchtime 100x -timeout 24h -count 6 -trace <fileNm>.trace -cpuprofile <fileNm>.prof

import (
	"testing"
	"time"

	"github.com/NII-DG/gogs/internal/ipfs"
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

var op = ipfs.IpfsOperation{
	Commander: ipfs.NewCommand(),
}

func bench(b *testing.B, filePath string, f func(string, string) (string, error)) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		address, err := f(filePath, password)
		if err != nil {
			b.Logf("Fialure Encrypted(). Error : %v\n", err)
		}
		//IPFSのGCを実行
		b.StopTimer()
		if err := op.PinRm(address); err != nil {
			b.Logf("Fialure PinRm(). Error : %v\n", err)
		}
		if err := op.RepoGc(); err != nil {
			b.Logf("Fialure RepoGc(). Error : %v\n", err)
		}
		b.StartTimer()
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
