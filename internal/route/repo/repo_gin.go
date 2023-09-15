package repo

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	log "unknwon.dev/clog/v2"

	"github.com/unknwon/com"
	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/tool"
	"github.com/NII-DG/gogs/internal/process"
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
	var f repoUtil
	generateMaDmp(c, f)
}

// generateMaDmp is RCOS specific code.
// This generates maDMP(machine actionable DMP) based on
// DMP information created by the user in the repository.
func generateMaDmp(c context.AbstructContext, f AbstructRepoUtil) {
	// GitHubテンプレートNotebookを取得
	// refs: 1. https://zenn.dev/snowcait/scraps/3d51d8f7841f0c
	//       2. https://qiita.com/taizo/items/c397dbfed7215969b0a5
	templateUrl := getTemplateUrl() + "maDMP.ipynb"

	var decodedMaDmp string
	src, err := f.FetchContentsOnGithub(c, templateUrl)
	if err != nil && !c.IsInternalError() {
		log.Error("maDMP blob could not be fetched: %v", err)
		c.Redirect(c.GetRepo().GetRepoLink())
		return
	} else if err != nil && c.IsInternalError() {
		log.Error("maDMP blob could not be fetched: %v", err)
		c.Error(fmt.Errorf(c.Tr("rcos.server.error")), "")
		return
	}

	decodedMaDmp, err = f.DecodeBlobContent(src)
	if err != nil {
		log.Error("maDMP blob could not be decorded: %v", err)
		failedGenereteMaDmp(c, c.Tr("rcos.madmp.error.fetch"))
		return
	}

	/* DMPの内容によって、DockerFileを利用しないケースがあったため、
	　 DMPの内容を取得した後に、DockerFileを取得するように修正 */
	// コード付帯機能の起動時間短縮のための暫定的な定義

	// ユーザが作成したDMP情報取得
	entry, err := c.GetRepo().GetCommit().Blob("/dmp.json")
	if err != nil || entry == nil {
		log.Error("dmp.json blob could not be retrieved: %v", err)

		failedGenereteMaDmp(c, c.Tr("rcos.madmp.error.read"))
		return
	}
	buf, err := entry.Bytes()
	if err != nil {
		log.Error("dmp.json data could not be read: %v", err)

		failedGenereteMaDmp(c, c.Tr("rcos.madmp.error.read"))
		return
	}

	var dmp interface{}
	err = json.Unmarshal(buf, &dmp)
	if err != nil {
		log.Error("Unmarshal DMP info: %v", err)

		failedGenereteMaDmp(c, c.Tr("rcos.madmp.error.read"))
		return
	}

	// dmp.jsonに"fields"プロパティがある想定
	property := []string{"workflowIdentifier", "contentSize", "datasetStructure", "useDocker"}
	/* maDMPへ埋め込む情報を追加する際は
	上記リストに追加すること
	e.g.
	, hasGrdm
	*/
	selected := make(map[string]interface{})
	var errProperty string
	for _, v := range property {
		selected[v] = dmp.(map[string]interface{})[v]
		// Check if the value is entered
		if len(selected[v].(string)) == 0 {
			if len(errProperty) == 0 {
				errProperty = v
			} else {
				errProperty = errProperty + ", " + v
			}
		}
	}
	if len(errProperty) > 0 {
		failedDmp(c, c.Tr("rcos.dmp.error", errProperty))
		return
	}

	pathToMaDmp := "maDMP.ipynb"
	err = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
		LastCommitID: c.GetRepo().GetLastCommitIdStr(),
		OldBranch:    c.GetRepo().GetBranchName(),
		NewBranch:    c.GetRepo().GetBranchName(),
		OldTreeName:  "",
		NewTreeName:  pathToMaDmp,
		Message:      "[GIN] Generate maDMP",
		Content: fmt.Sprintf(
			decodedMaDmp,                   // この行が埋め込み先: maDMP
			selected["workflowIdentifier"], // ここより以下は埋め込む値: DMP情報
			selected["contentSize"],
			selected["datasetStructure"],
			selected["useDocker"],
			/* maDMPへ埋め込む情報を追加する際は
			ここに追記のこと
			e.g.
			selected["hasGrdm"] */
		),
		IsNewFile: true,
	})
	if err != nil {
		log.Error("failed generating maDMP: %v", err)
		failedGenereteMaDmp(c, c.Tr("rcos.madmp.error.exist"))
		return
	}

	/* Dockerfileか、binderフォルダを取得する。 */
	if selected["useDocker"] == "YES" {
		/* dockerファイルを取得する */
		fetchDockerfile(c)
	} else {
		/* binderフォルダ配下の環境構成ファイルを取得する */
		fetchEmviromentfile(c)
	}

	/* 共通で使用する imageファイルを取得する */
	fetchImagefile(c)

	c.GetFlash().Success(c.Tr("rcos.madmp.success"))
	c.Redirect(c.GetRepo().GetRepoLink() + "/container")
}

