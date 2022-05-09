package ipfs

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"unsafe"

	logv2 "unknwon.dev/clog/v2"
)

//mockファイルの生成
//mockgen -source ipfs_operation.go -destination mock/mock_ipfs_opetation.go

type IFIpfsOperation interface {
	FilesCopy(contentAddress, fullRepoFilePath string) error
	FilesStatus(folderPath string) (string, error)
	FilesRemove(folderPath string) error
	FilesIs(folderPath string) ([]string, error)
}

type IpfsOperation struct {
	Commander IFIpfsCommand
}

//　ipfs files cp....コマンド
// @param contentAddress コピーするコンテンツアドレス ex : QmT8LDwxQQqEBbChjBn4zEhiWtfRHNwwQYguNDjJZ9tME1
// @param fullFilePath コピー先ディレクトリ ex : /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/...../FileNm.txt
func (i *IpfsOperation) FilesCopy(contentAddress, fullRepoFilePath string) error {
	logv2.Info("[Copying IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	i.Commander.RemoveArgs()
	contentParam := "/ipfs/" + contentAddress
	i.Commander.AddArgs("files", "cp", contentParam, "-p", fullRepoFilePath)
	if _, err := i.Commander.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs files cp ...] Content Adress : %v, FullRepoFilePath : %v, err : %v", contentAddress, fullRepoFilePath, err)
	}
	logv2.Info("[Completion of Copy IPFS Filse] Content Adress : %v, FullRepoFilePath : %v", contentAddress, fullRepoFilePath)
	return nil
}

// ipfs files stat...コマンド
// @param folderPath ex /RepoOwnerNm/RepoNm/BranchNm/DatasetFoleder/input
func (i *IpfsOperation) FilesStatus(folderPath string) (string, error) {
	i.Commander.RemoveArgs()
	i.Commander.AddArgs("files", "stat", folderPath)
	msg, err := i.Commander.Run()
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
	i.Commander.RemoveArgs()
	i.Commander.AddArgs("files", "rm", "-r", folderPath)

	if _, err := i.Commander.Run(); err != nil {
		return fmt.Errorf("[Failure ipfs file rm ...] FolderPath : %v", folderPath)
	}
	logv2.Info("[Remove IPFS Folder] FolderPath: %v", folderPath)
	return nil
}

func (i *IpfsOperation) FilesIs(folderPath string) ([]string, error) {
	i.Commander.RemoveArgs()
	i.Commander.AddArgs("files", "ls", folderPath)
	msg, err := i.Commander.Run()
	if err != nil {
		return nil, fmt.Errorf("[Failure ipfs file is ...] <%v>, FolderPath : %v", err, folderPath)
	}
	strMsg := *(*string)(unsafe.Pointer(&msg))
	reg := "\r\n|\n"
	splitByline := regexp.MustCompile(reg).Split(strMsg, -1)
	return splitByline, nil
}

//直接、データをIPFSへのアップロードする。（echo [data] | ipfs add）
func DirectlyAdd(data string) (string, error) {

	echoCmd := exec.Command("echo", data)
	addCmd := exec.Command("ipfs", "add")

	pipe, err := echoCmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("Cannot getting StdoutPipe. Error Msg : [%v]", err)
	}
	defer pipe.Close()

	addCmd.Stdin = pipe

	echoCmd.Start()

	res, err := addCmd.Output()
	if err != nil {
		return "", fmt.Errorf("Failure Running Command <echo data | ipfs add>. Error Msg : [%v]", err)
	}
	arrMsg := strings.Split(string(res), " ")

	return arrMsg[2], nil
}
