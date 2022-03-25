package dataset

import (
	"fmt"
	"strings"

	logv2 "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/ipfs"
)

//mockファイルの生成
//mockgen -source create_dataset_in_IPFS.go -destination mock/mock_create_dataset_in_IPFS.go

type IFDatasetCreater interface {
	GetDatasetAddress(datasetPath string, datasetData db.DatasetInfo) (bcapi.UploadDatasetInfo, error)
	createDatasetStructure(contentList []db.ContentInfo) error
	getUploadDatasetInfo(datasetPath string) (bcapi.UploadDatasetInfo, error)
	isDatasetFolderOnIPFS(datasetPath string) (bool, error)
}

type DatasetCreater struct {
	Operater ipfs.IFIpfsOperation
}

func (d *DatasetCreater) GetDatasetAddress(datasetPath string, datasetData db.DatasetInfo) (bcapi.UploadDatasetInfo, error) {

	//指定にデータセットフォルダがIPFS上に存在しないことを確認する。
	//存在している場合、実行ユーザ以外の者がデータセット登録をしようとしているか > 前回の同ディレクトリの削除がうまくいかなかった場合
	is, err := d.isDatasetFolderOnIPFS(datasetPath)
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
	if err := d.createDatasetStructure(allContentList); err != nil {
		//IPFS上のフォルダー構成を削除する
		if rmErr := d.Operater.FilesRemove(datasetPath); rmErr != nil {
			return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Creating Foleder on IPFS] <%v>,<%v>", err, rmErr)
		}
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Create Foleder on IPFS] <%v>, Than Remove Creating Foleder", err)
	}

	// /input/, /src/, /output/ フォルダのフォルダアドレスを取得
	uploadDataset, err := d.getUploadDatasetInfo(datasetPath)
	if err != nil {
		return uploadDataset, err
	}

	//IPFS上のフォルダー構成を削除する
	if rmErr := d.Operater.FilesRemove(datasetPath); rmErr != nil {
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Remove Created Foleder on IPFS] %v", rmErr)
	}
	return uploadDataset, nil
}

func (d *DatasetCreater) createDatasetStructure(contentList []db.ContentInfo) error {
	for _, content := range contentList {
		if err := d.Operater.FilesCopy(content.Address, content.File); err != nil {
			return err
		}
	}
	return nil
}

func (d *DatasetCreater) getUploadDatasetInfo(datasetPath string) (bcapi.UploadDatasetInfo, error) {
	inputPath := datasetPath + "/" + db.INPUT_FOLDER_NM
	inputAddress, inputErr := d.Operater.FilesStatus(inputPath)

	srcPath := datasetPath + "/" + db.SRC_FOLDER_NM
	srcAddress, srcErr := d.Operater.FilesStatus(srcPath)

	outputPath := datasetPath + "/" + db.OUTPUT_FOLDER_NM
	outputAddress, outputErr := d.Operater.FilesStatus(outputPath)

	if inputErr != nil || srcErr != nil || outputErr != nil {
		return bcapi.UploadDatasetInfo{}, fmt.Errorf("[Failure Get Upload Dataset Address From IPFS] <INPUT : %v>, <SRC : %v>, <OUTPUT : %v>", inputErr, srcErr, outputErr)
	}

	return bcapi.UploadDatasetInfo{
		InputAddress:  inputAddress,
		SrcAddress:    srcAddress,
		OutputAddress: outputAddress,
	}, nil
}

func (d *DatasetCreater) isDatasetFolderOnIPFS(datasetPath string) (bool, error) {
	_, err := d.Operater.FilesIs(datasetPath)
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
