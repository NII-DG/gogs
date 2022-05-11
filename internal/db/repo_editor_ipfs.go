package db

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/unknwon/com"

	"github.com/gogs/git-module"

	"github.com/NII-DG/gogs/internal/annex_ipfs"
	encyrptfile "github.com/NII-DG/gogs/internal/encyrpt_file"
	"github.com/NII-DG/gogs/internal/osutil"

	log "unknwon.dev/clog/v2"
)

type UploadRepoFileOptionsForIPFS struct {
	LastCommitID  string
	OldBranch     string
	NewBranch     string
	TreePath      string
	Message       string
	Files         []string // In UUID format
	UpperRopoPath string   //RepoOwnerNm / RepoNm / branchNm
}

//TODO[2022-05-06]：プライベート or パブリック　レポジトリで異なる処理を行う
func (repo *Repository) UploadRepoFilesToIPFS(doer *User, opts UploadRepoFileOptionsForIPFS, isPrivate bool) (contentMap map[string]AnnexUploadInfo, err error) {
	//プライベート or パブリック　レポジトリで異なる処理を行う
	if isPrivate {
		log.Info("Private Repository Upload Files. User: %v", doer.FullName)
		//プライベートレポジトリの場合
		suberr := &err
		contentMap, *suberr = repo.PrivateUploadRepoFiles(doer, opts)
	} else {
		log.Info("Public Repository Upload Files. User: %v", doer.FullName)
		//パブリックレポジトリの場合
		suberr := &err
		contentMap, *suberr = repo.PublicUploadRepoFiles(doer, opts)
	}
	return contentMap, err
}

//パブリックレポジトリのUploadRepoFiles
func (repo *Repository) PublicUploadRepoFiles(doer *User, opts UploadRepoFileOptionsForIPFS) (contentMap map[string]AnnexUploadInfo, err error) {
	if len(opts.Files) == 0 {
		log.Error("Error 1: %v", len(opts.Files))
		return nil, nil
	}

	for _, fi := range opts.Files {
		log.Info("[opts.Files] %v", fi)
	}

	uploads, err := GetUploadsByUUIDs(opts.Files)
	if err != nil {
		return nil, fmt.Errorf("get uploads by UUIDs[%v]: %v", opts.Files, err)
	}

	repoWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer repoWorkingPool.CheckOut(com.ToStr(repo.ID))

	if err = repo.DiscardLocalRepoBranchChanges(opts.OldBranch); err != nil {
		return nil, fmt.Errorf("discard local repo branch[%s] changes: %v", opts.OldBranch, err)
	} else if err = repo.UpdateLocalCopyBranch(opts.OldBranch); err != nil {
		return nil, fmt.Errorf("update local copy branch[%s]: %v", opts.OldBranch, err)
	}

	if opts.OldBranch != opts.NewBranch {
		if err = repo.CheckoutNewBranch(opts.OldBranch, opts.NewBranch); err != nil {
			return nil, fmt.Errorf("checkout new branch[%s] from old branch[%s]: %v", opts.NewBranch, opts.OldBranch, err)
		}
	}

	localPath := repo.LocalCopyPath()
	dirPath := path.Join(localPath, opts.TreePath)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}
	// Copy uploaded files into repository
	for _, upload := range uploads {
		tmpPath := upload.LocalPath()
		if !osutil.IsFile(tmpPath) {
			continue
		}

		// Prevent copying files into .git directory, see https://gogs.io/gogs/issues/5558.
		if isRepositoryGitPath(upload.Name) { //upload.Name = ex. dataset1/src/main/src_data.txt
			continue
		}

		targetPath := path.Join(dirPath, upload.Name)
		// GIN: Create subdirectory for dirtree uploads
		if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("mkdir: %v", err)
		}

		//アップロードファイルをローカルレポジトリディレクトリにコピーする。
		log.Info("Copy %v TO %v", tmpPath, targetPath)
		if err = com.Copy(tmpPath, targetPath); err != nil {
			return nil, fmt.Errorf("copy: %v", err)
		}
	}

	annexSetupForIPFS(localPath) // Initialise annex and set configuration (with add filter for filesizes)
	var annexAddRes []annex_ipfs.AnnexAddResponse
	if annexAddRes, err = annexAddAndGetInfo(localPath, true); err != nil {
		return nil, fmt.Errorf("git annex add: %v", err)
	} else if err = git.RepoCommit(localPath, doer.NewGitSig(), opts.Message); err != nil {
		return nil, fmt.Errorf("commit changes on %q: %v", localPath, err)
	}

	envs := ComposeHookEnvs(ComposeHookEnvsOptions{
		AuthUser:  doer,
		OwnerName: repo.MustOwner().Name,
		OwnerSalt: repo.MustOwner().Salt,
		RepoID:    repo.ID,
		RepoName:  repo.Name,
		RepoPath:  repo.RepoPath(),
	})
	if err = git.RepoPush(localPath, "origin", opts.NewBranch, git.PushOptions{Envs: envs}); err != nil {
		return nil, fmt.Errorf("git push origin %s: %v", opts.NewBranch, err)
	}
	contentMap, err = PublicannexUpload(opts.UpperRopoPath, localPath, "ipfs", annexAddRes)
	if err != nil { // Copy new files
		return nil, fmt.Errorf("annex copy %s: %v", localPath, err)
	}
	AnnexUninit(localPath) // Uninitialise annex to prepare for deletion
	StartIndexing(*repo)   // Index the new data
	//localPathのディレクトリの削除
	if err := RemoveFilesFromLocalRepository(dirPath, uploads...); err != nil {
		return nil, err
	}

	return contentMap, DeleteUploads(uploads...)

}

