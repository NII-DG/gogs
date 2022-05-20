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
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/NII-DG/gogs/internal/ipfs"
	ef "github.com/NII-DG/gogs/internal/ipfs/encyrpt_file"
)

var password = "cekYSYu3cTQL3yiKFoEwTWC4YATazRcL"

var testDataDir = "D:/Myrepository/testdata/gogs/"
var tmpDir = "D:/Myrepository/testdata/gogs/tmp/"

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

func benchEncrypt(b *testing.B, fileNm string, f func(string, string) (string, error)) {
	filePath := filepath.Join(testDataDir, fileNm)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		address, err := f(filePath, password)
		if err != nil {
			b.Logf("Fialure Encrypted(). Error : %v\n", err)
		} else {
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
}

func BenchmarkEncrypted_1k(b *testing.B) {
	benchEncrypt(b, "1K.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10k(b *testing.B) {
	benchEncrypt(b, "10K.txt", ef.Encrypted)
}

func BenchmarkEncrypted_100k(b *testing.B) {
	benchEncrypt(b, "100K.txt", ef.Encrypted)
}

func BenchmarkEncrypted_1M(b *testing.B) {
	benchEncrypt(b, "1M.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10M(b *testing.B) {
	benchEncrypt(b, "10M.txt", ef.Encrypted)
}

func BenchmarkEncrypted_100M(b *testing.B) {
	benchEncrypt(b, "100M.txt", ef.Encrypted)
}

func BenchmarkEncrypted_1G(b *testing.B) {
	benchEncrypt(b, "1G.txt", ef.Encrypted)
}

func BenchmarkEncrypted_10G(b *testing.B) {
	benchEncrypt(b, "10G.txt", ef.Encrypted)
}

//以下、Decrypt()のベンチマークテストコード

//実行コマンド
//cd internal\ipfs\encyrpt_file
//全て実行
//go.exe test -benchmem -run=^$ -bench . -timeout 24h -count 3 -trace <fileNm>.trace -cpuprofile <fileNm>.prof -benchtime 100x

//go tool trace --http localhost:6060 a.trace

//go tool pprof -http :6060 a.prof

//一つのベンチマークテストのみ実行の場合（例）
//go.exe test -benchmem -run=^$ -bench ^BenchmarkEncrypted_1k$ github.com/NII-DG/gogs/internal/ipfs/encyrpt_file -benchtime 100x -timeout 24h -count 6 -trace <fileNm>.trace -cpuprofile <fileNm>.prof

func benchDecrypt(b *testing.B, testFileNm string, f func(string, string, string) error) {
	testfilePath := filepath.Join(testDataDir, testFileNm)
	outputPath := filepath.Join(tmpDir, testFileNm)
	address, err := ef.Encrypted(testfilePath, password)
	if err != nil {
		b.Logf("Fialure Encrypted(). Error : %v\n", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := f(address, password, outputPath)
		if err != nil {
			b.Logf("Fialure Decrypted(). Error : %v\n", err)
		}
		b.StopTimer()
		os.Remove(outputPath)
		b.StartTimer()
	}
	b.StopTimer()
	if err := op.PinRm(address); err != nil {
		b.Logf("Fialure PinRm(). Error : %v\n", err)
	}
	if err := op.RepoGc(); err != nil {
		b.Logf("Fialure RepoGc(). Error : %v\n", err)
	}
	b.StartTimer()
}

func BenchmarkDecrypted_1k(b *testing.B) {
	testFileNm := "1K.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_10k(b *testing.B) {
	testFileNm := "10K.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_100k(b *testing.B) {
	testFileNm := "100K.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_1M(b *testing.B) {
	testFileNm := "1M.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_10M(b *testing.B) {
	testFileNm := "10M.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_100M(b *testing.B) {
	testFileNm := "100M.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_1G(b *testing.B) {
	testFileNm := "1G.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}

func BenchmarkDecrypted_10G(b *testing.B) {
	testFileNm := "10G.txt"
	benchDecrypt(b, testFileNm, ef.Decrypted)
}
