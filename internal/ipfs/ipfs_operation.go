package ipfs

import (
	"fmt"
	"regexp"
	"unsafe"

	logv2 "unknwon.dev/clog/v2"
)

type IpfsOperation struct {
	command IFIpfsCommand
}

//　ipfs files cp....コマンド
// @param contentAddress コピーするコンテンツアドレス ex : QmT8LDwxQQqEBbChjBn4zEhiWtfRHNwwQYguNDjJZ9tME1
// @param fullFilePath コピー先ディレクトリ ex : /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/...../FileNm.txt
func (i *IpfsOperation) FilesCopy(contentAddress, fullRepoFilePath string) error {
	logv2.Info("[Copying IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	contentParam := "/ipfs/" + contentAddress
	i.command = NewCommand("files", "cp", contentParam, "-p", fullRepoFilePath)
	if _, err := i.command.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs files cp ...] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	}
	logv2.Info("[Completion of Copy IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	return nil
}

// ipfs files stat...コマンド
// @param folderPath ex /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/input
func (i *IpfsOperation) FilesStatus(folderPath string) (string, error) {
	i.command = NewCommand("files", "stat", folderPath)
	msg, err := i.command.Run()
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
	i.command = NewCommand("files", "rm", "-r", folderPath)

	if _, err := i.command.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs file rm ...] FolderPath : %v", folderPath)
	}
	logv2.Info("[Remove IPFS Folder] FolderPath: %v", folderPath)
	return nil
}

func (i *IpfsOperation) FilesIs(folderPath string) ([]string, error) {
	i.command = NewCommand("files", "ls", folderPath)
	msg, err := i.command.Run()
	if err != nil {
		return nil, fmt.Errorf("[Failure ipfs file is ...] <%v>, FolderPath : %v", err, folderPath)
	}
	strMsg := *(*string)(unsafe.Pointer(&msg))
	reg := "\r\n|\n"
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1)
	return splitByline, nil
}
