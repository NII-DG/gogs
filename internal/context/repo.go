// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package context

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	"github.com/editorconfig/editorconfig-core-go/v2"
	"github.com/pkg/errors"
	"gopkg.in/macaron.v1"

	"github.com/gogs/git-module"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/db"
)

type AbstructCtxRepository interface {
	GetTreePath() string
	GetRepoLink() string
	GetBranchName() string
	GetCommit() *git.Commit
	GetCommitId() *git.SHA1
	GetLastCommitIdStr() string
	GetDbRepo() db.AbstructDbRepository
	GetGitRepo() *git.Repository
}

type AbstructGitRepository interface {
	BranchCommit(branch string, opts ...git.CatFileCommitOptions) (*git.Commit, error)
}

type AbstructCommit interface {
	Blob(subpath string, opts ...git.LsTreeOptions) (*git.Blob, error)
}

type PullRequest struct {
	BaseRepo *db.Repository
	Allowed  bool
	SameRepo bool
	HeadInfo string // [<user>:]<branch>
}

type Repository struct {
	AccessMode   db.AccessMode
	IsWatching   bool
	IsViewBranch bool
	IsViewTag    bool
	IsViewCommit bool
	Repository   *db.Repository
	Owner        *db.User
	Commit       *git.Commit
	Tag          *git.Tag
	GitRepo      *git.Repository
	BranchName   string
	TagName      string
	TreePath     string
	IsCanEditTreePath bool
	CommitID     string
	RepoLink     string
	CloneLink    db.CloneLink
	CommitsCount int64
	Mirror       *db.Mirror

	PullRequest *PullRequest
}

// GetTreePath is RCOS specific code.
// This returns value of "TreePath" field.
func (r *Repository) GetTreePath() string {
	return r.TreePath
}

// GetRepoLink is RCOS specific code.
// This returns value of "RepoLink" field.
func (r *Repository) GetRepoLink() string {
	return r.RepoLink
}

// GetBranchName is RCOS specific code.
// This returns value of "BranchName" field.
func (r *Repository) GetBranchName() string {
	return r.BranchName
}

// GetCommit is ROCS specific code.
// This returns value of "Commit" field.
func (r *Repository) GetCommit() *git.Commit {
	return r.Commit
}

// GetCommitId is RCOS specific code.
// This returns value of Commit.ID field in github.com/gogs/git-module.
func (r *Repository) GetCommitId() *git.SHA1 {
	return r.Commit.ID
}

// GetLastCommitIdStr is RCOS specific code.
// This returns value of "CommitID" field.
func (r *Repository) GetLastCommitIdStr() string {
	return r.CommitID
}

// GetDbRepo is RCOS specific code.
// This returns value of "Repository" field.
func (r *Repository) GetDbRepo() db.AbstructDbRepository {
	return r.Repository
}

// GetGitRepo is RCOS specific code.
// This returns value of "GitRepo" field.
func (r *Repository) GetGitRepo() *git.Repository {
	return r.GitRepo
}

// IsOwner returns true if current user is the owner of repository.
func (r *Repository) IsOwner() bool {
	return r.AccessMode >= db.AccessModeOwner
}

// IsAdmin returns true if current user has admin or higher access of repository.
func (r *Repository) IsAdmin() bool {
	return r.AccessMode >= db.AccessModeAdmin
}

// IsWriter returns true if current user has write or higher access of repository.
func (r *Repository) IsWriter() bool {
	return r.AccessMode >= db.AccessModeWrite
}

// HasAccess returns true if the current user has at least read access for this repository
func (r *Repository) HasAccess() bool {
	return r.AccessMode >= db.AccessModeRead
}

// CanEnableEditor returns true if repository is editable and user has proper access level.
func (r *Repository) CanEnableEditor() bool {
	// RCOS specific code.
	// Added determination of updatable Treepaths
	return r.Repository.CanEnableEditor() && r.IsViewBranch && r.IsWriter() && !r.Repository.IsBranchRequirePullRequest(r.BranchName) && r.IsCanEditTreePath
}

