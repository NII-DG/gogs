// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/pkg/errors"
	"github.com/unknwon/com"
	log "unknwon.dev/clog/v2"

	"github.com/gogs/git-module"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/form"
	"github.com/NII-DG/gogs/internal/tool"
	"github.com/NII-DG/gogs/internal/utils/const_utils"
	gen "github.com/NII-DG/gogs/internal/utils/generator"
)

const (
	CREATE  = "repo/create"
	MIGRATE = "repo/migrate"
	LAUNCH  = "repo/launch"
)

func MustBeNotBare(c *context.Context) {
	if c.Repo.Repository.IsBare {
		c.NotFound()
	}
}

func checkContextUser(c *context.Context, uid int64) *db.User {
	orgs, err := db.GetOwnedOrgsByUserIDDesc(c.User.ID, "updated_unix")
	if err != nil {
		c.Error(err, "get owned organization by user ID")
		return nil
	}
	c.Data["Orgs"] = orgs

	// Not equal means current user is an organization.
	if uid == c.User.ID || uid == 0 {
		return c.User
	}

	org, err := db.GetUserByID(uid)
	if db.IsErrUserNotExist(err) {
		return c.User
	}

	if err != nil {
		c.Error(err, "get user by ID")
		return nil
	}

	// Check ownership of organization.
	if !org.IsOrganization() || !(c.User.IsAdmin || org.IsOwnedBy(c.User.ID)) {
		c.Status(http.StatusForbidden)
		return nil
	}
	return org
}

func Create(c *context.Context) {
	c.Title("new_repo")

	// Give default value for template to render.
	c.Data["Gitignores"] = db.Gitignores
	c.Data["Licenses"] = db.Licenses
	c.Data["Readmes"] = db.Readmes
	c.Data["readme"] = "Default"
	c.Data["private"] = c.User.LastRepoVisibility
	c.Data["IsForcedPrivate"] = conf.Repository.ForcePrivate
	c.Data["auto_init"] = conf.Repository.AutoInit

	ctxUser := checkContextUser(c, c.QueryInt64("org"))
	if c.Written() {
		return
	}
	c.Data["ContextUser"] = ctxUser

	c.Success(CREATE)
}

func handleCreateError(c *context.Context, owner *db.User, err error, name, tpl string, form interface{}) {
	switch {
	case db.IsErrReachLimitOfRepo(err):
		c.RenderWithErr(c.Tr("repo.form.reach_limit_of_creation", owner.RepoCreationNum()), tpl, form)
	case db.IsErrRepoAlreadyExist(err):
		c.Data["Err_RepoName"] = true
		c.RenderWithErr(c.Tr("form.repo_name_been_taken"), tpl, form)
	case db.IsErrNameNotAllowed(err):
		c.Data["Err_RepoName"] = true
		c.RenderWithErr(c.Tr("repo.form.name_not_allowed", err.(db.ErrNameNotAllowed).Value()), tpl, form)
	default:
		c.Error(err, name)
	}
}

func CreatePost(c *context.Context, f form.CreateRepo) {
	c.Data["Title"] = c.Tr("new_repo")

	c.Data["Gitignores"] = db.Gitignores
	c.Data["Licenses"] = db.Licenses
	c.Data["Readmes"] = db.Readmes

	ctxUser := checkContextUser(c, f.UserID)
	if c.Written() {
		return
	}
	c.Data["ContextUser"] = ctxUser

	if c.HasError() {
		c.Success(CREATE)
		return
	}

	// velidate Research Project Name
	projectname_has_char := false
	for _, char := range f.ProjectName {
		if unicode.IsLetter(char) || unicode.Is(unicode.Hiragana, char) || unicode.Is(unicode.Katakana, char) || unicode.Is(unicode.Han, char) {
			projectname_has_char = true
		}

	}
	if projectname_has_char == false {
		c.RenderWithErr(c.Tr("form.projectname_has_no_char"), CREATE, &f)
		return
	}

	repo, err := db.CreateRepository(c.User, ctxUser, db.CreateRepoOptions{
		Name:               f.RepoName,
		Description:        strings.ReplaceAll(f.Description, "\r\n", "\n"),
		Gitignores:         "",
		License:            "",
		Readme:             "Default",
		IsPrivate:          f.Private || conf.Repository.ForcePrivate,
		IsUnlisted:         f.Unlisted,
		AutoInit:           true,
		ProjectName:        strings.ReplaceAll(f.ProjectName, "\r\n", "\n"),
		ProjectDescription: strings.ReplaceAll(f.ProjectDescription, "\r\n", "\n"),
	})
	if err == nil {
		log.Trace("Repository created [%d]: %s/%s", repo.ID, ctxUser.Name, repo.Name)
		c.Redirect(conf.Server.Subpath + "/" + ctxUser.Name + "/" + repo.Name)
		return
	}

	if repo != nil {
		if errDelete := db.DeleteRepository(ctxUser.ID, repo.ID); errDelete != nil {
			log.Error("DeleteRepository: %v", errDelete)
		}
	}

	handleCreateError(c, ctxUser, err, "CreatePost", CREATE, &f)
}

