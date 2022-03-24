package ipfs

import (
	"fmt"
	"regexp"
	"unsafe"

	logv2 "unknwon.dev/clog/v2"
)

type IpfsOperation struct {
	Command IFIpfsCommand
}

//　ipfs files cp....コマンド
// @param contentAddress コピーするコンテンツアドレス ex : QmT8LDwxQQqEBbChjBn4zEhiWtfRHNwwQYguNDjJZ9tME1
// @param fullFilePath コピー先ディレクトリ ex : /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/...../FileNm.txt
func (i *IpfsOperation) FilesCopy(contentAddress, fullRepoFilePath string) error {
	logv2.Info("[Copying IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	contentParam := "/ipfs/" + contentAddress
	i.Command.AddArgs("files", "cp", contentParam, "-p", fullRepoFilePath)
	logv2.Info("[i.Command] %v", i.Command)
	if _, err := i.Command.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs files cp ...] Content Adress : %v, FullRepoFilePath : %v, err : %v", contentAddress, fullRepoFilePath, err)
	}
	logv2.Info("[Completion of Copy IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	return nil
}

// ipfs files stat...コマンド
// @param folderPath ex /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/input
func (i *IpfsOperation) FilesStatus(folderPath string) (string, error) {
	i.Command.AddArgs("files", "stat", folderPath)
	msg, err := i.Command.Run()
	if err != nil {
		return "", fmt.Errorf("[Failure ipfs files stat ...] FolderPath : %v", folderPath)
	}
	//msgからフォルダーアドレスを取得
	strMsg := *(*string)(unsafe.Pointer(&msg))
	reg := "\r\n|\n"
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1)
	return splitByline[0], nil
}

// ipfs file rm...コマンド
// @param folderNm ex /RepoOwnerNm/RepoNm/BranchNm/DatasetFolederNm
func (i *IpfsOperation) FilesRemove(folderPath string) error {
	logv2.Info("[Removing IPFS Folder] FolderPath: %v", folderPath)
	i.Command.AddArgs("files", "rm", "-r", folderPath)

	if _, err := i.Command.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs file rm ...] FolderPath : %v", folderPath)
	}
	logv2.Info("[Remove IPFS Folder] FolderPath: %v", folderPath)
	return nil
}

func (i *IpfsOperation) FilesIs(folderPath string) ([]string, error) {
	i.Command.AddArgs("files", "ls", folderPath)
	msg, err := i.Command.Run()
	if err != nil {
		return nil, fmt.Errorf("[Failure ipfs file is ...] <%v>, FolderPath : %v", err, folderPath)
	}
	strMsg := *(*string)(unsafe.Pointer(&msg))
	reg := "\r\n|\n"
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1)
	return splitByline, nil
}