// Editorconfig returns the ".editorconfig" definition if found in the HEAD of the default branch.
func (r *Repository) Editorconfig() (*editorconfig.Editorconfig, error) {
	commit, err := r.GitRepo.BranchCommit(r.Repository.DefaultBranch)
	if err != nil {
		return nil, errors.Wrapf(err, "get commit of branch %q ", r.Repository.DefaultBranch)
	}

	entry, err := commit.TreeEntry(".editorconfig")
	if err != nil {
		return nil, errors.Wrap(err, "get .editorconfig")
	}

	p, err := entry.Blob().Bytes()
	if err != nil {
		return nil, errors.Wrap(err, "read .editorconfig")
	}
	return editorconfig.Parse(bytes.NewReader(p))
}

// MakeURL accepts a string or url.URL as argument and returns escaped URL prepended with repository URL.
func (r *Repository) MakeURL(location interface{}) string {
	switch location := location.(type) {
	case string:
		tempURL := url.URL{
			Path: r.RepoLink + "/" + location,
		}
		return tempURL.String()
	case url.URL:
		location.Path = r.RepoLink + "/" + location.Path
		return location.String()
	default:
		panic("location type must be either string or url.URL")
	}
}

// PullRequestURL returns URL for composing a pull request.
// This function does not check if the repository can actually compose a pull request.
func (r *Repository) PullRequestURL(baseBranch, headBranch string) string {
	repoLink := r.RepoLink
	if r.PullRequest.BaseRepo != nil {
		repoLink = r.PullRequest.BaseRepo.Link()
	}
	return fmt.Sprintf("%s/compare/%s...%s:%s", repoLink, baseBranch, r.Owner.Name, headBranch)
}

