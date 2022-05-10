package repo

import (
	"bytes"
	"fmt"
	gotemplate "html/template"
	"path"
	"strings"
	"time"

	"github.com/G-Node/libgin/libgin/annex"
	"github.com/gogs/git-module"
	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/bcapi"
	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/gitutil"
	"github.com/NII-DG/gogs/internal/markup"
	"github.com/NII-DG/gogs/internal/template"
	"github.com/NII-DG/gogs/internal/template/highlight"
	"github.com/NII-DG/gogs/internal/tool"
	logv2 "unknwon.dev/clog/v2"
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

func renderFileFromIPFS(c *context.Context, entry *git.TreeEntry, treeLink, rawLink string, contentLocation string) {
	c.Data["IsViewFile"] = true

	blob := entry.Blob()
	p, err := blob.Bytes()
	if err != nil {
		c.Error(err, "read blob")
		return
	}

	c.Data["FileSize"] = blob.Size()
	c.Data["FileName"] = blob.Name()
	c.Data["HighlightClass"] = highlight.FileNameToHighlightClass(blob.Name())
	c.Data["RawFileLink"] = rawLink + "/" + c.Repo.TreePath

	// GIN mod: Replace existing buffer p with annexed content buffer (only if
	// it's an annexed ptr file)
	p, err = resolveAnnexedContentFromIPFS(c, p, contentLocation)
	if err != nil {
		return
	}
	isTextFile := tool.IsTextFile(p)
	c.Data["IsTextFile"] = isTextFile

	// Assume file is not editable first.
	if !isTextFile {
		c.Data["EditFileTooltip"] = c.Tr("repo.editor.cannot_edit_non_text_files")
	}

	canEnableEditor := c.Repo.CanEnableEditor()
	switch {
	case isTextFile:
		// GIN mod: Use c.Data["FileSize"] which is replaced by annexed content
		// size in resolveAnnexedContent() when necessary
		logv2.Trace("[FileSize] %v", c.Data["FileSize"].(int64))
		if c.Data["FileSize"].(int64) >= conf.UI.MaxDisplayFileSize {
			c.Data["IsFileTooLarge"] = true
			break
		}

		c.Data["ReadmeExist"] = markup.IsReadmeFile(blob.Name())

		switch markup.Detect(blob.Name()) {
		case markup.MARKDOWN:
			c.Data["IsMarkdown"] = true
			c.Data["FileContent"] = string(markup.Markdown(p, path.Dir(treeLink), c.Repo.Repository.ComposeMetas()))
		case markup.ORG_MODE:
			c.Data["IsMarkdown"] = true
			c.Data["FileContent"] = string(markup.OrgMode(p, path.Dir(treeLink), c.Repo.Repository.ComposeMetas()))
		case markup.IPYTHON_NOTEBOOK:
			c.Data["IsIPythonNotebook"] = true
			// GIN mod: JSON, YAML, and odML render support with jsTree
		case markup.JSON:
			c.Data["IsJSON"] = true
			c.Data["RawFileContent"] = string(p)
			fallthrough
		case markup.YAML:
			c.Data["IsYAML"] = true
			c.Data["RawFileContent"] = string(p)
			fallthrough
		case markup.XML:
			// pass XML down to ODML checker
			fallthrough
		case markup.ODML:
			if tool.IsODMLFile(p) {
				c.Data["IsODML"] = true
				c.Data["ODML"] = string(markup.MarshalODML(p))
			}
			fallthrough
		default:
			// Building code view blocks with line number on server side.
			var fileContent string
			if err, content := template.ToUTF8WithErr(p); err != nil {
				if err != nil {
					log.Error("ToUTF8WithErr: %s", err)
				}
				fileContent = string(p)
			} else {
				fileContent = content
			}

			var output bytes.Buffer
			lines := strings.Split(fileContent, "\n")
			// Remove blank line at the end of file
			if len(lines) > 0 && len(lines[len(lines)-1]) == 0 {
				lines = lines[:len(lines)-1]
			}
			// > GIN
			if len(lines) > conf.UI.MaxLineHighlight {
				c.Data["HighlightClass"] = "nohighlight"
			}
			// < GIN
			for index, line := range lines {
				output.WriteString(fmt.Sprintf(`<li class="L%d" rel="L%d">%s</li>`, index+1, index+1, gotemplate.HTMLEscapeString(strings.TrimRight(line, "\r"))) + "\n")
			}
			c.Data["FileContent"] = gotemplate.HTML(output.String())

			output.Reset()
			for i := 0; i < len(lines); i++ {
				output.WriteString(fmt.Sprintf(`<span id="L%d">%d</span>`, i+1, i+1))
			}
			c.Data["LineNums"] = gotemplate.HTML(output.String())
		}

		isannex := tool.IsAnnexedFile(p)
		if canEnableEditor && !isannex {
			c.Data["CanEditFile"] = true
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.edit_this_file")
		} else if !c.Repo.IsViewBranch {
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.must_be_on_a_branch")
		} else if !c.Repo.IsWriter() {
			c.Data["EditFileTooltip"] = c.Tr("repo.editor.fork_before_edit")
		}

	case tool.IsPDFFile(p) && (c.Data["FileSize"].(int64) < conf.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsPDFFile"] = true
	case tool.IsVideoFile(p) && (c.Data["FileSize"].(int64) < conf.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsVideoFile"] = true
	case tool.IsImageFile(p) && (c.Data["FileSize"].(int64) < conf.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsImageFile"] = true
	case tool.IsAnnexedFile(p) && (c.Data["FileSize"].(int64) < conf.Repository.RawCaptchaMinFileSize*annex.MEGABYTE ||
		c.IsLogged):
		c.Data["IsAnnexedFile"] = true
	}

	if canEnableEditor {
		c.Data["CanDeleteFile"] = true
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.delete_this_file")
	} else if !c.Repo.IsViewBranch {
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.must_be_on_a_branch")
	} else if !c.Repo.IsWriter() {
		c.Data["DeleteFileTooltip"] = c.Tr("repo.editor.must_have_write_access")
	}
}
