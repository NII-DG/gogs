// Copyright 2016 The Gogs Authors. All rights reserved.

// Use of this source code is governed by a MIT-style

// license that can be found in the LICENSE file.

package repo

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/NII-DG/gogs/internal/context"
	"github.com/NII-DG/gogs/internal/fileutil"
	mock_context "github.com/NII-DG/gogs/internal/mocks/context"
	mock_db "github.com/NII-DG/gogs/internal/mocks/db"
	mock_fileutil "github.com/NII-DG/gogs/internal/mocks/fileutil"
	mock_repo "github.com/NII-DG/gogs/internal/mocks/repo"
	"github.com/gogs/git-module"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func Test_dmpUtil_CreateDmp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	mockDmpUtil := mock_repo.NewMockAbstructDmpUtil(mockCtrl)
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)
	mockCtxRepo := mock_context.NewMockAbstructCtxRepository(mockCtrl)
	mockDbRepo := mock_db.NewMockAbstructDbRepository(mockCtrl)

	org := "jst"

	// c.CallData()された値を確認するための変数
	dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		PrepateMockDmpUtil  func() AbstructDmpUtil
		PrepareMockContexts func() context.AbstructContext
		wantErr             bool
	}{
		// TODO: Add test cases.
		{
			name: "正常系：エラーが発生しない",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(nil)
				mockDmpUtil.EXPECT().GetCombinedDmp(org).Return("Success", nil)
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				mockDbRepo.EXPECT().FullName().Return("dummyRepo")

				mockCtx.EXPECT().Success("repo/editor/edit")

				return mockCtx
			},
			wantErr: false,
		},
		{
			name: "異常系:BidingDmpSchemaList()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "異常系:FetchDmpSchema()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
		{
			name: "異常系:GetCombinedDmp()でエラーが発生する",
			PrepateMockDmpUtil: func() AbstructDmpUtil {
				mockDmpUtil.EXPECT().BidingDmpSchemaList(mockCtx).Return(nil)
				mockDmpUtil.EXPECT().FetchDmpSchema(mockCtx, org).Return(nil)
				mockDmpUtil.EXPECT().GetCombinedDmp(org).Return("", fmt.Errorf("Error"))
				return mockDmpUtil
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().QueryEscape("schema").Return(org)
				mockCtx.EXPECT().PageIs("Edit")
				mockCtx.EXPECT().RequireHighlightJS()
				mockCtx.EXPECT().RequireSimpleMDE()
				mockCtx.EXPECT().GetRepo().AnyTimes().Return(mockCtxRepo)
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)

				mockCtxRepo.EXPECT().GetTreePath().AnyTimes().Return("")
				mockCtxRepo.EXPECT().GetRepoLink().AnyTimes().Return("localhost:3080/dummyUser/dummyRepo")
				mockCtxRepo.EXPECT().GetBranchName().AnyTimes().Return("master")
				mockCtxRepo.EXPECT().GetDbRepo().AnyTimes().Return(mockDbRepo)
				mockCtxRepo.EXPECT().GetCommitId().AnyTimes().Return(&git.SHA1{})

				return mockCtx
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			createDmp(tt.PrepareMockContexts(), tt.PrepateMockDmpUtil())
			if !tt.wantErr {
				// 正常系の際はdummyDataの中身を確認
				if reflect.DeepEqual(fmt.Sprintf("%v", dummyData["FileContent"]), `{"name":"dummySchema"}`) {
					t.Errorf("DMP Schema is wrong,  got:  %v,  want: `[{\"name\":\"dummySchema\"}]`", dummyData["FileContent"])
				}
			}
		})
	}
}

func Test_dmpUtil_fetchDmpSchema(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileUtil := mock_fileutil.NewMockIFFileUtil(mockCtrl)
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)

	dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		d                   dmpUtil
		PrepareMockContexts func() context.AbstructContext
		path                string
		orgName             string
		wantErr             bool
	}{
		// TODO: Add test cases.
		{
			name: "正常系",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			path:    "sample",
			orgName: "meti",
			wantErr: false,
		},
		{
			name: "異常系：GetFileBypath()でエラーが返却される。",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			path:    "sample",
			orgName: "meti",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			schemaName := "schema_dmp_" + tt.orgName
			exPath := filepath.Join(tt.path, "dmp", "json_schema", schemaName)
			if !tt.wantErr {
				mockFileUtil.EXPECT().GetFileBypath(exPath).Return([]byte("OK"), nil)
				tt.d.fileUtil = mockFileUtil
			} else if tt.wantErr {
				mockFileUtil.EXPECT().GetFileBypath(exPath).Return([]byte(""), fmt.Errorf("NG"))
				tt.d.fileUtil = mockFileUtil
			}
			err := tt.d.fetchDmpSchema(tt.PrepareMockContexts(), tt.path, tt.orgName)

			if tt.wantErr {
				assert.Equal(t, fmt.Errorf("NG"), err)
			}
		})
	}
}

