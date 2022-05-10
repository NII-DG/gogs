package repo

import (
	"fmt"
	"strings"

	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/form"
	"github.com/NII-DG/gogs/internal/gitutil"
	"github.com/NII-DG/gogs/internal/ipfs"
	"github.com/NII-DG/gogs/internal/route/dataset"
	logv2 "unknwon.dev/clog/v2"
)

func CreateDataset(c *context.Context, f form.DatasetFrom) {

	c.Data["PageIsViewFiles"] = true

	if c.Repo.Repository.IsBare {
		c.Success(BARE)
		return
	}

	title := c.Repo.Repository.Owner.Name + "/" + c.Repo.Repository.Name
	if len(c.Repo.Repository.Description) > 0 {
		title += ": " + c.Repo.Repository.Description
	}
	c.Data["Title"] = title
	if c.Repo.BranchName != c.Repo.Repository.DefaultBranch {
		c.Data["Title"] = title + " @ " + c.Repo.BranchName
	}
	c.Data["RequireHighlightJS"] = true

	//コンテンツロケーションの定義
	var contentLocation string

	branchLink := c.Repo.RepoLink + "/src/" + c.Repo.BranchName
	treeLink := branchLink
	rawLink := c.Repo.RepoLink + "/raw/" + c.Repo.BranchName
	datasetLink := c.Repo.RepoLink + "/dataset/" + c.Repo.BranchName

	isRootDir := false
	if len(c.Repo.TreePath) > 0 {
		treeLink += "/" + c.Repo.TreePath
		log.Info("[Debug_1 Add treeLink] new path : %v, add path : %v", treeLink, c.Repo.TreePath)
		temploc := &contentLocation
		*temploc = c.Repo.RepoLink + "/" + c.Repo.BranchName + "/" + c.Repo.TreePath

	} else {
		isRootDir = true

		// Only show Git stats panel when view root directory
		var err error
		c.Repo.CommitsCount, err = c.Repo.Commit.CommitsCount()
		if err != nil {
			c.Error(err, "count commits")
			return
		}
		c.Data["CommitsCount"] = c.Repo.CommitsCount
	}
	c.Data["PageIsRepoHome"] = isRootDir

	// Get current entry user currently looking at.
	//選択フォルダーの下にフォルダー、ファイルの確認
	entry, err := c.Repo.Commit.TreeEntry(c.Repo.TreePath)

	if err != nil {
		c.NotFoundOrError(gitutil.NewError(err), "get tree entry")
		return
	}

	if entry.IsTree() {
		renderDirectoryFromBcapi(c, treeLink)
	} else {
		renderFileFromIPFS(c, entry, treeLink, rawLink, contentLocation)
	}
	if c.Written() {
		return
	}

	setEditorconfigIfExists(c)
	if c.Written() {
		return
	}

	var treeNames []string
	paths := make([]string, 0, 5)
	if len(c.Repo.TreePath) > 0 {
		treeNames = strings.Split(c.Repo.TreePath, "/")
		for i := range treeNames {
			paths = append(paths, strings.Join(treeNames[:i+1], "/"))
		}

		c.Data["HasParentPath"] = true
		if len(paths)-2 >= 0 {
			c.Data["ParentPath"] = "/" + paths[len(paths)-2]
		}
	}

	c.Data["Paths"] = paths
	c.Data["TreeLink"] = treeLink
	c.Data["TreeNames"] = treeNames
	c.Data["BranchLink"] = branchLink
	c.Data["DatasetLink"] = datasetLink

	//実行ユーザ
	userCode := c.User.Name
	//レポジトリパス
	repoBranchPath := c.Repo.RepoLink + "/" + c.Repo.BranchName
	//登録データセット（フォルダー名）
	datasetList := f.DatasetList
	//ブランチ
	branchNm := c.Repo.BranchName

	//レポジトリのクローン
	if err := c.Repo.Repository.CloneRepo(branchNm); err != nil {
		c.Error(err, "[Error] CheckDatadetAndGetContentAddress()")
		return
	}

	//データセットフォーマットのチェック（datasetFolder : [input, src, output]フォルダーがあること、かつ、その配下にファイルがあること）
	if err := c.Repo.Repository.CheckDatasetFormat(datasetList); err != nil {
		msg := fmt.Sprint(err)
		c.RenderWithErr(msg, HOME, &f)
		return
	}

	//各データセットパスとその内のフォルダ内のコンテンツ情報を持つMapを取得する。
	datasetNmToFileMap, err := c.Repo.Repository.GetContentAddress(datasetList, repoBranchPath)
	if err != nil {
		c.Error(err, "[Error] CheckDatadetAndGetContentAddress()")
		return
	}

	//ローカルレポジトリの削除
	db.AnnexUninit(c.Repo.Repository.LocalCopyPath()) // Uninitialise annex to prepare for deletion
	if err := db.RemoveLocalRepository(c.Repo.Repository.LocalCopyPath()); err != nil {
		c.Error(err, "[Error] Cannot Remove Local Repository")
		return
	}

	//データセット内のコンテンツがBC上に存在するかをチェック
	for datasetPath, datasetData := range datasetNmToFileMap {
		if bcContentList, err := bcapi.GetContentByFolder(userCode, datasetPath); err != nil {
			c.Error(err, "Error In Exchanging BCAPI ")
			return
		} else if !isContainDatasetFileInBC(datasetData, bcContentList) {
			logv2.Error("[A Part Of Dataset File Is Not Registered In BC] Dataset Name : %v", datasetPath)
			msg := fmt.Sprintf("アップロードされたファイルがブロックチェーンに未登録または登録処理中の可能性があります。再度、データセット登録を申請してください。")
			c.RenderWithErr(msg, HOME, &f)
			return
		}
	}

	//IPFS上でデータセット構築
	uploadDatasetMap := map[string]bcapi.UploadDatasetInfo{}
	for datasetPath, datasetData := range datasetNmToFileMap {
		datasetCreater := dataset.DatasetCreater{
			Operater: &ipfs.IpfsOperation{
				Commander: ipfs.NewCommand(),
			},
		}
		if uploadDataset, err := datasetCreater.GetDatasetAddress(datasetPath, datasetData); err != nil {
			logv2.Error("[Get each Address IN Dataset] %v", err)
			c.Error(err, "データセット内の各フォルダアドレスが取得できませんでした")
			return
		} else {
			uploadDatasetMap[datasetPath] = uploadDataset
		}
	}

	//データセットのBC登録
	notCreatedDataset, err := bcapi.CreateDatasetToken(userCode, uploadDatasetMap)
	if err != nil {
		logv2.Error("[Failure Create Dataset Token] %v", err)
		c.Error(err, "データセットのトークン化に失敗しました")
		return
	}
	if len(notCreatedDataset.DatasetList) > 0 {
		//登録できなかったデータセットの表示
		notCreatesDatasetList := ""
		for _, dataset := range notCreatedDataset.DatasetList {
			logv2.Warn("[Already Exists Dataset Token] %v", dataset.DatasetLocation)
			temStr := &notCreatesDatasetList
			addDataset := "[" + dataset.DatasetLocation + "] "
			*temStr = *temStr + addDataset

		}
		msg := fmt.Sprintf("%vは既に登録されています。", notCreatesDatasetList)
		c.Flash.ErrorMsg = msg
	}

	//登録依頼したデータセットの表示
	isDisplaySubmitDataset := false
	createDatasetListStr := ""
	for k := range uploadDatasetMap {
		isSumbitBc := true
		for _, nonDataset := range notCreatedDataset.DatasetList {
			if k == nonDataset.DatasetLocation {
				isTmp := &isSumbitBc
				*isTmp = false
			}
		}
		if isSumbitBc {
			isTmpDisplay := &isDisplaySubmitDataset
			*isTmpDisplay = true
			tmpStr := &createDatasetListStr
			*tmpStr = *tmpStr + "[" + k + "]  "
		}
	}
	if isDisplaySubmitDataset {
		c.Flash.InfoMsg = fmt.Sprintf("%vをブロックチェーンへ登録申請しました。", createDatasetListStr)

	}
	c.Data["Flash"] = c.Flash
	c.Success(HOME)
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
		if contentData.File == bcContent.ContentLocation && contentData.Address == bcContent.IpfsCid {
			return true
		}
	}
	return false
}
