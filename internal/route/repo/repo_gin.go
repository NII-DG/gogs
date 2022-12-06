package repo

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"

	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/tool"
	"github.com/gogs/git-module"
)

func serveAnnexedData(ctx *context.Context, name string, buf []byte) error {
	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]
	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(ctx.Repo.Repository.RepoPath())
	if err != nil {
		log.Error("Failed to find content location for file %q with key %q", name, key)
		return err
	}
	// always trim space from output for git command
	contentPath = bytes.TrimSpace(contentPath)
	return serveAnnexedKey(ctx, name, string(contentPath))
}

func serveAnnexedKey(ctx *context.Context, name string, contentPath string) error {
	fullContentPath := filepath.Join(ctx.Repo.Repository.RepoPath(), contentPath)
	annexfp, err := os.Open(fullContentPath)
	if err != nil {
		log.Error("Failed to open annex file at %q: %s", fullContentPath, err.Error())
		return err
	}
	defer annexfp.Close()
	annexReader := bufio.NewReader(annexfp)

	info, err := annexfp.Stat()
	if err != nil {
		log.Error("Failed to stat file at %q: %s", fullContentPath, err.Error())
		return err
	}

	buf, _ := annexReader.Peek(1024)

	ctx.Resp.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size()))
	if !tool.IsTextFile(buf) {
		if !tool.IsImageFile(buf) {
			ctx.Resp.Header().Set("Content-Disposition", "attachment; filename=\""+name+"\"")
			ctx.Resp.Header().Set("Content-Transfer-Encoding", "binary")
		}
	} else if !conf.Repository.EnableRawFileRenderMode || !ctx.QueryBool("render") {
		ctx.Resp.Header().Set("Content-Type", "text/plain; charset=utf-8")
	}

	log.Trace("Serving annex content for %q: %q", name, contentPath)
	if ctx.Req.Method == http.MethodHead {
		// Skip content copy when request method is HEAD
		log.Trace("Returning header: %+v", ctx.Resp.Header())
		return nil
	}
	_, err = io.Copy(ctx.Resp, annexReader)
	return err
}

// readDmpJson is RCOS specific code.
func readDmpJson(c context.AbstructContext) {
	log.Trace("Reading dmp.json file")
	entry, err := c.GetRepo().GetCommit().Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error("dmp.json blob could not be retrieved: %v", err)
		c.CallData()["HasDmpJson"] = false
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error("dmp.json data could not be read: %v", err)
		c.CallData()["HasDmpJson"] = false
		return
	}
	c.CallData()["DOIInfo"] = string(buf)
}

// GenerateMaDmp is RCOS specific code.
func GenerateMaDmp(c context.AbstructContext) {
	generateMaDmp(c)
}

// generateMaDmp is RCOS specific code.
// This generates maDMP(machine actionable DMP) based on
// DMP information created by the user in the repository.
func generateMaDmp(c context.AbstructContext) {
	// テンプレートNotebookを取得
	notebookPath := filepath.Join(getDgContentsPath(), "notebooks", "maDMP.ipynb")
	log.Trace("[RCOS] Getting maDMP.ipynb, file path : %v", notebookPath)
	notebookSrc, err := ioutil.ReadFile(notebookPath)
	if err != nil {
		log.Error("Cannot Read File. file path : %v", notebookPath)
		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed")
		return
	}

	/* DMPの内容によって、DockerFileを利用しないケースがあったため、
	　 DMPの内容を取得した後に、DockerFileを取得するように修正 */
	// コード付帯機能の起動時間短縮のための暫定的な定義
	// fetchDockerfile(c)

	// ユーザが作成したDMP情報取得
	entry, err := c.GetRepo().GetCommit().Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error("dmp.json blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error("dmp.json data could not be read: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}

	var dmp interface{}
	err = json.Unmarshal(buf, &dmp)
	if err != nil {
		log.Error("Unmarshal DMP info: %v", err)

		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: DMP could not read")
		return
	}

	// dmp.jsonに"fields"プロパティがある想定
	selectedField := dmp.(map[string]interface{})["field"]
	selectedDataSize := dmp.(map[string]interface{})["dataSize"]
	selectedDatasetStructure := dmp.(map[string]interface{})["datasetStructure"]
	selectedUseDocker := dmp.(map[string]interface{})["useDocker"]
	/* maDMPへ埋め込む情報を追加する際は
	ここに追記のこと
	e.g.
	hasGrdm := dmp.(map[string]interface{})["hasGrdm"]
	*/

	pathToMaDmp := "maDMP.ipynb"
	err = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
		LastCommitID: c.GetRepo().GetLastCommitIdStr(),
		OldBranch:    c.GetRepo().GetBranchName(),
		NewBranch:    c.GetRepo().GetBranchName(),
		OldTreeName:  "",
		NewTreeName:  pathToMaDmp,
		Message:      "[GIN] Generate maDMP",
		Content: fmt.Sprintf(
			string(notebookSrc), // この行が埋め込み先: maDMP
			selectedField,       // ここより以下は埋め込む値: DMP情報
			selectedDataSize,
			selectedDatasetStructure,
			selectedUseDocker,
			/* maDMPへ埋め込む情報を追加する際は
			ここに追記のこと
			e.g.
			hasGrdm, */
		),
		IsNewFile: true,
	})
	if err != nil {
		log.Error("failed generating maDMP: %v", err)

		failedGenereteMaDmp(c, "Faild gerate maDMP: Already exist")
		return
	}

	/* Dockerfileか、binderフォルダを取得する。 */
	if selectedUseDocker == "YES" {
		/* dockerファイルを取得する */
		fetchDockerfile(c)
	} else {
		/* binderフォルダ配下の環境構成ファイルを取得する */
		fetchEmviromentfile(c)
	}

	/* 共通で使用する imageファイルを取得する */
	fetchImagefile(c)

	c.GetFlash().Success("maDMP generated!")
	c.Redirect(c.GetRepo().GetRepoLink())
}

