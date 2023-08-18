// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package admin

import (
	"fmt"
	"strings"
	"net/url"

	"github.com/unknwon/com"
	log "unknwon.dev/clog/v2"

	"github.com/NII-DG/gogs/internal/conf"
	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/db"
	"github.com/NII-DG/gogs/internal/email"
	"github.com/NII-DG/gogs/internal/form"
	"github.com/NII-DG/gogs/internal/route"
	"github.com/NII-DG/gogs/internal/utils/regex"
)

const (
	USERS     = "admin/user/list"
	USER_NEW  = "admin/user/new"
	USER_EDIT = "admin/user/edit"
)

func Users(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	route.RenderUserSearch(c, &route.UserSearchOptions{
		Type:     db.UserIndividual,
		Counter:  db.CountUsers,
		Ranger:   db.ListUsers,
		PageSize: conf.UI.Admin.UserPagingNum,
		OrderBy:  "id ASC",
		TplName:  USERS,
	})
}

func NewUser(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users.new_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	c.Data["login_type"] = "0-0"

	sources, err := db.LoginSources.List(db.ListLoginSourceOpts{})
	if err != nil {
		c.Error(err, "list login sources")
		return
	}
	c.Data["Sources"] = sources

	c.Data["CanSendEmail"] = conf.Email.Enabled
	c.Success(USER_NEW)
}