func Migrate(c *context.Context) {
	c.Data["Title"] = c.Tr("new_migrate")
	c.Data["private"] = c.User.LastRepoVisibility
	c.Data["IsForcedPrivate"] = conf.Repository.ForcePrivate
	c.Data["mirror"] = c.Query("mirror") == "1"

	ctxUser := checkContextUser(c, c.QueryInt64("org"))
	if c.Written() {
		return
	}
	c.Data["ContextUser"] = ctxUser

	c.Success(MIGRATE)
}

func MigratePost(c *context.Context, f form.MigrateRepo) {
	c.Data["Title"] = c.Tr("new_migrate")

	ctxUser := checkContextUser(c, f.Uid)
	if c.Written() {
		return
	}
	c.Data["ContextUser"] = ctxUser

	if c.HasError() {
		c.Success(MIGRATE)
		return
	}

	remoteAddr, err := f.ParseRemoteAddr(c.User)
	if err != nil {
		if db.IsErrInvalidCloneAddr(err) {
			c.Data["Err_CloneAddr"] = true
			addrErr := err.(db.ErrInvalidCloneAddr)
			switch {
			case addrErr.IsURLError:
				c.RenderWithErr(c.Tr("form.url_error"), MIGRATE, &f)
			case addrErr.IsPermissionDenied:
				c.RenderWithErr(c.Tr("repo.migrate.permission_denied"), MIGRATE, &f)
			case addrErr.IsInvalidPath:
				c.RenderWithErr(c.Tr("repo.migrate.invalid_local_path"), MIGRATE, &f)
			default:
				c.Error(err, "unexpected error")
			}
		} else {
			c.Error(err, "parse remote address")
		}
		return
	}

	repo, err := db.MigrateRepository(c.User, ctxUser, db.MigrateRepoOptions{
		Name:        f.RepoName,
		Description: f.Description,
		IsPrivate:   f.Private || conf.Repository.ForcePrivate,
		IsUnlisted:  f.Unlisted,
		IsMirror:    f.Mirror,
		RemoteAddr:  remoteAddr,
	})
	if err == nil {
		log.Trace("Repository migrated [%d]: %s/%s", repo.ID, ctxUser.Name, f.RepoName)
		c.Redirect(conf.Server.Subpath + "/" + ctxUser.Name + "/" + f.RepoName)
		return
	}

	if repo != nil {
		if errDelete := db.DeleteRepository(ctxUser.ID, repo.ID); errDelete != nil {
			log.Error("DeleteRepository: %v", errDelete)
		}
	}

	if strings.Contains(err.Error(), "Authentication failed") ||
		strings.Contains(err.Error(), "could not read Username") {
		c.Data["Err_Auth"] = true
		c.RenderWithErr(c.Tr("form.auth_failed", db.HandleMirrorCredentials(err.Error(), true)), MIGRATE, &f)
		return
	} else if strings.Contains(err.Error(), "fatal:") {
		c.Data["Err_CloneAddr"] = true
		c.RenderWithErr(c.Tr("repo.migrate.failed", db.HandleMirrorCredentials(err.Error(), true)), MIGRATE, &f)
		return
	}

	handleCreateError(c, ctxUser, err, "MigratePost", MIGRATE, &f)
}

func Action(c *context.Context) {
	var err error
	switch c.Params(":action") {
	case "watch":
		err = db.WatchRepo(c.User.ID, c.Repo.Repository.ID, true)
	case "unwatch":
		if userID := c.QueryInt64("user_id"); userID != 0 {
			if c.User.IsAdmin {
				err = db.WatchRepo(userID, c.Repo.Repository.ID, false)
			}
		} else {
			err = db.WatchRepo(c.User.ID, c.Repo.Repository.ID, false)
		}
	case "star":
		err = db.StarRepo(c.User.ID, c.Repo.Repository.ID, true)
	case "unstar":
		err = db.StarRepo(c.User.ID, c.Repo.Repository.ID, false)
	case "desc": // FIXME: this is not used
		if !c.Repo.IsOwner() {
			c.NotFound()
			return
		}

		c.Repo.Repository.Description = c.Query("desc")
		c.Repo.Repository.Website = c.Query("site")
		err = db.UpdateRepository(c.Repo.Repository, false)
	}

	if err != nil {
		c.Errorf(err, "action %q", c.Params(":action"))
		return
	}

	redirectTo := c.Query("redirect_to")
	if !tool.IsSameSiteURLPath(redirectTo) {
		redirectTo = c.Repo.RepoLink
	}
	c.Redirect(redirectTo)
}

