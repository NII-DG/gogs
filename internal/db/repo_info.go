package db

import (
	"fmt"
	"regexp"
	"strings"
	"unsafe"

	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
)

// レポジトリに付随するブランチ名リストを取得する。
//@param repoPath string システム上のレポジトリパス
func GetBranchList(repoPath string) ([]string, error) {
	branchList := []string{}
	msg, err := git.NewCommand("branch").RunInDir(repoPath)
	if err != nil {
		return nil, fmt.Errorf("[git branch] Failure Getting Branch List. Error Mag[%v]", err)
	} else {
		strMsg := *(*string)(unsafe.Pointer(&msg)) //[]byte to string
		reg := "\r\n|\n"
		list := regexp.MustCompile(reg).Split(strMsg, -1) //改行分割
		for _, unit := range list {
			log.Trace("branchNm : %v", unit)
			if !strings.Contains(unit, "synced/") {
				unit = strings.Replace(unit, " ", "", -1)
				unit = strings.Replace(unit, "*", "", -1)
				if len(unit) > 0 {
					branchList = append(branchList, unit)
				}
			}
		}
	}
	return branchList, nil
}
