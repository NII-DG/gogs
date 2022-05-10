package db

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/NII-DG/gogs/internal/annex_ipfs"
	encyrptfile "github.com/NII-DG/gogs/internal/encyrpt_file"
	"github.com/NII-DG/gogs/internal/osutil"
	"github.com/gogs/git-module"
	"github.com/unknwon/com"
	log "unknwon.dev/clog/v2"
)

//TODO[2022-05-06]：プライベート or パブリック　レポジトリで異なる処理を行う
func (repo *Repository) UploadRepoFilesToIPFS(doer *User, opts UploadRepoFileOptions, isPrivate bool) (contentMap map[string]AnnexUploadInfo, err error) {
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
func (repo *Repository) PublicUploadRepoFiles(doer *User, opts UploadRepoFileOptions) (contentMap map[string]AnnexUploadInfo, err error) {
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

	annexSetup(localPath) // Initialise annex and set configuration (with add filter for filesizes)
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
func (repo *Repository) PrivateUploadRepoFiles(doer *User, opts UploadRepoFileOptions) (map[string]AnnexUploadInfo, error) {
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

	annexSetup(localPath) // Initialise annex and set configuration (with add filter for filesizes)

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
