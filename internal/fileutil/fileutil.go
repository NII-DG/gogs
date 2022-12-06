package fileutil

import (
	"io/fs"
	"io/ioutil"
)

type IFFileUtil interface {
	GetFileBypath(path string) ([]byte, error)
	ReadDirBypath(path string) ([]fs.FileInfo, error)
}

type FileUtil struct {
}

// 指定パスのファイルをバイナリで取得する。
func (f *FileUtil) GetFileBypath(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// 指定パスのディレクトリ内の情報を取得する。
func (f *FileUtil) ReadDirBypath(path string) ([]fs.FileInfo, error) {
	return ioutil.ReadDir(path)
}
