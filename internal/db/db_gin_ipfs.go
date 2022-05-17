package db

import (
	"fmt"

	"github.com/G-Node/libgin/libgin/annex"
	"github.com/NII-DG/gogs/internal/annex_ipfs"
	"github.com/gogs/git-module"
	log "gopkg.in/clog.v1"
	logv2 "unknwon.dev/clog/v2"
)

type AnnexUploadInfo struct {
	FullContentHash string
	IpfsCid         string
	IsPrivate       bool
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
NOTE : methods : [sync and copy] locations are invert
ToDo : IPFSへアップロードしたコンテンツアドレスをupploadFileMapに追加する。
*/
func PublicAnnexUpload(upperpath, repoPath, remote string, annexAddRes []annex_ipfs.AnnexAddResponse) (map[string]AnnexUploadInfo, error) {
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
			return nil, fmt.Errorf("[git annex whereis Error] err : %v", err)
		} else {
			contentInfo, err := annex_ipfs.GetAnnexContentInfo(&msgWhereis)
			if err != nil {
				return nil, fmt.Errorf("[JSON Convert] err : %v ,fromPath : %v", err, repoPath)
			}
			contentLocation := upperpath + "/" + content.File
			contentMap[contentLocation] = AnnexUploadInfo{
				FullContentHash: content.Key,
				IpfsCid:         contentInfo.Key,
				IsPrivate:       false,
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

	for k, v := range contentMap {
		logv2.Trace("[contentMap] key : %v, FullContentHash : %v, IpfsCid: %v, IsPrivate : %v", k, v.FullContentHash, v.IpfsCid, v.IsPrivate)
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

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
*/
func annexSetupForIPFS(path string) {
	logv2.Info("Running annex add (with filesize filter) in '%s'", path)

	// Initialise annex in case it's a new repository

	if _, err := annex.Init(path); err != nil {
		logv2.Error("[Annex init failed] path : %v, err : %v", path, err)
		return
	} else {
		logv2.Info("[Annex init Success] path : %v", path)
	}

	// Upgrade to v8 in case the directory was here before and wasn't cleaned up properly
	if msg, err := annex.Upgrade(path); err != nil {
		logv2.Error("Annex upgrade failed: %v (%s)", err, msg)
		return
	}

	// Enable addunlocked for annex v8
	if msg, err := annex.SetAddUnlocked(path); err != nil {
		logv2.Error("Failed to set 'addunlocked' annex option: %v (%s)", err, msg)
	}

	// Set MD5 as default backend
	if msg, err := annex.MD5(path); err != nil {
		logv2.Error("Failed to set default backend to 'MD5': %v (%s)", err, msg)
	}

	// Set size filter in config
	//large fileのサイズを定義する（不要）
	// if msg, err := annex.SetAnnexSizeFilter(path, conf.Repository.Upload.AnnexFileMinSize*annex.MEGABYTE); err != nil {
	// 	logv2.Error("Failed to set size filter for annex: %v (%s)", err, msg)
	//}
	//conf.Repository.Upload.AnnexFileMinSize * annex.MEGABYTE

	//Setting initremote ipfs
	if err := setRemoteIPFS(path); err != nil {
		logv2.Warn("[Warn Initremoto IPFS] path : %v,  error : %v", path, err)

		if _, err := git.NewCommand("annex", "enableremote", "ipfs").RunInDir(path); err != nil {
			logv2.Error("[Failure enable remote(ipfs)] err : %v, path : %v", err, path)
		} else {
			logv2.Info("[Success enable remote(ipfs)] path : %v", path)
		}
		return
	} else {
		logv2.Info("[Initremoto IPFS] path : %v", path)
	}
}

/**
UPDATE : 2022/02/01
AUTHOR : dai.tsukioka
NOTE : Setting initremote ipfs
*/
func setRemoteIPFS(path string) error {
	cmd := git.NewCommand("annex", "initremote")
	cmd.AddArgs("ipfs", "type=external", "externaltype=ipfs", "encryption=none")
	_, err := cmd.RunInDir(path)
	return err
}