func NewUserPost(c *context.Context, f form.AdminCrateUser) {
	c.Data["Title"] = c.Tr("admin.users.new_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true

	sources, err := db.LoginSources.List(db.ListLoginSourceOpts{})
	if err != nil {
		c.Error(err, "list login sources")
		return
	}
	c.Data["Sources"] = sources

	c.Data["CanSendEmail"] = conf.Email.Enabled

	if c.HasError() {
		c.Success(USER_NEW)
		return
	}

	u := &db.User{
		Name:     f.UserName,
		Email:    f.Email,
		Passwd:   f.Password,
		IsActive: true,
	}

	if len(f.LoginType) > 0 {
		fields := strings.Split(f.LoginType, "-")
		if len(fields) == 2 {
			u.LoginSource = com.StrTo(fields[1]).MustInt64()
			u.LoginName = f.LoginName
		}
	}

	if err := db.CreateUser(u); err != nil {
		switch {
		case db.IsErrUserAlreadyExist(err):
			c.Data["Err_UserName"] = true
			c.RenderWithErr(c.Tr("form.username_been_taken"), USER_NEW, &f)
		case db.IsErrEmailAlreadyUsed(err):
			c.Data["Err_Email"] = true
			c.RenderWithErr(c.Tr("form.email_been_used"), USER_NEW, &f)
		case db.IsErrNameNotAllowed(err):
			c.Data["Err_UserName"] = true
			c.RenderWithErr(c.Tr("user.form.name_not_allowed", err.(db.ErrNameNotAllowed).Value()), USER_NEW, &f)
		default:
			c.Error(err, "create user")
		}
		return
	}
	log.Trace("Account created by admin (%s): %s", c.User.Name, u.Name)

	// Send email notification.
	if f.SendNotify && conf.Email.Enabled {
		email.SendRegisterNotifyMail(c.Context, db.NewMailerUser(u))
	}

	c.Flash.Success(c.Tr("admin.users.new_success", u.Name))
	c.Redirect(conf.Server.Subpath + "/admin/users/" + com.ToStr(u.ID))
}

func prepareUserInfo(c *context.Context) *db.User {
	u, err := db.GetUserByID(c.ParamsInt64(":userid"))
	if err != nil {
		c.Error(err, "get user by ID")
		return nil
	}
	c.Data["User"] = u

	if u.LoginSource > 0 {
		c.Data["LoginSource"], err = db.LoginSources.GetByID(u.LoginSource)
		if err != nil {
			c.Error(err, "get login source by ID")
			return nil
		}
	} else {
		c.Data["LoginSource"] = &db.LoginSource{}
	}

	sources, err := db.LoginSources.List(db.ListLoginSourceOpts{})
	if err != nil {
		c.Error(err, "list login sources")
		return nil
	}
	c.Data["Sources"] = sources

	return u
}

func EditUser(c *context.Context) {
	c.Data["Title"] = c.Tr("admin.users.edit_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true
	c.Data["EnableLocalPathMigration"] = conf.Repository.EnableLocalPathMigration
	list, err := db.GetAffiliationList()
	if err != nil {
		log.Error("Failed to get affiliation: %v", err)
	}
	c.Data["AffiliationList"] = list

	prepareUserInfo(c)
	if c.Written() {
		return
	}

	c.Success(USER_EDIT)
}

func EditUserPost(c *context.Context, f form.AdminEditUser) {
	c.Data["Title"] = c.Tr("admin.users.edit_account")
	c.Data["PageIsAdmin"] = true
	c.Data["PageIsAdminUsers"] = true
	c.Data["EnableLocalPathMigration"] = conf.Repository.EnableLocalPathMigration
	list, err := db.GetAffiliationList()
	if err != nil {
		log.Error("Failed to get affiliation: %v", err)
	}
	c.Data["AffiliationList"] = list

	u := prepareUserInfo(c)
	if c.Written() {
		return
	}

	if c.HasError() {
		c.Success(USER_EDIT)
		return
	}

	fields := strings.Split(f.LoginType, "-")
	if len(fields) == 2 {
		loginSource := com.StrTo(fields[1]).MustInt64()

		if u.LoginSource != loginSource {
			u.LoginSource = loginSource
		}
	}

	if len(f.Password) > 0 {
		u.Passwd = f.Password
		var err error
		if u.Salt, err = db.GetUserSalt(); err != nil {
			c.Error(err, "get user salt")
			return
		}
		u.EncodePassword()
	}

	// check telephone format
	if len(f.Telephone) > 0 && !regex.CheckTelephoneFormat(f.Telephone) {
		c.FormErr("Telephone")
		c.RenderWithErr(c.Tr("form.enterred_invalid_telephone"), USER_EDIT, &f)
		return
	}
	// check ORCID URL
	// if PersonalURL is set
	if len( f.PersonalURL ) > 0 {
		orcid_domain := "orcid.org"
		// No Error check  because Checked for URL format at bind time
		parsedURL, _ := url.Parse(f.PersonalURL)
		urlDomain := parsedURL.Hostname()
		if strings.EqualFold( urlDomain, orcid_domain ) {
			value := parsedURL.Path
			fmt.Println("value = ", value)
			fmt.Println("value[1:] = ", value[1:])
			if !regex.CheckORCIDFormat(value[1:]) {
				c.FormErr("PersonalUrl")
				c.RenderWithErr(c.Tr("form.enterred_invalid_orcid_url"), USER_EDIT, &f)
				return
			}
		}
	}
	// check e-Rad Rearcher Number
	if len(f.ERadResearcherNumber) > 0 && !regex.CheckERadRearcherNumberFormat(f.ERadResearcherNumber) {
		c.FormErr("ERad")
		c.RenderWithErr(c.Tr("form.enterred_invalid_erad"), USER_EDIT, &f)
		return
	}
	// generate User.FullName
	fullName := ""
	if !regex.CheckAlphabet(f.FirstName) || !regex.CheckAlphabet(f.LastName) {
		// japanese user name
		fullName = fmt.Sprintf("%s %s", f.LastName, f.FirstName)
	} else {
		fullName = fmt.Sprintf("%s %s", f.FirstName, f.LastName)
	}

	u.LoginName = f.LoginName
	u.FirstName = f.FirstName
	u.LastName = f.LastName
	u.FullName = fullName
	u.AliasName = f.AliasName
	u.AffiliationId = f.AffiliationId
	u.Email = f.Email
	u.Telephone = f.Telephone
	u.ERadResearcherNumber = f.ERadResearcherNumber
	u.PersonalURL = f.PersonalURL
	u.Website = f.Website
	u.Location = f.Location
	u.MaxRepoCreation = f.MaxRepoCreation
	u.IsActive = f.Active
	u.IsAdmin = f.Admin
	u.AllowGitHook = f.AllowGitHook
	u.AllowImportLocal = f.AllowImportLocal
	u.ProhibitLogin = f.ProhibitLogin

	if err := db.UpdateUser(u); err != nil {
		if db.IsErrEmailAlreadyUsed(err) {
			c.Data["Err_Email"] = true
			c.RenderWithErr(c.Tr("form.email_been_used"), USER_EDIT, &f)
		} else {
			c.Error(err, "update user")
		}
		return
	}
	log.Trace("Account profile updated by admin (%s): %s", c.User.Name, u.Name)

	c.Flash.Success(c.Tr("admin.users.update_profile_success"))
	c.Redirect(conf.Server.Subpath + "/admin/users/" + c.Params(":userid"))
}

func DeleteUser(c *context.Context) {
	u, err := db.GetUserByID(c.ParamsInt64(":userid"))
	if err != nil {
		c.Error(err, "get user by ID")
		return
	}

	if err = db.DeleteUser(u); err != nil {
		switch {
		case db.IsErrUserOwnRepos(err):
			c.Flash.Error(c.Tr("admin.users.still_own_repo"))
			c.JSONSuccess(map[string]interface{}{
				"redirect": conf.Server.Subpath + "/admin/users/" + c.Params(":userid"),
			})
		case db.IsErrUserHasOrgs(err):
			c.Flash.Error(c.Tr("admin.users.still_has_org"))
			c.JSONSuccess(map[string]interface{}{
				"redirect": conf.Server.Subpath + "/admin/users/" + c.Params(":userid"),
			})
		default:
			c.Error(err, "delete user")
		}
		return
	}
	log.Trace("Account deleted by admin (%s): %s", c.User.Name, u.Name)

	c.Flash.Success(c.Tr("admin.users.deletion_success"))
	c.JSONSuccess(map[string]interface{}{
		"redirect": conf.Server.Subpath + "/admin/users",
	})
}
