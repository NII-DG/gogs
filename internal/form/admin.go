// Copyright 2014 The Gogs Authors. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package form

import (
	"github.com/go-macaron/binding"
	"gopkg.in/macaron.v1"
)

type AdminCrateUser struct {
	LoginType  string `binding:"Required"`
	LoginName  string
	UserName   string `binding:"Required;AlphaDashDot;MaxSize(35)"`
	Email      string `binding:"Required;Email;MaxSize(254)"`
	Password   string `binding:"MaxSize(255);AlphaDash"`
	SendNotify bool
}

func (f *AdminCrateUser) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}

type AdminEditUser struct {
	LoginType            string `binding:"Required"`
	LoginName            string
	FirstName            string `binding:"Required;MaxSize(100)"`
	LastName             string `binding:"Required;MaxSize(100)"`
	AliasName            string `binding:"MaxSize(255)"`
	AffiliationId        int64  `binding:"Required"`
	Email                string `binding:"Required;Email;MaxSize(254)"`
	Telephone            string
	Password             string `binding:"MaxSize(255);AlphaDash"`
	ERadResearcherNumber string
	PersonalURL          string `binding:"Url"`
	Website              string `binding:"MaxSize(50)"`
	Location             string `binding:"MaxSize(50)"`
	MaxRepoCreation      int
	Active               bool
	Admin                bool
	AllowGitHook         bool
	AllowImportLocal     bool
	ProhibitLogin        bool
}

func (f *AdminEditUser) Validate(ctx *macaron.Context, errs binding.Errors) binding.Errors {
	return validate(errs, ctx.Data, f, ctx.Locale)
}
