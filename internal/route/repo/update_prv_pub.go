package repo

import (
	"fmt"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/form"
	log "unknwon.dev/clog/v2"
)

//
func UpdateDataPrvToPub(c *context.Context, f form.RepoSetting) {
	repo := c.Repo.Repository
	ownerRepoNm := fmt.Sprintf("/%v/%v", c.Repo.Owner.Name, repo.Name) // /OwnerNm/RepoNm
	//BCからコンテンツ情報を取得
	_, err := bcapi.GetContentByFolder(c.User.Name, ownerRepoNm)
	if err != nil {
		log.Error("Failure Getting ContentInfo From BC In %v. Error Massage : %v", ownerRepoNm, err)
		c.RenderWithErr(c.Tr("ブロックチェーンへの登録中にエラーが発生し、失敗しました"), SETTINGS_OPTIONS, &f)
	}

	//BCにコンテンツ情報を登録する。
	// if err := bcapi.CreateContentHistory(c.User.Name, contentMap); err != nil {
	// 	log.Error("Failure Update Private Data To Public Data. Error Massage : %v", err)
	// 	c.RenderWithErr(c.Tr("ブロックチェーンへの登録中にエラーが発生し、失敗しました"), SETTINGS_OPTIONS, &f)
	// }

}
