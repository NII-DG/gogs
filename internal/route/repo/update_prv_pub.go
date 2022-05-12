package repo

import (
	"fmt"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	log "unknwon.dev/clog/v2"
)

//
func UpdateDataPrvToPub(c *context.Context) {
	repo := c.Repo.Repository
	ownerRepoNm := fmt.Sprintf("/%v/%v", c.Repo.Owner.Name, repo.Name) // /OwnerNm/RepoNm
	userCode := c.User.Name
	//BCからコンテンツ情報を取得
	response, err := bcapi.GetContentByFolder(userCode, ownerRepoNm)
	if err != nil {
		log.Error("Failure Getting ContentInfo From BC In %v. Error Massage : %v", ownerRepoNm, err)
		c.RenderWithErr(c.Tr("ブロックチェーンからコンテンツ情報取得に失敗しました"), SETTINGS_OPTIONS, nil)
	}

	//非公開データ情報のみ抽出
	var bcContentInfoList []db.ContentInfo
	for _, v := range response.ContentsInFolder {
		if len(v.FullContentHash) > 0 {
			bcContentInfoList = append(bcContentInfoList, db.ContentInfo{File: v.ContentLocation,
				FullContentHash: v.FullContentHash,
				Address:         v.IpfsCid})
		}
	}

	//非公開データを公開データして、IPFSへのアップロードする。
	contentMap, err := repo.UpdateFilePrvToPub(db.UploadRepoOption{
		Branch:            c.Repo.BranchName,
		UpperRopoPath:     ownerRepoNm,
		BcContentInfoList: bcContentInfoList,
	})
	if err != nil {
		log.Error("Failure Getting ContentInfo From BC In %v(%v). Error Massage : %v", ownerRepoNm, repo.LocalCopyPath(), err)
		c.RenderWithErr(c.Tr("レポジトリファイルの公開化に失敗しました"), SETTINGS_OPTIONS, nil)
	}
	//BCにコンテンツ情報を登録する。
	if err := bcapi.CreateContentHistory(c.User.Name, contentMap); err != nil {
		log.Error("Failure Applying for Registering ContentHistory To BC In %v. Error Massage : %v", ownerRepoNm, err)
		c.RenderWithErr(c.Tr("ブロックチェーンへの登録申請に失敗しました"), SETTINGS_OPTIONS, nil)
	}

}
