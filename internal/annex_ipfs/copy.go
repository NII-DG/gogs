package annex_ipfs

import (
	"fmt"

	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
)

//git annex copy --from <reote> --key <key>　実行メソッド
//
//@param remote string 指定リモート（github, ipfs ....）
//
//@param key string Annexキー
//
//@param repoPath string　実行レポジトリパス
func CopyByKey(remote, key, repoPath string) error {
	log.Trace("Conducting <git annex copy --from %v --key %v> In %v", remote, key, repoPath)
	cmd := git.NewCommand("annex", "copy", "--from", remote, "--key", key)
	if _, err := cmd.RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git annex copy to ipfs] err : %v ,fromPath : %v", err, repoPath)
	}
	return nil
}

//git annex copy --from <reote> 実行メソッド
//
//@param remote string 指定リモート（github, ipfs ....）
//
//@param repoPath string　実行レポジトリパス
func Copy(remote, repoPath string) error {
	log.Trace("Conducting <git annex copy --from %v> In %v", remote, repoPath)
	cmd := git.NewCommand("annex", "copy", "--from", remote)
	if _, err := cmd.RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git annex copy to ipfs] err : %v ,fromPath : %v", err, repoPath)
	}
	return nil
}