//プライベートレポジトリのUploadRepoFiles
//TODO: ファイルを暗号化して取り扱う
func (repo *Repository) PrivateUploadRepoFiles(doer *User, opts UploadRepoFileOptionsForIPFS) (map[string]AnnexUploadInfo, error) {
	if len(opts.Files) == 0 {
		log.Error("Error 1: %v", len(opts.Files))
		return nil, nil
	}

	for _, fi := range opts.Files {
		log.Info("[opts.Files] %v", fi)
	}

	uploads, err := GetUploadsByUUIDs(opts.Files)
	if err != nil {
		return nil, fmt.Errorf("get uploads by UUIDs[%v]: %v", opts.Files, err)
	}

	repoWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer repoWorkingPool.CheckOut(com.ToStr(repo.ID))

	if err = repo.DiscardLocalRepoBranchChanges(opts.OldBranch); err != nil {
		return nil, fmt.Errorf("discard local repo branch[%s] changes: %v", opts.OldBranch, err)
	} else if err = repo.UpdateLocalCopyBranch(opts.OldBranch); err != nil {
		return nil, fmt.Errorf("update local copy branch[%s]: %v", opts.OldBranch, err)
	}

	if opts.OldBranch != opts.NewBranch {
		if err = repo.CheckoutNewBranch(opts.OldBranch, opts.NewBranch); err != nil {
			return nil, fmt.Errorf("checkout new branch[%s] from old branch[%s]: %v", opts.NewBranch, opts.OldBranch, err)
		}
	}

	localPath := repo.LocalCopyPath()
	dirPath := path.Join(localPath, opts.TreePath)
	if err = os.MkdirAll(dirPath, os.ModePerm); err != nil {
		return nil, err
	}

	annexSetupForIPFS(localPath) // Initialise annex and set configuration (with add filter for filesizes)

	uploadInfo := map[string]AnnexUploadInfo{} //ファイルパスとファイルハッシュ値

	// Copy uploaded files into repository
	for _, upload := range uploads {
		tmpPath := upload.LocalPath()
		if !osutil.IsFile(tmpPath) {
			continue
		}

		// Prevent copying files into .git directory, see https://gogs.io/gogs/issues/5558.
		if isRepositoryGitPath(upload.Name) {
			continue
		}

		//ハッシュ値を取得(git annex calckey)
		fullContentHash, err := annexCalcKey(localPath, tmpPath)
		if err != nil {
			return nil, fmt.Errorf("[Failure Calculating Hash From Target File : %v. Error Msg : %v] ", tmpPath, err)
		}

		targetPath := path.Join(dirPath, upload.Name)
		// GIN: Create subdirectory for dirtree uploads
		if err = os.MkdirAll(filepath.Dir(targetPath), os.ModePerm); err != nil {
			return nil, fmt.Errorf("mkdir: %v", err)
		}
		//ファイルの暗号化し、暗号化ファイルをIPFS上にアップロードする。
		address, err := encyrptfile.Encrypted(tmpPath, repo.Password)
		if err != nil {
			return nil, err
		}
		filePath := path.Join(opts.UpperRopoPath, upload.Name)
		//コンテンツ情報のインスタンスを定義
		uploadInfo[filePath] = AnnexUploadInfo{FullContentHash: fullContentHash, IpfsCid: address}
		//ハッシュ値にGit管轄ディレクトリに格納する。
		wFile, err := os.Create(targetPath)
		if err != nil {
			return nil, fmt.Errorf("[Failure Creating Hash File In %v. Error Msg : %v ]", targetPath, err)
		}
		defer wFile.Close()

		b := []byte(fullContentHash)
		_, err = wFile.Write(b)
		if err != nil {
			return nil, fmt.Errorf("[Failure Writting Hash In %v. Error Msg : %v ]", targetPath, err)
		}
	}

	var annexAddRes []annex_ipfs.AnnexAddResponse
	if annexAddRes, err = annexAddAndGetInfo(localPath, true); err != nil {
		return nil, fmt.Errorf("git annex add: %v", err)
	} else if err = git.RepoCommit(localPath, doer.NewGitSig(), opts.Message); err != nil {
		return nil, fmt.Errorf("commit changes on %q: %v", localPath, err)
	}

	envs := ComposeHookEnvs(ComposeHookEnvsOptions{
		AuthUser:  doer,
		OwnerName: repo.MustOwner().Name,
		OwnerSalt: repo.MustOwner().Salt,
		RepoID:    repo.ID,
		RepoName:  repo.Name,
		RepoPath:  repo.RepoPath(),
	})
	if err = git.RepoPush(localPath, "origin", opts.NewBranch, git.PushOptions{Envs: envs}); err != nil {
		return nil, fmt.Errorf("git push origin %s: %v", opts.NewBranch, err)
	}
	err = PrivateAnnexUpload(opts.UpperRopoPath, localPath, "ipfs", annexAddRes)
	if err != nil { // Copy new files
		return nil, fmt.Errorf("annex copy %s: %v", localPath, err)
	}
	AnnexUninit(localPath) // Uninitialise annex to prepare for deletion
	StartIndexing(*repo)   // Index the new data
	//localPathのディレクトリの削除
	if err := RemoveFilesFromLocalRepository(dirPath, uploads...); err != nil {
		return nil, err
	}

	return uploadInfo, DeleteUploads(uploads...)

}