func Download(c *context.Context) {

	var (
		uri           = c.Params("*")
		refName       string
		ext           string
		archivePath   string
		archiveFormat git.ArchiveFormat
	)

	switch {
	case strings.HasSuffix(uri, ".zip"):
		ext = ".zip"
		archivePath = filepath.Join(c.Repo.GitRepo.Path(), "archives", "zip")
		archiveFormat = git.ArchiveZip
	case strings.HasSuffix(uri, ".tar.gz"):
		ext = ".tar.gz"
		archivePath = filepath.Join(c.Repo.GitRepo.Path(), "archives", "targz")
		archiveFormat = git.ArchiveTarGz
	default:
		log.Trace("Unknown format: %s", uri)
		c.NotFound()
		return
	}
	refName = strings.TrimSuffix(uri, ext)

	if !com.IsDir(archivePath) {
		if err := os.MkdirAll(archivePath, os.ModePerm); err != nil {
			c.Error(err, "create archive directory")
			return
		}
	}

	log.Info("Downloaded: %v", c.Repo.GitRepo.Path())

	// Get corresponding commit.
	var (
		commit *git.Commit
		err    error
	)
	gitRepo := c.Repo.GitRepo
	if gitRepo.HasBranch(refName) {
		commit, err = gitRepo.BranchCommit(refName)
		if err != nil {
			c.Error(err, "get branch commit")
			return
		}
	} else if gitRepo.HasTag(refName) {
		commit, err = gitRepo.TagCommit(refName)
		if err != nil {
			c.Error(err, "get tag commit")
			return
		}
	} else if len(refName) >= 7 && len(refName) <= 40 {
		commit, err = gitRepo.CatFileCommit(refName)
		if err != nil {
			c.NotFound()
			return
		}
	} else {
		c.NotFound()
		return
	}

	archivePath = path.Join(archivePath, tool.ShortSHA1(commit.ID.String())+ext)
	if !com.IsFile(archivePath) {
		if err := commit.CreateArchive(archiveFormat, archivePath); err != nil {
			c.Error(err, "creates archive")
			return
		}
	}

	c.ServeFile(archivePath, c.Repo.Repository.Name+"-"+refName+ext)
}

func LaunchMadmp(c *context.Context) {
	if c.Repo.Repository.IsPrivate {
		c.Title("launch binder")
		c.Data["Dest"] = "madmp"
		c.Success(LAUNCH)
	} else {
		url := fmt.Sprintf("%s://%s/%s/%s", c.Data["Scheme"], c.Data["Host"], c.Repo.Owner.Name, c.Repo.Repository.Name)
		url = strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F", "/", "%2F").Replace(url)
		c.RawRedirect("https://binder.cs.rcos.nii.ac.jp/v2/git/" + url + ".git/master?filepath=maDMP.ipynb")
	}
}

func LaunchResearch(c *context.Context) {
	oldPath := "WORKFLOWS/base_FLOW.ipynb"
	newPath := "WORKFLOWS/notebooks/research/base_FLOW.ipynb"
	if !context.HasFileInRepo(c, oldPath) && !context.HasFileInRepo(c, newPath) {
		c.Flash.Error(c.Tr("rcos.broken_flow"))
		c.Redirect(c.GetRepo().GetRepoLink() + "/container")
		return
	}

	if c.Repo.Repository.IsPrivate {
		c.Title("launch binder")
		c.Data["Dest"] = "research"
		c.Success(LAUNCH)
	} else {
		repoName := fmt.Sprintf("%s://%s/%s/%s.git", c.Data["Scheme"], c.Data["Host"], c.Repo.Owner.Name, c.Repo.Repository.Name)
		repoName = strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F", "/", "%2F").Replace(repoName)
		binderUrl := "https://binder.cs.rcos.nii.ac.jp/v2/git"
		branch := "master"
		filepath := newPath
		if !context.HasFileInRepo(c, filepath) {
			filepath = oldPath
		}
		redirectURL := fmt.Sprintf("%s/%s/%s?filepath=%s", binderUrl, repoName, branch, filepath)
		c.RawRedirect(redirectURL)
	}
}

