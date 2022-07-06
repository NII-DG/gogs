package db_test

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/G-Node/libgin/libgin/annex"
	"github.com/NII-DG/gogs/internal/annex_ipfs"
	"github.com/NII-DG/gogs/internal/ipfs"
	"github.com/gogs/git-module"
)

var testDataDir = "D:/Myrepository/testdata/gogs/"
var gitDir = "D:/Myrepository/testdata/gogs/git"

var op = ipfs.IpfsOperation{
	Commander: ipfs.NewCommand(),
}

func benchAnnexCopyToIPFS(b *testing.B, fileNm string) {

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		//前処理
		err := os.MkdirAll(gitDir, 0777)
		if err != nil {
			b.Logf("Fialure MkdirAll(). Error : %v\n", err)
		}
		_, err = git.NewCommand("init").RunInDir(gitDir)
		if err != nil {
			b.Logf("Fialure git init. Error : %v\n", err)
		}
		if _, err := annex.Init(gitDir); err != nil {
			b.Logf("Fialure git annex init. Error : %v\n", err)
		}
		if _, err := git.NewCommand("annex", "initremote", "ipfs", "type=external", "externaltype=ipfs", "encryption=none").RunInDir(gitDir); err != nil {
			b.Logf("Fialure git annex ON IPFS. Error : %v\n", err)
		}

		srcName := filepath.Join(testDataDir, fileNm)
		src, err := os.Open(srcName)
		if err != nil {
			b.Logf("Fialure Open File. Error : %v\n", err)
		}
		defer src.Close()

		dstName := filepath.Join(gitDir, fileNm)
		dst, err := os.Create(dstName)
		if err != nil {
			b.Logf("Fialure Create File. Error : %v\n", err)
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			b.Logf("Fialure Copy File. Error : %v\n", err)
		}

		res, err := annex_ipfs.AddByFileNm(gitDir, fileNm)
		if err != nil {
			b.Logf("Fialure git annex add. Error : %v\n", err)
		}

		b.StartTimer()

		//測定対象
		if err = annex_ipfs.CopyToByKey("ipfs", res.Key, gitDir); err != nil {
			b.Logf("Fialure git annex copy to ipfs. Error : %v\n", err)
		}

		b.StopTimer()

		//後処理
		content, err := annex_ipfs.WhereisByKey(gitDir, res.Key)
		if err != nil {
			b.Logf("Fialure git annex whereis by key. Error : %v\n", err)
		}

		if err := op.PinRm(content.IpfsCid); err != nil {
			b.Logf("Fialure PinRm(). Error : %v\n", err)
		}
		if err := op.RepoGc(); err != nil {
			b.Logf("Fialure RepoGc(). Error : %v\n", err)
		}
		//db.AnnexUninit(gitDir)
		os.RemoveAll(gitDir)
		b.StartTimer()
	}
}

func BenchmarkAnnexCopyToIPFS_1k(b *testing.B) {
	benchAnnexCopyToIPFS(b, "1k.txt")
}