type DatasetInfo struct {
	InputList  []ContentInfo
	SrcList    []ContentInfo
	OutputList []ContentInfo
}

type ContentInfo struct {
	File    string //ex : datasetNm/Folder/...../File
	Address string
}

// @value input
var INPUT_FOLDER_NM string = "input"

// @value src
var SRC_FOLDER_NM string = "src"

// @value output
var OUTPUT_FOLDER_NM string = "output"

func (repo *Repository) CloneRepo(branch string) (err error) {
	repoWorkingPool.CheckIn(com.ToStr(repo.ID))
	defer repoWorkingPool.CheckOut(com.ToStr(repo.ID))
	if err = repo.DiscardLocalRepoBranchChanges(branch); err != nil {
		return fmt.Errorf("discard local repo branch[%s] changes: %v", branch, err)
	} else if err = repo.LocalCopyBranch(branch); err != nil {
		return fmt.Errorf("update local copy branch[%s]: %v", branch, err)
	}
	return nil
}

func (repo *Repository) CheckDatasetFormat(datasetNmList []string) (err error) {
	//ローカルレポジトリの操作するためのディレクトリ取得
	localPath := repo.LocalCopyPath()

	//フォーマットのチェック
	for _, datasetNm := range datasetNmList {
		err = CheckDatasetFormat(localPath, datasetNm)
		if err != nil {
			return err
		}
	}
	log.Info("[OK Dataset Format] %s", datasetNmList)
	return nil
}

