package dataset

import (
	"github.com/ivis-yoshida/gogs/internal/context"
	"github.com/ivis-yoshida/gogs/internal/form"
)

//データセット登録処理
func CreateDataset(c *context.Context, f form.DatasetFrom) {
	//実行ユーザ
	// userCode := c.User.Name
	// //レポジトリパス
	// repoBranchNm := c.Repo.RepoLink + "/" + c.Repo.BranchName
	// //登録データセット（フォルダー名）
	// datasetList := f.Datasets
	// //ブランチ
	// branch := c.Repo.BranchName

	//データセットフォーマットのチェック（datasetFolder : [input, src, output]フォルダーがあること、かつ、その配下にファイルがあること）

	//データセット内のコンテンツがBC上に存在するかをチェック

	//IPFS上でデータセット構築

	//データセットのBC登録

}
