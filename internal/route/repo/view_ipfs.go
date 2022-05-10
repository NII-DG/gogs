package repo

import (
	"bytes"
	"path"
	"time"

	"github.com/gogs/git-module"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/gitutil"
	"github.com/NII-DG/gogs/internal/markup"
	"github.com/NII-DG/gogs/internal/tool"
)

func renderDirectoryFromBcapi(c *context.Context, treeLink string) {
	tree, err := c.Repo.Commit.Subtree(c.Repo.TreePath)
	if err != nil {
		c.NotFoundOrError(gitutil.NewError(err), "get subtree")
		return
	}

	entries, err := tree.Entries()
	if err != nil {
		c.Error(err, "list entries")
		return
	}
	entries.Sort()

	currentFolederPath := c.Repo.RepoLink + "/" + c.Repo.BranchName
	if c.Repo.TreePath != "" {
		tmpPath := &currentFolederPath
		*tmpPath = *tmpPath + "/" + c.Repo.TreePath
	}
	//指定したフォルダーがデータセットであるか確認及び真の場合トークン情報を取得
	res, err := bcapi.GetDatasetInfoByLocation(c.User.Name, currentFolederPath)
	if err != nil {
		c.Error(err, "exchange BC-API")
		return
	}
	var inputAddress string
	var srcCodeAddress string
	var outputAddress string
	isDataset := false
	if res.DatasetLocation != "" {
		tmpInput := &inputAddress
		*tmpInput = res.InputAddress
		tmpSrc := &srcCodeAddress
		*tmpSrc = res.SrcCodeAddress
		tmpOut := &outputAddress
		*tmpOut = res.OutputAddress
		tmpIsDataset := &isDataset
		*tmpIsDataset = true
	}

	//フォルダーおよびファイルコンテンツアドレスを取得
	altFileDataList := []AltEntryCommitInfo{}
	var filesDataList []*git.EntryCommitInfo
	filesDataList, err = entries.CommitsInfo(c.Repo.Commit, git.CommitsInfoOptions{
		Path:           c.Repo.TreePath,
		MaxConcurrency: conf.Repository.CommitsFetchConcurrency,
		Timeout:        5 * time.Minute,
	})
	if err != nil {
		c.Error(err, "get commits info")
		return
	}

	resList, err := bcapi.GetContentByFolder(c.User.Name, currentFolederPath)
	if err != nil {
		c.Error(err, "list entries")
		return
	}

	for _, data := range filesDataList {
		flg := false
		if data.Entry.Type() == git.ObjectBlob {
			for _, resData := range resList.ContentsInFolder {
				fullPath := currentFolederPath + "/" + data.Entry.Name()
				if fullPath == resData.ContentLocation {
					altFileDataList = append(altFileDataList, AltEntryCommitInfo{
						Entry:          data.Entry,
						Index:          data.Index,
						Commit:         data.Commit,
						Submodule:      data.Submodule,
						ContentAddress: resData.IpfsCid,
					})
					tmpFlg := &flg
					*tmpFlg = true
				}
			}
		} else if data.Entry.Type() == git.ObjectTree && isDataset {
			inputTreePath := res.DatasetLocation + "/" + db.INPUT_FOLDER_NM
			srcTreePath := res.DatasetLocation + "/" + db.SRC_FOLDER_NM
			outputTreePath := res.DatasetLocation + "/" + db.OUTPUT_FOLDER_NM
			treePath := currentFolederPath + "/" + data.Entry.Name()
			switch treePath {
			case inputTreePath:
				altFileDataList = append(altFileDataList, AltEntryCommitInfo{
					Entry:          data.Entry,
					Index:          data.Index,
					Commit:         data.Commit,
					Submodule:      data.Submodule,
					ContentAddress: inputAddress,
				})
				tmpFlg := &flg
				*tmpFlg = true
			case srcTreePath:
				altFileDataList = append(altFileDataList, AltEntryCommitInfo{
					Entry:          data.Entry,
					Index:          data.Index,
					Commit:         data.Commit,
					Submodule:      data.Submodule,
					ContentAddress: srcCodeAddress,
				})
				tmpFlg := &flg
				*tmpFlg = true
			case outputTreePath:
				altFileDataList = append(altFileDataList, AltEntryCommitInfo{
					Entry:          data.Entry,
					Index:          data.Index,
					Commit:         data.Commit,
					Submodule:      data.Submodule,
					ContentAddress: outputAddress,
				})
				tmpFlg := &flg
				*tmpFlg = true
			}
		}
		if !flg {
			altFileDataList = append(altFileDataList, AltEntryCommitInfo{
				Entry:          data.Entry,
				Index:          data.Index,
				Commit:         data.Commit,
				Submodule:      data.Submodule,
				ContentAddress: "",
			})
		}
	}

	c.Data["Files"] = altFileDataList

	if c.Data["HasDmpJson"].(bool) {
		readDmpJson(c)
	} else {
		bidingDmpSchemaList(c, "conf/dmp")
	}

	var readmeFile *git.Blob
	for _, entry := range entries {
		if entry.IsTree() || !markup.IsReadmeFile(entry.Name()) {
			continue
		}

		// TODO(unknwon): collect all possible README files and show with priority.
		readmeFile = entry.Blob()
		break
	}

	if readmeFile != nil {
		c.Data["RawFileLink"] = ""
		c.Data["ReadmeInList"] = true
		c.Data["ReadmeExist"] = true

		p, err := readmeFile.Bytes()
		if err != nil {
			c.Error(err, "read file")
			return
		}

		// GIN mod: Replace existing buffer p and reader with annexed content buffer
		p, err = resolveAnnexedContent(c, p)
		if err != nil {
			return
		}

		isTextFile := tool.IsTextFile(p)
		c.Data["IsTextFile"] = isTextFile
		c.Data["FileName"] = readmeFile.Name()
		if isTextFile {
			switch markup.Detect(readmeFile.Name()) {
			case markup.MARKDOWN:
				c.Data["IsMarkdown"] = true
				p = markup.Markdown(p, treeLink, c.Repo.Repository.ComposeMetas())
			case markup.ORG_MODE:
				c.Data["IsMarkdown"] = true
				p = markup.OrgMode(p, treeLink, c.Repo.Repository.ComposeMetas())
			case markup.IPYTHON_NOTEBOOK:
				c.Data["IsIPythonNotebook"] = true
				c.Data["RawFileLink"] = c.Repo.RepoLink + "/raw/" + path.Join(c.Repo.BranchName, c.Repo.TreePath, readmeFile.Name())
			default:
				p = bytes.Replace(p, []byte("\n"), []byte(`<br>`), -1)
			}
			c.Data["FileContent"] = string(p)
		}
	}

	// Show latest commit info of repository in table header,
	// or of directory if not in root directory.
	latestCommit := c.Repo.Commit
	if len(c.Repo.TreePath) > 0 {
		latestCommit, err = c.Repo.Commit.CommitByPath(git.CommitByRevisionOptions{Path: c.Repo.TreePath})
		if err != nil {
			c.Error(err, "get commit by path")
			return
		}
	}
	c.Data["LatestCommit"] = latestCommit
	c.Data["LatestCommitUser"] = db.ValidateCommitWithEmail(latestCommit)

	if c.Repo.CanEnableEditor() {
		c.Data["CanAddFile"] = true
		c.Data["CanUploadFile"] = conf.Repository.Upload.Enabled
	}
}