//データセットフォーマットのチェックとコンテンツアドレスの取得(map[stirng]DatasetInfo)
func (repo *Repository) GetContentAddress(datasetNmList []string, repoBranchNm string) (datasetNmToFileMap map[string]DatasetInfo, err error) {

	//ローカルレポジトリの操作するためのディレクトリ取得
	localPath := repo.LocalCopyPath()

	//ローカルのリポートリポジトリのIPFS有効化
	//ベアレポジトリをIPFSへ連携
	if _, err := git.NewCommand("annex", "enableremote", "ipfs").RunInDir(localPath); err != nil {
		return nil, fmt.Errorf("[Failure enable remote(ipfs)] err : %v, localPath : %v", err, localPath)
	} else {
		log.Info("[Success enable remote(ipfs)] repoPath : %v", localPath)
	}

	//コンテンツアドレスの取得
	datasetToContentsMap := map[string][]annex_ipfs.AnnexContentInfo{}
	if msgWhereis, err := git.NewCommand("annex", "whereis", "--json").RunInDir(localPath); err != nil {
		log.Error("[git annex whereis Error] err : %v", err)
	} else {
		if datasetToContentsMap, err = annex_ipfs.GetAnnexContentInfoListByDatasetNm(&msgWhereis, datasetNmList); err != nil {
			return nil, fmt.Errorf("[JSON Convert] err : %v ,fromPath : %v", err, localPath)
		}
	}

	datasetNmToFileMap = map[string]DatasetInfo{}
	for datasetNm, annexContentInfoList := range datasetToContentsMap {
		datasetInfo := DatasetInfo{}
		log.Trace("[Picking up annex content info] dataset name : %v", datasetNm)
		for _, content := range annexContentInfoList {
			inputPath := datasetNm + "/" + INPUT_FOLDER_NM //ex : datasetNm/input
			srcPath := datasetNm + "/" + SRC_FOLDER_NM
			OutputPath := datasetNm + "/" + OUTPUT_FOLDER_NM

			filePath := content.File                      // ex datasetNm/FolderNm/...../FileNm
			fullFilePath := repoBranchNm + "/" + filePath // ex RepoOwnerNm/RepoNm/BranchNm/datasetNm/FolderNm/...../FileNm
			if strings.HasPrefix(filePath, inputPath) {
				datasetInfo.InputList = append(datasetInfo.InputList, ContentInfo{fullFilePath, content.Hash})
			} else if strings.HasPrefix(filePath, srcPath) {
				datasetInfo.SrcList = append(datasetInfo.SrcList, ContentInfo{fullFilePath, content.Hash})
			} else if strings.HasPrefix(filePath, OutputPath) {
				datasetInfo.OutputList = append(datasetInfo.OutputList, ContentInfo{fullFilePath, content.Hash})
			}
		}
		datasetPath := repoBranchNm + "/" + datasetNm
		datasetNmToFileMap[datasetPath] = datasetInfo
		log.Trace("[datasetNmToFileMap] %v", datasetNmToFileMap)
	}
	return datasetNmToFileMap, nil
}