// [0]: issues, [1]: wiki
func RepoAssignment(pages ...bool) macaron.Handler {
	return func(c *Context) {
		var (
			owner        *db.User
			err          error
			isIssuesPage bool
			isWikiPage   bool
		)

		if len(pages) > 0 {
			isIssuesPage = pages[0]
		}
		if len(pages) > 1 {
			isWikiPage = pages[1]
		}

		ownerName := c.Params(":username")
		repoName := strings.TrimSuffix(c.Params(":reponame"), ".git")

		// Check if the user is the same as the repository owner
		if c.IsLogged && c.User.LowerName == strings.ToLower(ownerName) {
			owner = c.User
		} else {
			owner, err = db.GetUserByName(ownerName)
			if err != nil {
				c.NotFoundOrError(err, "get user by name")
				return
			}
		}
		c.Repo.Owner = owner
		c.Data["Username"] = c.Repo.Owner.Name

		repo, err := db.GetRepositoryByName(owner.ID, repoName)
		if err != nil {
			c.NotFoundOrError(err, "get repository by name")
			return
		}

		c.Repo.Repository = repo
		c.Data["RepoName"] = c.Repo.Repository.Name
		c.Data["IsBareRepo"] = c.Repo.Repository.IsBare
		c.Repo.RepoLink = repo.Link()
		c.Data["RepoLink"] = c.Repo.RepoLink
		c.Data["RepoRelPath"] = c.Repo.Owner.Name + "/" + c.Repo.Repository.Name

		// Admin has super access
		if c.IsLogged && c.User.IsAdmin {
			c.Repo.AccessMode = db.AccessModeOwner
		} else {
			mode, err := db.UserAccessMode(c.UserID(), c.Repo.Repository)
			if err != nil {
				c.Error(err, "get user access mode")
				return
			}
			c.Repo.AccessMode = mode
		}

		// If the authenticated user has no direct access, see if the repository is a fork
		// and whether the user has access to the base repository.
		if c.Repo.AccessMode == db.AccessModeNone && c.Repo.Repository.IsFork {
			mode, err := db.UserAccessMode(c.UserID(), c.Repo.Repository.BaseRepo)
			if err != nil {
				c.Error(err, "get user access mode of base repository")
				return
			}
			// Users shouldn't have indirect access level higher than write.
			if mode > db.AccessModeWrite {
				mode = db.AccessModeWrite
			}
			c.Repo.AccessMode = mode
		}

		// Check access
		if c.Repo.AccessMode == db.AccessModeNone {
			// Redirect to any accessible page if not yet on it
			if repo.IsPartialPublic() &&
				(!(isIssuesPage || isWikiPage) ||
					(isIssuesPage && !repo.CanGuestViewIssues()) ||
					(isWikiPage && !repo.CanGuestViewWiki())) {
				switch {
				case repo.CanGuestViewIssues():
					c.Redirect(repo.Link() + "/issues")
				case repo.CanGuestViewWiki():
					c.Redirect(repo.Link() + "/wiki")
				default:
					c.NotFound()
				}
				return
			}

			// Response 404 if user is on completely private repository or possible accessible page but owner doesn't enabled
			if !repo.IsPartialPublic() ||
				(isIssuesPage && !repo.CanGuestViewIssues()) ||
				(isWikiPage && !repo.CanGuestViewWiki()) {
				c.NotFound()
				return
			}

			c.Repo.Repository.EnableIssues = repo.CanGuestViewIssues()
			c.Repo.Repository.EnableWiki = repo.CanGuestViewWiki()
		}

		if repo.IsMirror {
			c.Repo.Mirror, err = db.GetMirrorByRepoID(repo.ID)
			if err != nil {
				c.Error(err, "get mirror by repository ID")
				return
			}
			c.Data["MirrorEnablePrune"] = c.Repo.Mirror.EnablePrune
			c.Data["MirrorInterval"] = c.Repo.Mirror.Interval
			c.Data["Mirror"] = c.Repo.Mirror
		}

		gitRepo, err := git.Open(db.RepoPath(ownerName, repoName))
		if err != nil {
			c.Error(err, "open repository")
			return
		}
		c.Repo.GitRepo = gitRepo

		tags, err := c.Repo.GitRepo.Tags()
		if err != nil {
			c.Error(err, "get tags")
			return
		}
		c.Data["Tags"] = tags
		c.Repo.Repository.NumTags = len(tags)

		c.Data["Title"] = owner.Name + "/" + repo.Name
		c.Data["Repository"] = repo
		c.Data["Owner"] = c.Repo.Repository.Owner
		c.Data["IsRepositoryOwner"] = c.Repo.IsOwner()
		c.Data["IsRepositoryAdmin"] = c.Repo.IsAdmin()
		c.Data["IsRepositoryWriter"] = c.Repo.IsWriter()

		c.Data["DisableSSH"] = conf.SSH.Disabled
		c.Data["DisableHTTP"] = conf.Repository.DisableHTTPGit
		c.Data["ShowHTTP"] = conf.Repository.ShowHTTPGit
		c.Data["CloneLink"] = repo.CloneLink()
		c.Data["WikiCloneLink"] = repo.WikiCloneLink()

		u, _ := url.Parse(conf.Server.ExternalURL)
		ginURL := fmt.Sprintf("%s://%s", u.Scheme, u.Host)
		c.Data["ginURL"] = url.QueryEscape(ginURL)
		c.Data["Scheme"] = u.Scheme
		c.Data["Host"] = u.Host

		if c.IsLogged {
			c.Data["IsWatchingRepo"] = db.IsWatching(c.User.ID, repo.ID)
			c.Data["IsStaringRepo"] = db.IsStaring(c.User.ID, repo.ID)
			c.Data["HasForked"] = c.User.HasForkedRepo(c.Repo.Repository.ID)
		}

		// repo is bare and display enable
		if c.Repo.Repository.IsBare {
			return
		}

		c.Data["TagName"] = c.Repo.TagName
		branches, err := c.Repo.GitRepo.Branches()
		if err != nil {
			c.Error(err, "get branches")
			return
		}
		/**
		As of 2023/06/01, Gin-fork supports only master branch. This has led to the hiding of the branch pulldown feature in the repository top UI. The following code is a modification for this purpose. The commented out code will be left as it is expected to be recovered in future development when branching is supported.

		START
		*/
		// checking branche list has 'master' branch
		hasMasterBranch := false
		for _, baranchName := range branches {
			if baranchName == "master" {
				hasMasterBranch = true
			}
		}
		if !hasMasterBranch {
			// If the master branch cannot be retrieved, return a 404 screen
			c.NotFoundWithErrMsg("Cannot find master branch. Please wait for 1~2 minutes and try accessing again.")
			return
		}
		// c.Data["Branches"] = branches
		// c.Data["BrancheCount"] = len(branches)
		/**
		END
		*/

		// If not branch selected, try default one.
		// If default branch doesn't exists, fall back to some other branch.
		if len(c.Repo.BranchName) == 0 {
			if len(c.Repo.Repository.DefaultBranch) > 0 && gitRepo.HasBranch(c.Repo.Repository.DefaultBranch) {
				c.Repo.BranchName = c.Repo.Repository.DefaultBranch
			} else if len(branches) > 0 {
				c.Repo.BranchName = branches[0]
			}
		}
		c.Data["BranchName"] = c.Repo.BranchName
		c.Data["CommitID"] = c.Repo.CommitID

		c.Data["IsGuest"] = !c.Repo.HasAccess()

		hasDmp := HasFileInRepo(c, "/dmp.json")
		c.Data["HasDmpJson"] = hasDmp
		if hasDmp {
			c.Data["IsDOIReady"] = isDOIReady(c)
		}

		c.Data["HasMaDmp"] = HasFileInRepo(c, "/maDMP.ipynb")
		c.Data["HasExperiments"] = HasTreeInRepo(c, "/experiments")

		// if doi := getRepoDOI(c); doi != "" && libgin.IsRegisteredDOI(doi) {
		// 	c.Data["DOI"] = doi
		// }

		if doi := getRepoDOI(c); doi != "" {
			c.Data["DOI"] = doi
		}

	}
}

