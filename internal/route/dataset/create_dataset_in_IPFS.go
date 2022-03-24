package dataset

import (
	"fmt"
	"strings"

	logv2 "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/ipfs"
)

func GetDatasetAddress(datasetPath string, datasetData db.DatasetInfo) (bcapi.UploadDatasetInfo, error) {
	o := ipfs.IpfsOperation{}

	//指定にデータセットフォルダがIPFS上に存在しないことを確認する。
	//存在している場合、実行ユーザ以外の者がデータセット登録をしようとしているか > 前回の同ディレクトリの削除がうまくいかなかった場合
	is, err := isDatasetFolderOnIPFS(datasetPath)
	if err != nil {
		//内部エラー
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("<%v>", err)
	} else if is {
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("There Is A Possibility That Another User Upload Dataset To IPFS")
	}

	//IPFS上でデータセットのフォルダー構成を再現
	allContentList := datasetData.InputList
	allContentList = append(allContentList, datasetData.SrcList...)
	allContentList = append(allContentList, datasetData.OutputList...)
	if err := createDatasetStructure(allContentList); err != nil {
		//IPFS上のフォルダー構成を削除する
		if rmErr := o.FilesRemove(datasetPath); rmErr != nil {
			return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Creating Foleder on IPFS] <%v>,<%v>", err, rmErr)
		}
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Create Foleder on IPFS] <%v>, Than Remove Creating Foleder", err)
	}

	// /input/, /src/, /output/ フォルダのフォルダアドレスを取得
	uploadDataset, err := getUploadDatasetInfo(datasetPath)
	if err != nil {
		return uploadDataset, err
	}

	//IPFS上のフォルダー構成を削除する
	if rmErr := o.FilesRemove(datasetPath); rmErr != nil {
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Created Foleder on IPFS] %v", rmErr)
	}
	return uploadDataset, nil
}

func createDatasetStructure(contentList []db.ContentInfo) error {
	o := ipfs.IpfsOperation{}
	for _, content := range contentList {
		if err := o.FilesCopy(content.Address, content.File); err != nil {
			return err
		}
	}
	return nil
}

func getUploadDatasetInfo(datasetPath string) (bcapi.UploadDatasetInfo, error) {
	o := ipfs.IpfsOperation{}
	inputPath := datasetPath + "/" + db.INPUT_FOLDER_NM
	srcPath := datasetPath + "/" + db.SRC_FOLDER_NM
	outputPath := datasetPath + "/" + db.OUTPUT_FOLDER_NM

	inputAddress, inputErr := o.FilesStatus(inputPath)
	srcAddress, srcErr := o.FilesStatus(srcPath)
	outputAddress, outputErr := o.FilesStatus(outputPath)

	if inputErr != nil || srcErr != nil || outputErr != nil {
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Get Upload Dataset Address From IPFS] <INPUT : %v>, <SRC : %v>, <OUTPUT : %v>", inputErr, srcErr, outputErr)
	}

	return bcapi.UploadDatasetInfo{
		InputAddress:  inputAddress,
		SrcAddress:    srcAddress,
		OutputAddress: outputAddress,
	}, nil
}

func isDatasetFolderOnIPFS(datasetPath string) (bool, error) {
	o := ipfs.IpfsOperation{}
	_, err := o.FilesIs(datasetPath)
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
