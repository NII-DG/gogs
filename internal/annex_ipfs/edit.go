package annex_ipfs

import (
	"fmt"

	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
)

//git annex edit <file_path>　実行メソッド
//
//@param filePath string 編集可能ファイルのパス
//
//@param repoPath string　実行レポジトリパス
func editByFilePath(filePath, repoPath string) error {
	log.Trace("Conducting <git annex edit %v> In %v", filePath, repoPath)
	cmd := git.NewCommand("annex", "edit", filePath)
	if _, err := cmd.RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git annex edit %v] Error Msg : %v ,Repo Path : %v", filePath, err, repoPath)
	}
	return nil
}

//git annex edit .　実行メソッド
//
//@param filePath string 編集可能ファイルのパス
//
//@param repoPath string　実行レポジトリパス
func editAll(filePath, repoPath string) error {
	log.Trace("Conducting <git annex edit .> In %v", repoPath)
	cmd := git.NewCommand("annex", "edit", ".")
	if _, err := cmd.RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git annex edit .] Error Msg : %v ,Repo Path : %v", err, repoPath)
	}
	return nil
}