// RepoRef handles repository reference name including those contain `/`.
func RepoRef() macaron.Handler {
	return func(c *Context) {
		// Empty repository does not have reference information.
		if c.Repo.Repository.IsBare {
			return
		}

		var (
			refName string
			err     error
		)

		// For API calls.
		if c.Repo.GitRepo == nil {
			repoPath := db.RepoPath(c.Repo.Owner.Name, c.Repo.Repository.Name)
			c.Repo.GitRepo, err = git.Open(repoPath)
			if err != nil {
				c.Error(err, "open repository")
				return
			}
		}

		// Get default branch.
		if len(c.Params("*")) == 0 {
			refName = c.Repo.Repository.DefaultBranch
			if !c.Repo.GitRepo.HasBranch(refName) {
				branches, err := c.Repo.GitRepo.Branches()
				if err != nil {
					c.Error(err, "get branches")
					return
				}
				refName = branches[0]
			}
			c.Repo.Commit, err = c.Repo.GitRepo.BranchCommit(refName)
			if err != nil {
				c.Error(err, "get branch commit")
				return
			}
			c.Repo.CommitID = c.Repo.Commit.ID.String()
			c.Repo.IsViewBranch = true

		} else {
			hasMatched := false
			parts := strings.Split(c.Params("*"), "/")
			for i, part := range parts {
				refName = strings.TrimPrefix(refName+"/"+part, "/")

				if c.Repo.GitRepo.HasBranch(refName) ||
					c.Repo.GitRepo.HasTag(refName) {
					if i < len(parts)-1 {
						c.Repo.TreePath = strings.Join(parts[i+1:], "/")
					}
					hasMatched = true
					break
				}
			}
			if !hasMatched && len(parts[0]) == 40 {
				refName = parts[0]
				c.Repo.TreePath = strings.Join(parts[1:], "/")
			}

			if c.Repo.GitRepo.HasBranch(refName) {
				c.Repo.IsViewBranch = true

				c.Repo.Commit, err = c.Repo.GitRepo.BranchCommit(refName)
				if err != nil {
					c.Error(err, "get branch commit")
					return
				}
				c.Repo.CommitID = c.Repo.Commit.ID.String()

			} else if c.Repo.GitRepo.HasTag(refName) {
				c.Repo.IsViewTag = true
				c.Repo.Commit, err = c.Repo.GitRepo.TagCommit(refName)
				if err != nil {
					c.Error(err, "get tag commit")
					return
				}
				c.Repo.CommitID = c.Repo.Commit.ID.String()
			} else if len(refName) == 40 {
				c.Repo.IsViewCommit = true
				c.Repo.CommitID = refName

				c.Repo.Commit, err = c.Repo.GitRepo.CatFileCommit(refName)
				if err != nil {
					c.NotFound()
					return
				}
			} else {
				c.NotFound()
				return
			}
		}

		// RCOS specific code
		// Under the experiments folder and under the input_data folder or 
		// Under the experiments folder and output_data folder
		// Forbid add/upd/del file
		if strings.Contains( c.Repo.TreePath, "experiments/") && 
		   ( strings.Contains( c.Repo.TreePath, "input_data/") || strings.Contains( c.Repo.TreePath, "output_data/") ||
		     strings.HasSuffix( c.Repo.TreePath, "input_data") || strings.HasSuffix( c.Repo.TreePath, "output_data") ){
			c.Repo.IsCanEditTreePath = false
		}else{
			c.Repo.IsCanEditTreePath = true
		}

		c.Repo.BranchName = refName
		c.Data["BranchName"] = c.Repo.BranchName
		c.Data["CommitID"] = c.Repo.CommitID
		c.Data["TreePath"] = c.Repo.TreePath
		c.Data["IsViewBranch"] = c.Repo.IsViewBranch
		c.Data["IsViewTag"] = c.Repo.IsViewTag
		c.Data["IsViewCommit"] = c.Repo.IsViewCommit
		c.Data["IsCanEditTreePath"] = c.Repo.IsCanEditTreePath

		// People who have push access or have fored repository can propose a new pull request.
		if c.Repo.IsWriter() || (c.IsLogged && c.User.HasForkedRepo(c.Repo.Repository.ID)) {
			// Pull request is allowed if this is a fork repository
			// and base repository accepts pull requests.
			if c.Repo.Repository.BaseRepo != nil {
				if c.Repo.Repository.BaseRepo.AllowsPulls() {
					c.Repo.PullRequest.Allowed = true
					// In-repository pull requests has higher priority than cross-repository if user is viewing
					// base repository and 1) has write access to it 2) has forked it.
					if c.Repo.IsWriter() {
						c.Data["BaseRepo"] = c.Repo.Repository.BaseRepo
						c.Repo.PullRequest.BaseRepo = c.Repo.Repository.BaseRepo
						c.Repo.PullRequest.HeadInfo = c.Repo.Owner.Name + ":" + c.Repo.BranchName
					} else {
						c.Data["BaseRepo"] = c.Repo.Repository
						c.Repo.PullRequest.BaseRepo = c.Repo.Repository
						c.Repo.PullRequest.HeadInfo = c.User.Name + ":" + c.Repo.BranchName
					}
				}
			} else {
				// Or, this is repository accepts pull requests between branches.
				if c.Repo.Repository.AllowsPulls() {
					c.Data["BaseRepo"] = c.Repo.Repository
					c.Repo.PullRequest.BaseRepo = c.Repo.Repository
					c.Repo.PullRequest.Allowed = true
					c.Repo.PullRequest.SameRepo = true
					c.Repo.PullRequest.HeadInfo = c.Repo.BranchName
				}
			}
		}
		c.Data["PullRequestCtx"] = c.Repo.PullRequest
	}
}

func RequireRepoAdmin() macaron.Handler {
	return func(c *Context) {
		if !c.IsLogged || (!c.Repo.IsAdmin() && !c.User.IsAdmin) {
			c.NotFound()
			return
		}
	}
}

func RequireRepoWriter() macaron.Handler {
	return func(c *Context) {
		if !c.IsLogged || (!c.Repo.IsWriter() && !c.User.IsAdmin) {
			c.NotFound()
			return
		}
	}
}

// GitHookService checks if repository Git hooks service has been enabled.
func GitHookService() macaron.Handler {
	return func(c *Context) {
		if !c.User.CanEditGitHook() {
			c.NotFound()
			return
		}
	}
}
