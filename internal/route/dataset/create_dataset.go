package dataset

import (
	"fmt"

	"github.com/ivis-yoshida/gogs/internal/bcapi"
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/db"
	"github.com/ivis-yoshida/gogs/internal/form"
)

//データセット登録処理
func CreateDataset(c *context.Context, f form.DatasetFrom) {
	//実行ユーザ
	userCode := c.User.Name
	//レポジトリパス
	repoBranchNm := c.Repo.RepoLink + "/" + c.Repo.BranchName
	//登録データセット（フォルダー名）
	datasetList := f.Datasets
	//ブランチ
	branch := c.Repo.BranchName

	//データセットフォーマットのチェック（datasetFolder : [input, src, output]フォルダーがあること、かつ、その配下にファイルがあること）
	//各データセットパスとその内のフォルダ内のコンテンツ情報を持つMapを取得する。
	datasetNmToFileMap, err := c.Repo.Repository.CheckDatadetAndGetContentAddress(datasetList, branch, repoBranchNm)
	if err != nil {
		c.Error(err, "[Error] CheckDatadetAndGetContentAddress()")
		return
	}
	//データセット内のコンテンツがBC上に存在するかをチェック
	for datasetPath, datasetData := range datasetNmToFileMap {
		if bcContentList, err := bcapi.GetContentByFolder(userCode, datasetPath); err != nil {
			c.Error(err, "Error In Exchanging BCAPI ")
			return
		} else if !isContainDatasetFileInBC(datasetData, bcContentList) {
			var err error = fmt.Errorf("[A Part Of Dataset File Is Not Registered In BC] Dataset Name : %v", datasetPath)
			c.Error(err, "不正なファイルが含まれています")
			return
		}
	}
	//IPFS上でデータセット構築

	//データセットのBC登録

}

func isContainDatasetFileInBC(datasetData db.DatasetInfo, bcContentList bcapi.ResContentsInFolder) bool {
	for _, inputData := range datasetData.InputList {
		if !isContainFileInBc(inputData, bcContentList) {
			return false
		}
	}
	for _, srcData := range datasetData.SrcList {
		if !isContainFileInBc(srcData, bcContentList) {
			return false
		}
	}
	for _, outData := range datasetData.OutputList {
		if !isContainFileInBc(outData, bcContentList) {
			return false
		}
	}
	return true
}

func isContainFileInBc(contentData db.ContentInfo, bcContentList bcapi.ResContentsInFolder) bool {
	for _, bcContent := range bcContentList.ContentsInFolder {
		if contentData.File == bcContent.ContentLocation && contentData.Address == bcContent.ContentAddress {
			return true
		}
	}
	return false
}