type AbstructRepoUtil interface {
	FetchContentsOnGithub(c context.AbstructContext, blobPath string) ([]byte, error)
	DecodeBlobContent(blobInfo []byte) (string, error)
}

type repoUtil func()

func (f repoUtil) FetchContentsOnGithub(c context.AbstructContext, blobPath string) ([]byte, error) {
	apiToken := conf.DG.ApiToken
	if conf.DG.MaDMPTemplateRepoBranch != "" {
		blobPath = blobPath + fmt.Sprintf("?ref=%s", conf.DG.MaDMPTemplateRepoBranch)
		return f.fetchContentsOnGithub(c, blobPath, apiToken)
	}
	return f.fetchContentsOnGithub(c, blobPath, apiToken)
}

func (f repoUtil) DecodeBlobContent(blobInfo []byte) (string, error) {
	return f.decodeBlobContent(blobInfo)
}

// FetchContentsOnGithub is RCOS specific code.
// This uses the Github API to retrieve information about the file
// specified in the argument, and returns it in the type of []byte.
// If any processing fails, it will return error.
// refs: https://docs.github.com/en/rest/reference/repos#contents
func (f repoUtil) fetchContentsOnGithub(c context.AbstructContext, blobPath string, apiToken string) ([]byte, error) {
	req, err := http.NewRequest("GET", blobPath, nil)
	if err != nil {
		c.CallData()["IsInternalError"] = true
		return nil, fmt.Errorf("do not Create Request. blobPath : %s, Error Msg : %v", blobPath, err)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")

	bearerToken := fmt.Sprintf("Bearer %s", apiToken)
	// When token is set, Github API rate limit increase.
	req.Header.Set("Authorization", bearerToken)

	client := new(http.Client)

	resp, err := client.Do(req)
	if err != nil {
		c.CallData()["IsInternalError"] = true
		return nil, fmt.Errorf("do not Request. blobPath : %s, Error Msg : %v", blobPath, err)
	}
	defer resp.Body.Close()
	log.Trace("Github api rate limit Remaining : %s", resp.Header.Values("X-RateLimit-Remaining")[0])

	if resp.StatusCode == http.StatusNotFound {
		c.CallData()["IsInternalError"] = true
		return nil, fmt.Errorf("blob not found. blobPath : %s, Error Msg : %v", blobPath, err)
	} else if resp.StatusCode == http.StatusUnauthorized {
		c.CallData()["IsInternalError"] = true
		return nil, fmt.Errorf("failure Authorization bacause Github API Token is invalid. blobPath : %s, Error Msg : %v", blobPath, err)
	} else if resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("failure Request for GitHub bacause Github API rate limit exceeded blobPath : %s, Error Msg : %v", blobPath, err)
	}

	contents, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.CallData()["IsInternalError"] = true
		return nil, fmt.Errorf("connot Read Response Body. blobPath : %s, Error Msg : %v", blobPath, err)
	}

	return contents, nil
}

