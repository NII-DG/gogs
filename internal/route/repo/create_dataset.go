package repo

import (
	"fmt"
	"path/filepath"
	"strings"

	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/form"
	"github.com/NII-DG/gogs/internal/gitutil"
	"github.com/NII-DG/gogs/internal/ipfs"
	"github.com/NII-DG/gogs/internal/jsonfunc"
	"github.com/NII-DG/gogs/internal/route/dataset"
	"github.com/NII-DG/gogs/internal/util"
	logv2 "unknwon.dev/clog/v2"
)

func CreateDataset(c *context.Context, f form.DatasetFrom) {

	c.Data["PageIsViewFiles"] = true
	repository := c.Repo.Repository

	if repository.IsBare {
		c.Success(BARE)
		return
	}

	title := repository.Owner.Name + "/" + repository.Name
	if len(c.Repo.Repository.Description) > 0 {
		title += ": " + repository.Description
	}
	c.Data["Title"] = title
	if c.Repo.BranchName != repository.DefaultBranch {
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

	//データセットの構築
	UploadDataset(c, f)

	c.Data["Flash"] = c.Flash
	c.Success(HOME)
}

func UploadDataset(c *context.Context, f form.DatasetFrom) {
	repository := c.Repo.Repository
	//実行ユーザ
	userCode := c.User.Name
	//レポジトリパス
	repoBranchPath := c.Repo.RepoLink + "/" + c.Repo.BranchName
	//登録データセット（フォルダー名）
	datasetList := f.DatasetList
	//ブランチ
	branchNm := c.Repo.BranchName

	datasetCreater := dataset.DatasetCreater{
		Operater: &ipfs.IpfsOperation{
			Commander: ipfs.NewCommand(),
		},
	}

	//レポジトリのクローン
	repository.CheckIn()
	defer repository.CheckOut()
	if err := repository.CloneRepo(branchNm); err != nil {
		c.Error(err, "[Error] CheckDatadetAndGetContentAddress()")
		return
	}

	//データセットフォーマットのチェック（datasetFolder : [input, src, output]フォルダーがあること、かつ、その配下にファイルがあること）
	if err := repository.CheckDatasetFormat(datasetList); err != nil {
		msg := fmt.Sprint(err)
		c.RenderWithErr(msg, HOME, &f)
		return
	}

	//各データセットパスとその内のフォルダ内のコンテンツ情報を持つMapを取得する。
	datasetNmToContentInfo, err := repository.GetContentInfoByDatasetNm(datasetList, repoBranchPath)
	if err != nil {
		c.Error(err, "[Error] CheckDatadetAndGetContentAddress()")
		return
	}

	//ローカルレポジトリの削除
	db.AnnexUninit(repository.LocalCopyPath()) // Uninitialise annex to prepare for deletion
	if err := db.RemoveLocalRepository(repository.LocalCopyPath()); err != nil {
		c.Error(err, "[Error] Cannot Remove Local Repository")
		return
	}

	//データセット内のコンテンツがBC上に存在するかをチェック
	//チェックをパスしたコンテンツのコンテンツロケーションとハッシュ値の組リストを取得
	checkOKContentList := map[string][]dataset.CheckedContent{}
	for datasetNm, dataList := range datasetNmToContentInfo {
		datasetPath := filepath.Join(repoBranchPath, datasetNm)
		bcContentList, err := bcapi.GetContentByFolder(userCode, datasetPath)
		if err != nil {
			c.Error(err, "Error In Exchanging BCAPI ")
			return
		}
		//BC取得情報をコンテンツロケーションベースのMapを作成
		LocToInfo := exchangeLocationMapToContentInfo(bcContentList)

		//検証対象のBC登録情報を取得
		for _, annexInfo := range dataList {
			fullPath := filepath.Join(repoBranchPath, annexInfo.FileNm)
			bcInfo, is := LocToInfo[fullPath]
			if !is {
				//BCに登録されてない場合エラー
				c.Flash.ErrorMsg = fmt.Sprintf("ブロックチェーンにコンテンツ[%v]の情報がありません", fullPath)
				return
			}
			//コンテンツが公開データか非公開データにを判定
			if bcInfo.IsPrivate {
				//非公開データ
				bHash, err := datasetCreater.Operater.Cat(annexInfo.IpfsCid)
				if err != nil {
					c.Flash.ErrorMsg = fmt.Sprintf("IPFSからコンテンツ[%v]のハッシュ値の取得に失敗しました。", fullPath)
					return
				}
				hash := util.BytesToString(bHash)
				if hash == bcInfo.FullContentHash {

					checkOKContentList[datasetPath] = append(checkOKContentList[datasetPath], dataset.CheckedContent{ContentLocation: bcInfo.ContentLocation, Hash: bcInfo.FullContentHash})
				} else {
					//コンテンツのハッシュ値が一致しない場合はエラー
					logv2.Error("Not Match Private Content Hash. git annex [%v] VS BC [%v]", hash, bcInfo.FullContentHash)
					c.Flash.ErrorMsg = fmt.Sprintf("コンテンツ[%v]の情報がBC登録情報を一致しません", fullPath)
					return
				}
			} else {
				//公開データ
				if bcInfo.FullContentHash == annexInfo.Key {
					checkOKContentList[datasetPath] = append(checkOKContentList[datasetPath], dataset.CheckedContent{ContentLocation: bcInfo.ContentLocation, Hash: bcInfo.FullContentHash})
				} else {
					//コンテンツのハッシュ値が一致しない場合はエラー
					logv2.Error("Not Match Public Content Hash. git annex [%v] VS BC [%v]", annexInfo.Key, bcInfo.FullContentHash)
					c.Flash.ErrorMsg = fmt.Sprintf("コンテンツ[%v]の情報がBC登録情報を一致しません", fullPath)
					return
				}
			}
		}

	}

	//IPFS上でデータセット構築
	uploadDatasetMap := map[string]bcapi.UploadDatasetInfo{}
	for datasetPath, dataList := range checkOKContentList {

		if uploadDataset, err := datasetCreater.GetDatasetAddress(datasetPath, dataList); err != nil {
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

}

type BcContentInfo struct {
	ContentLocation string
	FullContentHash string
	IpfsCid         string
	IsPrivate       bool
}

func exchangeLocationMapToContentInfo(raw jsonfunc.ResContentsInFolder) map[string]BcContentInfo {
	bcContentInfo := map[string]BcContentInfo{}
	for _, v := range raw.ContentsInFolder {
		bcContentInfo[v.ContentLocation] = BcContentInfo{
			ContentLocation: v.ContentLocation,
			FullContentHash: v.FullContentHash,
			IpfsCid:         v.IpfsCid,
			IsPrivate:       v.IsPrivate,
		}
	}
	return bcContentInfo
}
