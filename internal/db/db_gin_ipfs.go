package db

import (
	"fmt"

	"github.com/NII-DG/gogs/internal/annex_ipfs"
	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"
	logv2 "unknwon.dev/clog/v2"
)

//ToDo :upploadFileMap(map)にKeyを追加する。
func annexAddAndGetInfo(repoPath string, all bool, files ...string) ([]annex_ipfs.AnnexAddResponse, error) {
	cmd := git.NewCommand("annex", "add", "--json")
	if all {
		cmd.AddArgs(".")
	}
	msg, err := cmd.AddArgs(files...).RunInDir(repoPath)
	if err == nil {
		reslist, err := annex_ipfs.GetAnnexAddInfo(&msg)
		if err != nil {
			return nil, fmt.Errorf("[Annex Add Json Error]: %v", err)
		}
		return reslist, nil
	}
	return nil, err
}

type AnnexUploadInfo struct {
	FullContentHash string
	IpfsCid         string
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
NOTE : methods : [sync and copy] locations are invert
ToDo : IPFSへアップロードしたコンテンツアドレスをupploadFileMapに追加する。
*/
func PublicannexUpload(upperpath, repoPath, remote string, annexAddRes []annex_ipfs.AnnexAddResponse) (map[string]AnnexUploadInfo, error) {
	contentMap := map[string]AnnexUploadInfo{}
	//ipfsへ実データをコピーする。
	logv2.Info("[Uploading annexed data to %v] path : %v", remote, repoPath)
	for _, content := range annexAddRes {
		cmd := git.NewCommand("annex", "copy", "--to", remote, "--key", content.Key)
		if _, err := cmd.RunInDir(repoPath); err != nil {
			return nil, fmt.Errorf("[Failure git annex copy to %v] err : %v ,fromPath : %v", remote, err, repoPath)
		}
	}

	//コンテンツアドレスの取得
	for _, content := range annexAddRes {
		if msgWhereis, err := git.NewCommand("annex", "whereis", "--json", "--key", content.Key).RunInDir(repoPath); err != nil {
			logv2.Error("[git annex whereis Error] err : %v", err)
		} else {
			contentInfo, err := annex_ipfs.GetAnnexContentInfo(&msgWhereis)
			if err != nil {
				return nil, fmt.Errorf("[JSON Convert] err : %v ,fromPath : %v", err, repoPath)
			}
			contentLocation := upperpath + "/" + content.File
			contentMap[contentLocation] = AnnexUploadInfo{
				FullContentHash: "",
				IpfsCid:         contentInfo.Hash,
			}

		}
	}
	//IPFSへアップロードしたコンテンツロケーションを表示
	index := 1
	for k := range contentMap {
		logv2.Info("[Upload to IPFS] No.%v file : %v", index, k)
		upload_No := &index
		*upload_No++
	}

	//リモートと同期（メタデータを更新）
	log.Info("Synchronising annex info : %v", repoPath)
	if msg, err := git.NewCommand("annex", "sync").RunInDir(repoPath); err != nil {
		return nil, fmt.Errorf("[Failure git-annex sync] err : %v, msg : %s", err, msg)
	} else {
		logv2.Info("[Success git-annex sync] path : %v", repoPath)
	}

	return contentMap, nil
}

func PrivateAnnexUpload(upperpath, repoPath, remote string, annexAddRes []annex_ipfs.AnnexAddResponse) error {
	upContentName := []string{}
	//ipfsへ実データをコピーする。
	logv2.Info("[Uploading annexed data to %v] path : %v", remote, repoPath)
	for _, content := range annexAddRes {
		cmd := git.NewCommand("annex", "copy", "--to", remote, "--key", content.Key)
		if _, err := cmd.RunInDir(repoPath); err != nil {
			return fmt.Errorf("[Failure git annex copy to %v] err : %v ,fromPath : %v", remote, err, repoPath)
		}
	}

	//コンテンツアドレスの取得
	for _, content := range annexAddRes {
		if _, err := git.NewCommand("annex", "whereis", "--json", "--key", content.Key).RunInDir(repoPath); err != nil {
			logv2.Error("[git annex whereis Error] err : %v", err)
		} else {
			contentLocation := upperpath + "/" + content.File
			upContentName = append(upContentName, contentLocation)

		}
	}
	//IPFSへアップロードしたコンテンツロケーションを表示
	index := 1
	for _, v := range upContentName {
		logv2.Info("[Upload to IPFS] No.%v file : %v", index, v)
		upload_No := &index
		*upload_No++
	}

	//リモートと同期（メタデータを更新）
	log.Info("Synchronising annex info : %v", repoPath)
	if msg, err := git.NewCommand("annex", "sync").RunInDir(repoPath); err != nil {
		return fmt.Errorf("[Failure git-annex sync] err : %v, msg : %s", err, msg)
	} else {
		logv2.Info("[Success git-annex sync] path : %v", repoPath)
	}

	return nil
}