func Test_dmpUtil_bidingDmpSchemaList(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileUtil := mock_fileutil.NewMockIFFileUtil(mockCtrl)
	mockCtx := mock_context.NewMockAbstructContext(mockCtrl)

	dummyData := make(map[string]interface{})

	tests := []struct {
		name                string
		d                   dmpUtil
		PrepareMockContexts func() context.AbstructContext
		path                string
		wantErr             bool
	}{
		// TODO: Add test cases.
		{
			name: "正常系",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			path:    "sample",
			wantErr: false,
		},
		{
			name: "異常系",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			PrepareMockContexts: func() context.AbstructContext {
				mockCtx.EXPECT().CallData().AnyTimes().Return(dummyData)
				return mockCtx
			},
			path:    "sample",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exPath := filepath.Join(tt.path, "dmp", "orgs")
			if !tt.wantErr {
				files := []fs.FileInfo{}
				mockFileUtil.EXPECT().ReadDirBypath(exPath).Return(files, nil)
				tt.d.fileUtil = mockFileUtil
			} else if tt.wantErr {
				mockFileUtil.EXPECT().ReadDirBypath(exPath).Return(nil, fmt.Errorf("NG"))
				tt.d.fileUtil = mockFileUtil
			}
			err := tt.d.bidingDmpSchemaList(tt.PrepareMockContexts(), tt.path)

			if tt.wantErr {
				assert.Equal(t, fmt.Errorf("NG"), err)
			}
		})
	}
}

func Test_dmpUtil_getCombinedDmp(t *testing.T) {
	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()

	mockFileUtil := mock_fileutil.NewMockIFFileUtil(mockCtrl)

	type args struct {
		path   string
		schema string
	}
	tests := []struct {
		name        string
		d           dmpUtil
		args        args
		want        string
		wantErr     bool
		wantErrMsg  error
		mockPattern func(path, schema string) fileutil.IFFileUtil
	}{
		// TODO: Add test cases.
		{
			name: "正常系",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			args: args{
				path:   "samaple",
				schema: "jst",
			},
			want:       "rtn1rtn2",
			wantErr:    false,
			wantErrMsg: nil,
			mockPattern: func(path, schema string) fileutil.IFFileUtil {
				BasicSchemaPath := filepath.Join(path, "dmp", "basic")
				mockFileUtil.EXPECT().GetFileBypath(BasicSchemaPath).Return([]byte("rtn1"), nil)
				orgSchemaPath := filepath.Join(path, "dmp", "orgs", schema)
				mockFileUtil.EXPECT().GetFileBypath(orgSchemaPath).Return([]byte("rtn2"), nil)
				return mockFileUtil
			},
		},
		{
			name: "異常系:GetFileBypath(BasicSchemaPath)からエラーが返される",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			args: args{
				path:   "samaple",
				schema: "jst",
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: fmt.Errorf("Error1"),
			mockPattern: func(path, schema string) fileutil.IFFileUtil {
				BasicSchemaPath := filepath.Join(path, "dmp", "basic")
				mockFileUtil.EXPECT().GetFileBypath(BasicSchemaPath).Return([]byte(""), fmt.Errorf("Error1"))
				return mockFileUtil
			},
		},
		{
			name: "異常系:GetFileBypath(orgSchemaPath)からエラーが返される",
			d: dmpUtil{
				fileUtil: mockFileUtil,
			},
			args: args{
				path:   "samaple",
				schema: "jst",
			},
			want:       "",
			wantErr:    true,
			wantErrMsg: fmt.Errorf("Error2"),
			mockPattern: func(path, schema string) fileutil.IFFileUtil {
				BasicSchemaPath := filepath.Join(path, "dmp", "basic")
				mockFileUtil.EXPECT().GetFileBypath(BasicSchemaPath).Return([]byte("rtn1"), nil)
				orgSchemaPath := filepath.Join(path, "dmp", "orgs", schema)
				mockFileUtil.EXPECT().GetFileBypath(orgSchemaPath).Return([]byte(""), fmt.Errorf("Error2"))
				return mockFileUtil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.d.fileUtil = tt.mockPattern(tt.args.path, tt.args.schema)
			got, err := tt.d.getCombinedDmp(tt.args.path, tt.args.schema)

			//結果確認
			if !tt.wantErr {
				//正常系
				assert.Equal(t, tt.want, got)
				assert.Equal(t, tt.wantErrMsg, nil)
			} else if tt.wantErr {
				//異常系
				assert.Equal(t, tt.want, "")
				assert.Equal(t, tt.wantErrMsg, err)
			}
		})
	}
}