func CheckDatasetFormat(localPath string, datasetNm string) (err error) {
	log.Info("[Checking Dataset Formant] LocalPath : %v, Dataset Name : %v", localPath, datasetNm)

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return err
	}

	//データセット配下にinput, src, outputフォルダが存在するかをチェック
	if err = CheckFolder(localPath, datasetNm); err != nil {
		return err
	} //pass
	return nil
}

func CheckFolder(localPath string, datasetNm string) error {
	datasetPath := localPath + "/" + datasetNm
	inputPath := datasetPath + "/" + INPUT_FOLDER_NM
	srcPath := datasetPath + "/" + SRC_FOLDER_NM
	outputPath := datasetPath + "/" + OUTPUT_FOLDER_NM
	//Input
	if f, err := os.Stat(inputPath); os.IsNotExist(err) || !f.IsDir() {
		return fmt.Errorf("データセットに\"%v\" folder が存在しません", INPUT_FOLDER_NM)
	}

	//Src
	if f, err := os.Stat(srcPath); os.IsNotExist(err) || !f.IsDir() {
		return fmt.Errorf("データセットに\"%v\" folder が存在しません", SRC_FOLDER_NM)
	}

	//Output
	if f, err := os.Stat(outputPath); os.IsNotExist(err) || !f.IsDir() {
		return fmt.Errorf("データセットに\"%v\" folder が存在しません", OUTPUT_FOLDER_NM)
	}

	//input, src, outフォルダにファイルが存在するか確認する。
	//Input
	if is, emptyPath := CheckWithFileInFolder(inputPath); !is {
		return fmt.Errorf("%v配下にファイルが存在していません", emptyPath)
	}
	//Src
	if is, emptyPath := CheckWithFileInFolder(srcPath); !is {
		return fmt.Errorf("%v配下にファイルが存在していません", emptyPath)
	}

	//Output
	if is, emptyPath := CheckWithFileInFolder(outputPath); !is {
		return fmt.Errorf("%v配下にファイルが存在していません", emptyPath)
	}
	return nil
}

//全てのフォルダーに1つ以上のファイルが入っている場合、true
//1つでも空のフォルダーがあった場合、false
func CheckWithFileInFolder(folderPath string) (bool, string) {
	dataList, _ := filepath.Glob(folderPath + "/*")
	if dataList == nil {
		//フォルダー内が空
		return false, folderPath
	}
	for _, d := range dataList {
		if f, _ := os.Stat(d); f.IsDir() {
			if is, emptyPath := CheckWithFileInFolder(d); !is {
				return false, emptyPath
			}
		}
	}
	return true, ""
}

type UploadRepoOption struct {
	LastCommitID  string
	Branch        string
	TreePath      string
	UpperRopoPath string //RepoOwnerNm / RepoNm
}

//非公開データを公開データして、IPFSへのアップロードをし、コンテンツ情報を返す。
//
//@param　doer *User
//
//@param
//
//@param
//
//@param
//
//@param
func (repo *Repository) UpdateDataPrvToPub(opts UploadRepoOption) (map[string]AnnexUploadInfo, error) {
	//リモートレポジトリをクローンする。
	// repoWorkingPool.CheckIn(com.ToStr(repo.ID))
	// defer repoWorkingPool.CheckOut(com.ToStr(repo.ID))

	// if err := repo.DiscardLocalRepoBranchChanges(opts.Branch); err != nil {
	// 	return nil, fmt.Errorf("discard local repo branch[%s] changes: %v", opts.Branch, err)
	// } else if err := repo.UpdateLocalCopyBranch(opts.Branch); err != nil {
	// 	return nil, fmt.Errorf("update local copy branch[%s]: %v", opts.Branch, err)
	// }

	log.Trace("repo.LocalCopyPath()[%v]", repo.LocalCopyPath()) ///home/gogs/gogs/data/tmp/local-repo/71
	log.Trace("opts.UpperRopoPath : %v", opts.UpperRopoPath)

	//BCからコンテンツ情報を取得する。

	//非公開データ情報の選別

	//ハッシュ値比較

	//

	return nil, nil
}