// // failedGenerateMaDmp is RCOS specific code.
// // This is a function used by GenerateMaDmp to emit an error message
// // on UI when maDMP generation fails.
func failedGenereteMaDmp(c context.AbstructContext, msg string) {
	c.GetFlash().Error(msg)
	c.Redirect(c.GetRepo().GetRepoLink())
}

// fetchDockerfile is RCOS specific code.
// This fetches the Dockerfile used when launching Binderhub.
func fetchDockerfile(c context.AbstructContext) {
	// コード付帯機能の起動時間短縮のための暫定的な定義
	dockerFilePath := filepath.Join(getDgContentsPath(), "build_files", "Dockerfile")
	log.Trace("[RCOS] Getting Dockerfile, file path : %v", dockerFilePath)
	dockerFile, err := ioutil.ReadFile(dockerFilePath)
	if err != nil {
		log.Error("Cannot Read File. file path : %v", dockerFilePath)
		failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed(Dockerfile)")
		return
	}

	pathToDockerfile := "Dockerfile"
	_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
		LastCommitID: c.GetRepo().GetLastCommitIdStr(),
		OldBranch:    c.GetRepo().GetBranchName(),
		NewBranch:    c.GetRepo().GetBranchName(),
		OldTreeName:  "",
		NewTreeName:  pathToDockerfile,
		Message:      "[GIN] fetch Dockerfile",
		Content:      string(dockerFile),
		IsNewFile:    true,
	})
}

// fetchEmviromentfile is RCOS specific code.
// This fetches the Dockerfile used when launching Binderhub.
func fetchEmviromentfile(c context.AbstructContext) {
	// コード付帯機能の起動時間短縮のための暫定的な定義
	binderPath := filepath.Join(getDgContentsPath(), "build_files", "binder")
	log.Trace("[RCOS] Reading Directory, dir path : %v", binderPath)
	files, err := ioutil.ReadDir(binderPath)
	if err != nil {
		log.Error("Cannot Read Directory. dir path : %v", binderPath)
	}

	for _, file := range files {
		filePath := filepath.Join(getDgContentsPath(), "build_files", "binder", file.Name())
		log.Trace("[RCOS] Getting binder file, file path : %v", filePath)
		binderfile, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error("Cannot Read File. file path : %v", filePath)
			failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed(Emviromentfile)")
			return
		}

		treeName := "binder/" + file.Name()
		message := "[GIN] fetch " + file.Name()
		_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
			LastCommitID: c.GetRepo().GetLastCommitIdStr(),
			OldBranch:    c.GetRepo().GetBranchName(),
			NewBranch:    c.GetRepo().GetBranchName(),
			OldTreeName:  "",
			NewTreeName:  treeName,
			Message:      message,
			Content:      string(binderfile),
			IsNewFile:    true,
		})
	}
}

// fetchImagefile is RCOS specific code.
func fetchImagefile(c context.AbstructContext) {
	imagesPath := filepath.Join(getDgContentsPath(), "images", "create_ma_dmp")
	log.Trace("[RCOS] Reading Directory, dir path : %v", imagesPath)
	files, err := ioutil.ReadDir(imagesPath)
	if err != nil {
		log.Error("Cannot Read Directory. dir path : %v", imagesPath)
	}
	for _, file := range files {
		filePath := filepath.Join(imagesPath, file.Name())
		log.Trace("[RCOS] Getting image file, file path : %v", filePath)
		imagefile, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Error("Cannot Read File. file path : %v", filePath)
			failedGenereteMaDmp(c, "Sorry, faild gerate maDMP: fetching template failed(ImageFile)")
			return
		}
		treeName := "images/" + file.Name()
		message := "[GIN] fetch " + file.Name()
		_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
			LastCommitID: c.GetRepo().GetLastCommitIdStr(),
			OldBranch:    c.GetRepo().GetBranchName(),
			NewBranch:    c.GetRepo().GetBranchName(),
			OldTreeName:  "",
			NewTreeName:  treeName,
			Message:      message,
			Content:      string(imagefile),
			IsNewFile:    true,
		})

	}
}

