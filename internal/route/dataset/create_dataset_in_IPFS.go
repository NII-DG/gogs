package dataset

import (
	"fmt"
	"strings"

	logv2 "unknwon.dev/clog/v2"

	"github.com/ivis-yoshida/gogs/internal/db"
	"github.com/ivis-yoshida/gogs/internal/ipfs"
)

type UploadDatasetInfo struct {
	InputAddress  string
	SrcAddress    string
	OutputAddress string
}

func GetDatasetAddress(repoBranchPath, datasetPath string, datasetData db.DatasetInfo) (UploadDatasetInfo, error) {

	//指定にデータセットフォルダがIPFS上に存在しないことを確認する。

	//IPFS上でデータセットのフォルダー構成を再現
	allContentList := datasetData.InputList
	allContentList = append(allContentList, datasetData.SrcList...)
	allContentList = append(allContentList, datasetData.OutputList...)
	if err := createDatasetStructure(allContentList); err != nil {
		//IPFS上のフォルダー構成を削除する
		if rmErr := ipfs.FilesRemove(datasetPath); rmErr != nil {
			return UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Creating Foleder on IPFS] <%v>,<%v>", err, rmErr)
		}
		return UploadDatasetInfo{}, fmt.Errorf("[Failure Create Foleder on IPFS] <%v>, Than Remove Creating Foleder", err)
	}

	// /input/, /src/, /output/ フォルダのフォルダアドレスを取得
	uploadDataset, err := getUploadDatasetInfo(datasetPath)
	if err != nil {
		return uploadDataset, err
	}

	//IPFS上のフォルダー構成を削除する
	if rmErr := ipfs.FilesRemove(datasetPath); rmErr != nil {
		return UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Created Foleder on IPFS]", rmErr)
	}
	return uploadDataset, nil
}

func createDatasetStructure(contentList []db.ContentInfo) error {

	for _, content := range contentList {
		if err := ipfs.FilesCopy(content.Address, content.File); err != nil {
			return err
		}
	}
	return nil
}

func getUploadDatasetInfo(datasetPath string) (UploadDatasetInfo, error) {
	inputPath := datasetPath + "/" + db.INPUT_FOLDER_NM
	srcPath := datasetPath + "/" + db.SRC_FOLDER_NM
	outputPath := datasetPath + "/" + db.OUTPUT_FOLDER_NM

	inputAddress, inputErr := ipfs.FilesStatus(inputPath)
	srcAddress, srcErr := ipfs.FilesStatus(srcPath)
	outputAddress, outputErr := ipfs.FilesStatus(outputPath)

	if inputErr != nil || srcErr != nil || outputErr != nil {
		return UploadDatasetInfo{}, fmt.Errorf("[Failure Get Upload Dataset Address From IPFS] <INPUT : %v>, <SRC : %v>, <OUTPUT : %v>", inputErr, srcErr, outputErr)
	}

	return UploadDatasetInfo{
		InputAddress:  inputAddress,
		SrcAddress:    srcAddress,
		OutputAddress: outputAddress,
	}, nil
}

func IsDatasetFolderOnIPFS(datasetPath string) (bool, error) {
	_, err := ipfs.FilesIs(datasetPath)
	if err != nil {
		logv2.Info("[err.Error()] %v", err.Error())
		if strings.Contains(err.Error(), "file does not exist") {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}