// DecodeBlobContent is RCOS specific code.
// This reads and decodes "content" value of the response byte slice
// retrieved from the GitHub API.
// refs: https://docs.github.com/en/rest/reference/repos#contents
func (f repoUtil) decodeBlobContent(blobInfo []byte) (string, error) {
	var blob interface{}
	err := json.Unmarshal(blobInfo, &blob)
	if err != nil {
		return "", err
	}

	raw := blob.(map[string]interface{})["content"]
	decodedBlobContent, err := base64.StdEncoding.DecodeString(raw.(string))
	if err != nil {
		return "", err
	}

	return string(decodedBlobContent), nil
}

// failedGenerateMaDmp is RCOS specific code.
// This is a function used by GenerateMaDmp to emit an error message
// on UI when maDMP generation fails.
func failedGenereteMaDmp(c context.AbstructContext, msg string) {
	c.GetFlash().Error(msg)
	c.Redirect(c.GetRepo().GetRepoLink())
}

func failedDmp(c context.AbstructContext, msg string) {
	c.GetFlash().Error(msg)
	c.Redirect(c.GetRepo().GetRepoLink() + "/_edit/" + c.GetRepo().GetBranchName() + "/dmp.json")
}

// fetchDockerfile is RCOS specific code.
// This fetches the Dockerfile used when launching Binderhub.
func fetchDockerfile(c context.AbstructContext) {
	// コード付帯機能の起動時間短縮のための暫定的な定義
	dockerfileUrl := getTemplateUrl() + "Dockerfile"

	var f repoUtil
	src, err := f.FetchContentsOnGithub(c, dockerfileUrl)
	if err != nil {
		log.Error("Dockerfile could not be fetched: %v", err)
		failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(Dockerfile)")
		return
	}

	decodedDockerfile, err := f.DecodeBlobContent(src)
	if err != nil {
		log.Error("Dockerfile could not be decorded: %v", err)
		failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(Dockerfile)")
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
		Content:      decodedDockerfile,
		IsNewFile:    true,
	})
}

//★
// fetchEmviromentfile is RCOS specific code.
// This fetches the Dockerfile used when launching Binderhub.
func fetchEmviromentfile(c context.AbstructContext) {
	// コード付帯機能の起動時間短縮のための暫定的な定義
	Emviromentfilepath := getTemplateUrl() + "binder/"

	var f repoUtil

	Emviromentfile := []string{"apt.txt", "postBuild"}

	for i := 0; i < len(Emviromentfile); i++ {
		path := Emviromentfilepath + Emviromentfile[i]
		src, err := f.FetchContentsOnGithub(c, path)
		if err != nil {
			log.Error("%s could not be fetched: %v", Emviromentfile[i], err)
			failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(Emviromentfile)")
			return
		}

		decodefile, err := f.DecodeBlobContent(src)
		if err != nil {
			log.Error("%s could not be decorded: %v", Emviromentfile[i], err)

			failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(Emviromentfile)")
			return
		}

		treeName := "binder/" + Emviromentfile[i]
		message := "[GIN] fetch " + Emviromentfile[i]
		_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
			LastCommitID: c.GetRepo().GetLastCommitIdStr(),
			OldBranch:    c.GetRepo().GetBranchName(),
			NewBranch:    c.GetRepo().GetBranchName(),
			OldTreeName:  "",
			NewTreeName:  treeName,
			Message:      message,
			Content:      decodefile,
			IsNewFile:    true,
		})
	}

}