// resolveAnnexedContent takes a buffer with the contents of a git-annex
// pointer file and an io.Reader for the underlying file and returns the
// corresponding buffer and a bufio.Reader for the underlying content file.
// The returned byte slice and bufio.Reader can be used to replace the buffer
// and io.Reader sent in through the caller so that any existing code can use
// the two variables without modifications.
// Any errors that occur during processing are stored in the provided context.
// The FileSize of the annexed content is also saved in the context (c.Data["FileSize"]).
func resolveAnnexedContent(c *context.Context, buf []byte) ([]byte, error) {
	if !tool.IsAnnexedFile(buf) {
		// not an annex pointer file; return as is
		return buf, nil
	}
	log.Trace("Annexed file requested: Resolving content for %q", bytes.TrimSpace(buf))

	keyparts := strings.Split(strings.TrimSpace(string(buf)), "/")
	key := keyparts[len(keyparts)-1]

	// get URL identify contents of the file on the Internet
	if strings.Contains(key, "URL") {
		err := getWebContentURL(c, key)
		return buf, err
	}

	contentPath, err := git.NewCommand("annex", "contentlocation", key).RunInDir(c.Repo.Repository.RepoPath())
	if err != nil {
		log.Error("Failed to find content location for key %q", key)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	// always trim space from output for git command
	contentPath = bytes.TrimSpace(contentPath)
	afp, err := os.Open(filepath.Join(c.Repo.Repository.RepoPath(), string(contentPath)))
	if err != nil {
		log.Trace("Could not open annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	info, err := afp.Stat()
	if err != nil {
		log.Trace("Could not stat annex file: %v", err)
		c.Data["IsAnnexedFile"] = true
		return buf, err
	}
	annexDataReader := bufio.NewReader(afp)
	annexBuf := make([]byte, 1024)
	n, _ := annexDataReader.Read(annexBuf)
	annexBuf = annexBuf[:n]
	c.Data["FileSize"] = info.Size()
	log.Trace("Annexed file size: %d B", info.Size())
	return annexBuf, nil
}

func GitConfig(c *context.Context) {
	log.Trace("RepoPath: %+v", c.Repo.Repository.RepoPath())
	configFilePath := path.Join(c.Repo.Repository.RepoPath(), "config")
	log.Trace("Serving file %q", configFilePath)
	if _, err := os.Stat(configFilePath); err != nil {
		c.Error(err, "GitConfig")
		// c.ServerError("GitConfig", err)
		return
	}
	c.ServeFileContent(configFilePath, "config")
}

func AnnexGetKey(c *context.Context) {
	filename := c.Params(":keyfile")
	key := c.Params(":key")
	contentPath := filepath.Join("annex/objects", c.Params(":hashdira"), c.Params(":hashdirb"), key, filename)
	log.Trace("Git annex requested key %q: %q", key, contentPath)
	err := serveAnnexedKey(c, filename, contentPath)
	if err != nil {
		c.Error(err, "AnnexGetKey")
	}
}

// getWebContentURL is RCOS specific code.
func getWebContentURL(ctx *context.Context, key string) error {
	subkey := &key
	// decode key --ref git://git-annex.branchable.com/ --dir Annex/Locations.hs
	*subkey = strings.Replace(key, "&a", "&", -1)
	key = strings.Replace(key, "&s", "%", -1)
	key = strings.Replace(key, "&c", ":", -1)
	key = strings.Replace(key, "%", "/", -1)
	// get URL
	location, err := git.NewCommand("annex", "whereis", "--key", key).RunInDir(ctx.Repo.Repository.RepoPath())
	start := strings.Index(string(location), "web: ")
	location = location[start+len("web: "):]
	end := strings.Index(string(location), "\n")
	download_url := location[:end]
	u, _ := url.Parse(string(download_url))
	if u.Hostname() == conf.Server.Domain {
		// GIN-forkの実データがaddurlされている場合は、実データファイルの閲覧画面をリンクする
		src_download_url := &url.URL{}
		src_download_url.Scheme = u.Scheme
		src_download_url.Host = u.Host
		src_download_url.Path = strings.Replace(u.Path, strings.Split(u.Path, "/")[3], "src", 1)
		ctx.Data["WebContentUrl"] = src_download_url.String()
		ctx.Data["IsOtherRepositoryContent"] = true
	} else {
		// S3などGIN-fork以外のインターネット上に実データがある場合
		ctx.Data["WebContentUrl"] = string(download_url)
		ctx.Data["IsWebContent"] = true
	}
	return err
}