func LaunchExperiment(c *context.Context) {
	oldPath := "WORKFLOWS/EX-WORKFLOWS/util/required_rebuild_container.ipynb"
	newPath := "WORKFLOWS/notebooks/experiment/required_rebuild_container.ipynb"
	if !context.HasFileInRepo(c, oldPath) && !context.HasFileInRepo(c, newPath) {
		c.Flash.Error(c.Tr("rcos.broken_flow"))
		c.Redirect(c.GetRepo().GetRepoLink() + "/container")
		return
	}

	if c.Repo.Repository.IsPrivate {
		c.Title("launch binder")
		c.Data["Dest"] = "experiment"
		c.Success(LAUNCH)
	} else {
		repoName := fmt.Sprintf("%s://%s/%s/%s.git", c.Data["Scheme"], c.Data["Host"], c.Repo.Owner.Name, c.Repo.Repository.Name)
		repoName = strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F", "/", "%2F").Replace(repoName)
		binderUrl := "https://binder.cs.rcos.nii.ac.jp/v2/git"
		branch := "master"
		filepath := newPath
		if !context.HasFileInRepo(c, filepath) {
			filepath = oldPath
		}
		redirectURL := fmt.Sprintf("%s/%s/%s?filepath=%s", binderUrl, repoName, branch, filepath)
		c.RawRedirect(redirectURL)
	}
}

// when a repository is private, show password form
func LaunchPost(c *context.Context, f form.Pass) {
	c.Title("launch binder")
	dest := c.Params(":dest")

	if c.HasError() {
		c.RenderWithErr(c.Tr("rcos.private_pass_warn"), LAUNCH, &f)
		return
	}
	loginSources, err := db.LoginSources.List(db.ListLoginSourceOpts{OnlyActivated: true})
	if err != nil {
		c.Error(err, "list activated login sources")
		return
	}
	c.Data["LoginSources"] = loginSources

	if c.HasError() {
		c.RenderWithErr(c.Tr("form.username_password_incorrect"), HOME, &f)
		return
	}
	// User Auth
	_, err = db.Users.Authenticate(c.User.Name, f.Password, 0)
	if err != nil {
		switch errors.Cause(err).(type) {
		case db.ErrUserNotExist:
			c.FormErr("Password")
			c.RenderWithErr(c.Tr("form.enterred_invalid_password"), LAUNCH, &f)

		case db.ErrLoginSourceMismatch:
			c.FormErr("LoginSource")
			c.Redirect(c.GetRepo().GetRepoLink() + "/launch/" + dest)
		default:
			c.Error(err, "authenticate user")
		}
		return
	}

	// create new token for building jupyter container
	randStr, err := gen.MakeRandomStrByAlphabetDigit(7)
	if err != nil {
		c.Error(err, "Failed to generate random string")
	}
	token_name := fmt.Sprintf("%s-%s", const_utils.Get_BUILD_TOKEN(), randStr)
	build_token, err := db.AccessTokens.Create(c.User.ID, token_name, conf.DG.BuildAccessTokenExpireMinutes)

	if err != nil {
		if db.IsErrAccessTokenAlreadyExist(err) {
			c.RenderWithErr(c.Tr("rocs.fial_preparing_for_build"), LAUNCH, &f)
			return
		} else {
			c.Error(err, "new access token")
			return
		}
	}

	// create redirect URL for building jupyter container using BinderHub powered by RCOS
	repoName := fmt.Sprintf("%s://%s:%s@%s/%s/%s.git", c.Data["Scheme"], c.User.Name, build_token.Sha1, c.Data["Host"], c.Repo.Owner.Name, c.Repo.Repository.Name)
	repoName = strings.NewReplacer("%", "%25", "#", "%23", " ", "%20", "?", "%3F", "/", "%2F").Replace(repoName)

	binderUrl := "https://binder.cs.rcos.nii.ac.jp/v2/git"
	branch := "master"
	if dest == "research" {
		filepath := "WORKFLOWS/notebooks/research/base_FLOW.ipynb"
		if !context.HasFileInRepo(c, filepath) {
			filepath = "WORKFLOWS/base_FLOW.ipynb"
		}
		redirectURL := fmt.Sprintf("%s/%s/%s?filepath=%s", binderUrl, repoName, branch, filepath)
		c.RawRedirect(redirectURL)
	} else if dest == "experiment" {
		filepath := "WORKFLOWS/notebooks/experiment/required_rebuild_container.ipynb"
		if !context.HasFileInRepo(c, filepath) {
			filepath = "WORKFLOWS/EX-WORKFLOWS/util/required_rebuild_container.ipynb"
		}
		redirectURL := fmt.Sprintf("%s/%s/%s?filepath=%s", binderUrl, repoName, branch, filepath)
		c.RawRedirect(redirectURL)
	} else {
		filepath := "maDMP.ipynb"
		redirectURL := fmt.Sprintf("%s/%s/%s?filepath=%s", binderUrl, repoName, branch, filepath)
		c.RawRedirect(redirectURL)
	}
}