// fetchImagefile is RCOS specific code.
func fetchImagefile(c context.AbstructContext) {

	ImageFilePath := getTemplateUrl() + "images/"

	var f repoUtil

	ImageFile := []string{} //add the image file name if maDMP.ipynb has images.

	for i := 0; i < len(ImageFile); i++ {
		path := ImageFilePath + ImageFile[i]
		src, err := f.FetchContentsOnGithub(c, path)
		if err != nil {
			log.Error("%s could not be fetched: %v", ImageFile[i], err)
			failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(ImageFile)")
			return
		}

		decodefile, err := f.DecodeBlobContent(src)
		if err != nil {
			log.Error("%s could not be decorded: %v", ImageFile[i], err)

			failedGenereteMaDmp(c, "Sorry, failed generate maDMP: fetching template failed(ImageFile)")
			return
		}

		treeName := "images/" + ImageFile[i]
		message := "[GIN] fetch " + ImageFile[i]
		_ = c.GetRepo().GetDbRepo().UpdateRepoFile(c.GetUser(), db.UpdateRepoFileOptions{
			LastCommitID: c.GetRepo().GetLastCommitIdStr(),
			OldBranch:    c.GetRepo().GetBranchName(),
			NewBranch:    c.GetRepo().GetBranchName(),
			OldTreeName:  "",
			NewTreeName:  treeName,
			Message:      message,
			Content:      decodefile,
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
		decodeURL, _ := url.QueryUnescape(src_download_url.String())
		ctx.Data["WebContentUrl"] = decodeURL
		ctx.Data["IsOtherRepositoryContent"] = true
	} else {
		// S3などGIN-fork以外のインターネット上に実データがある場合
		decodeURL, _ := url.QueryUnescape(string(download_url))
		ctx.Data["WebContentUrl"] = decodeURL
		ctx.Data["IsWebContent"] = true
	}
	return err
}

// canEditFile is RCOS specific code.
// this returns whether file can be edited and file path can be edited.
func canEditFile(filePath string) (canEditFile bool, canEditFilePath bool) {

	canEditFile = true
	canEditFilePath = true
	cannotEditSet := make(map[string]bool)
	cannotEditSet[".gitattributes"] = true
	cannotEditSet[".gitignore"] = true
	cannotEditSet[".repository_id"] = true
	cannotEditSet["maDMP.ipynb"] = true
	cannotEditSet[".datalad/.gitattributes"] = true
	cannotEditSet[".datalad/config"] = true
	cannotEditSet["validation_results/entity_ids.json"] = true
	cannotEditSet["validation_results/results.json"] = true
	cannotEditSet["validation_results/ro_crate.json"] = true
	canEditSet := make(map[string]bool)
	canEditSet["dmp.json"] = true
	canEditSet["Dockerfile"] = true
	canEditSet["README.md"] = true

	if cannotEditSet[filePath] {
		canEditFile = false
		canEditFilePath = false
		return canEditFile, canEditFilePath
	}
	if canEditSet[filePath] {
		canEditFile = true
		canEditFilePath = false
		return canEditFile, canEditFilePath
	}

	// experiment/[実験パッケージ名]配下のenvironment.yml, README.md, requirement.txt, Snakefileにマッチ
	re1 := regexp.MustCompile(`^experiments/.*?/(environment\.yml|README\.md|requirements\.txt|Snakefile)$`)
	if re1.MatchString(filePath) {
		canEditFile = true
		canEditFilePath = false
		return canEditFile, canEditFilePath
	}

	// experiment/[実験パッケージ名]配下の.gitkeepにマッチ
	re2 := regexp.MustCompile(`^experiments/.*?/\.gitkeep$`)

	// WORKFLOWS配下の全てにマッチ
	re3 := regexp.MustCompile(`^WORKFLOWS/.*?$`)
	if re2.MatchString(filePath) || re3.MatchString(filePath) {
		canEditFile = false
		canEditFilePath = false
		return canEditFile, canEditFilePath
	}

	return canEditFile, canEditFilePath
}

// SetupResearchFlow is RCOS specific code.
func SetupResearchFlow(c context.Context) {
	setupResearchFlow(c)
}

// setupResearchFlow is RCOS specific code.
// This setup Researchflow
// Researchflow created by the user in the repository.
func setupResearchFlow(c context.Context) {

	log.Trace("Start setupResearchFlow" )
	// テンプレートの取得場所を決定する。
	// テンプレートの取得場所は、custom/conf/app.iniに記載する
	// RFTemplateRepoURL = ResearchFlowの取得先(リポジトリのURL)
	// RFTemplateRepoBranch = リポジトリのブランチ
	var err error
	rfRepoUrl := conf.DG.RFTemplateRepoURL
	rfRepoBranch := conf.DG.RFTemplateRepoBranch
	repoPath := c.Repo.Repository.RepoPath()
	tmpBasePath := filepath.Join(conf.Server.AppDataPath, "tmp", "repos", com.ToStr(time.Now().Nanosecond())+".git")

	//*******************************************************************
	// git archiveコマンドを実行するとエラーとなる。
	// 「GIT_PROXY_COMMAND environment variable are NOT set.」
	// → git archiveはローカルリポジトリに発行するコマンドとのこと。
	// そのため、tmpフォルダを作る方向で修正。
	//*******************************************************************

	// Create temporary directory to store temporary copy of the ResearchFlow
	// and clean it up when operation finished regardless of succeed or not.
	// this code is copy from internal->db->pull.go Merge()
	if err = os.MkdirAll(filepath.Dir(tmpBasePath), os.ModePerm); err != nil {
		log.Trace("Fail mkdir  Path=[%s] %s", tmpBasePath, err )
		return
	}
	defer func() {
		_ = os.RemoveAll(filepath.Dir(tmpBasePath))
	}()

	// Clone the Researchflow repository to the defined temporary directory,
	var stderr string
	if _, stderr, err = process.ExecTimeout(5*time.Minute,
		fmt.Sprintf("git clone: %s %s", rfRepoUrl, tmpBasePath),
		"git", "clone", "-b", rfRepoBranch, rfRepoUrl, tmpBasePath); err != nil {
		log.Trace("Fail git clone Path=[%s] Branch=[%s] URL=[%s] %s", tmpBasePath, rfRepoBranch, rfRepoUrl, stderr )
		return
	}

	// ".git" dir is unnecessary. 
	if err := os.Remove( filepath.Join( tmpBasePath, ".git" )); err != nil {
		log.Trace("Fail RemoveDir Path=[%s] %s", filepath.Join( tmpBasePath, ".git" ), err )
		return
	}

	// Copy uploaded files into repository
	srcDir := tmpBasePath
	dstDir := repoPath
	err = filepath.Walk( srcDir, func(path string, info os.FileInfo , err error ) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			relativePath, err := filepath.Rel(srcDir, path)
			if err != nil {
				return err
			}

			dstPath := filepath.Join( dstDir, relativePath)

			if err := os.MkdirAll( filepath.Dir(dstPath), os.ModePerm); err != nil {
				return err
			}

			if err = com.Copy( path, dstPath); err != nil {
				return err
			}

			log.Trace("Copied %s to %s\n", path, dstPath)
		}

		return nil
	})

	if err != nil {
		log.Trace("Error: %s ", err)
		return
	}

	// Repository commit
	if err = git.RepoAdd( repoPath, git.AddOptions{All: true}); err != nil {
		log.Trace("git add --all: %v", err)
		return 
	} 
	
	if err = git.RepoCommit( repoPath, c.User.NewGitSig(), "ResearchFlow Download"); err != nil {
		log.Trace("commit changes on %q: %v", repoPath, err)
		return
	}


	envs := db.ComposeHookEnvs(db.ComposeHookEnvsOptions{
		AuthUser:  c.User,
		OwnerName: c.Repo.Repository.MustOwner().Name,
		OwnerSalt: c.Repo.Repository.MustOwner().Salt,
		RepoID:    c.Repo.Repository.ID,
		RepoName:  c.Repo.Repository.Name,
		RepoPath:  c.Repo.Repository.RepoPath(),
	})

	if err = git.RepoPush( repoPath, "origin", "master", git.PushOptions{Envs: envs}); err != nil {
		log.Trace("git push origin %s: %v", "master", err)
		return
		
	}

	// dockerfileもダウンロードする必要がある。

	c.GetFlash().Success(c.Tr("rcos.madmp.success"))

	// リポジトリTOP画面
	c.Redirect(c.GetRepo().GetRepoLink())
}