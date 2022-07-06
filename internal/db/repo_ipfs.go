package db

import (
	"fmt"
	_ "image/jpeg"
	"strings"

	"github.com/NII-DG/gogs/internal/process"
	"github.com/NII-DG/gogs/internal/strutil"
)

// internal/db/repo.go CreateRepository()改変
func AltCreateRepository(doer, owner *User, opts CreateRepoOptions) (_ *Repository, err error) {
	if !owner.CanCreateRepo() {
		return nil, ErrReachLimitOfRepo{Limit: owner.RepoCreationNum()}
	}
	//共通キー生成
	password, err := strutil.RandomChars(32)
	if err != nil {
		return nil, fmt.Errorf("[Cannot Generating Repository Password]]: %v", err)
	}

	repo := &Repository{
		OwnerID:      owner.ID,
		Owner:        owner,
		Name:         opts.Name,
		LowerName:    strings.ToLower(opts.Name),
		Description:  opts.Description,
		IsPrivate:    opts.IsPrivate,
		IsUnlisted:   opts.IsUnlisted,
		EnableWiki:   true,
		EnableIssues: true,
		EnablePulls:  true,
		Password:     password,
	}

	sess := x.NewSession()
	defer sess.Close()
	if err = sess.Begin(); err != nil {
		return nil, err
	}

	if err = createRepository(sess, doer, owner, repo); err != nil {
		return nil, err
	}

	// No need for init mirror.
	if !opts.IsMirror {
		repoPath := RepoPath(owner.Name, repo.Name)
		if err = initRepository(sess, repoPath, doer, repo, opts); err != nil {
			RemoveAllWithNotice("Delete repository for initialization failure", repoPath)
			return nil, fmt.Errorf("initRepository: %v", err)
		}

		_, stderr, err := process.ExecDir(-1,
			repoPath, fmt.Sprintf("CreateRepository 'git update-server-info': %s", repoPath),
			"git", "update-server-info")
		if err != nil {
			return nil, fmt.Errorf("CreateRepository 'git update-server-info': %s", stderr)
		}
	}

	return repo, sess.Commit()
}
